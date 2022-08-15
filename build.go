package dotnetexecute

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Masterminds/semver"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/sbom"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

//go:generate faux --interface SBOMGenerator --output fakes/sbom_generator.go
type SBOMGenerator interface {
	Generate(path string) (sbom.SBOM, error)
}

// Configuration enumerates the environment variable configuration options
// that govern the buildpack's behaviour.
type Configuration struct {
	// LogLevel             string `env:"BP_LOG_LEVEL"`

	// When BP_DEBUG_ENABLED=TRUE, the buildpack will include the Visual Studio
	// Debugger in the app launch image Remote debuggers can invoke vsdbg inside
	// the running app container and attach to vsdbg's exposed port.
	DebugEnabled bool `env:"BP_DEBUG_ENABLED"`

	// When BP_LIVE_RELOAD_ENABLED=TRUE, the buildpack will make the app's entrypoint
	// process reload on changes to program files in the app container. It will
	// include watchexec in the app launch image and make the default container
	// entrypoint watchexec + <the usual app entrypoint>. See
	// https://github.com/watchexec/watchexec for more on watchexec as a
	// reloadable process manager.
	LiveReloadEnabled bool `env:"BP_LIVE_RELOAD_ENABLED"`

	// When BP_DOTNET_PROJECT_PATH is set to a relative path, the buildpack
	// will look for project file(s) in that subdirectory to determine which
	// project to build into the app container.
	ProjectPath string `env:"BP_DOTNET_PROJECT_PATH"`
}

func Build(
	config Configuration,
	buildpackYMLParser BuildpackConfigParser,
	configParser ConfigParser,
	sbomGenerator SBOMGenerator,
	logger scribe.Emitter,
	clock chronos.Clock,
) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		projectPath, err := buildpackYMLParser.ParseProjectPath(filepath.Join(context.WorkingDir, "buildpack.yml"))
		if err != nil {
			return packit.BuildResult{}, fmt.Errorf("error parsing buildpack.yml: %w", err)
		}

		if projectPath != "" {
			nextMajorVersion := semver.MustParse(context.BuildpackInfo.Version).IncMajor()
			logger.Subprocess("WARNING: Setting the project path through buildpack.yml will be deprecated soon in .NET Execute Buildpack v%s.", nextMajorVersion.String())
			logger.Subprocess("Please specify the project path through the $BP_DOTNET_PROJECT_PATH environment variable instead. See README.md or the documentation on paketo.io for more information.")
		}

		runtimeConfig, err := configParser.Parse(filepath.Join(context.WorkingDir, "*.runtimeconfig.json"))
		if err != nil {
			return packit.BuildResult{}, fmt.Errorf("failed to find *.runtimeconfig.json: %w", err)
		}

		command := filepath.Join(context.WorkingDir, runtimeConfig.AppName)
		var args []string
		if !runtimeConfig.Executable {
			_, err := os.Stat(filepath.Join(context.WorkingDir, fmt.Sprintf("%s.dll", runtimeConfig.AppName)))
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return packit.BuildResult{}, err
			}
			if errors.Is(err, os.ErrNotExist) {
				return packit.BuildResult{}, fmt.Errorf("no entrypoint [%s.dll] found: %w ", runtimeConfig.AppName, err)
			}

			command = "dotnet"
			args = append(args, fmt.Sprintf("%s.dll", filepath.Join(context.WorkingDir, runtimeConfig.AppName)))
		}

		processes := []packit.Process{
			{
				Type:    runtimeConfig.AppName,
				Command: command,
				Args:    args,
				Default: true,
				Direct:  true,
			},
		}

		if config.LiveReloadEnabled {
			processes = []packit.Process{
				{
					Type:    fmt.Sprintf("reload-%s", runtimeConfig.AppName),
					Command: "watchexec",
					Args: append([]string{
						"--restart",
						"--watch", context.WorkingDir,
						"--shell", "none",
						"--",
						command,
					}, args...),
					Default: true,
					Direct:  true,
				},
				{
					Type:    runtimeConfig.AppName,
					Command: command,
					Args:    args,
					Direct:  true,
				},
			}
		}

		logger.LaunchProcesses(processes)

		portChooserLayer, err := context.Layers.Get("port-chooser")
		if err != nil {
			return packit.BuildResult{}, err
		}
		portChooserLayer.Launch = true
		portChooserLayer.ExecD = []string{filepath.Join(context.CNBPath, "bin", "port-chooser")}

		if config.DebugEnabled {
			portChooserLayer.LaunchEnv.Default("ASPNETCORE_ENVIRONMENT", "Development")
		}

		logger.EnvironmentVariables(portChooserLayer)

		logger.GeneratingSBOM(context.WorkingDir)
		var sbomContent sbom.SBOM
		duration, err := clock.Measure(func() error {
			sbomContent, err = sbomGenerator.Generate(context.WorkingDir)
			return err
		})
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Action("Completed in %s", duration.Round(time.Millisecond))
		logger.Break()

		logger.FormattingSBOM(context.BuildpackInfo.SBOMFormats...)
		sbomFormatter, err := sbomContent.InFormats(context.BuildpackInfo.SBOMFormats...)
		if err != nil {
			return packit.BuildResult{}, err
		}

		return packit.BuildResult{
			Layers: []packit.Layer{
				portChooserLayer,
			},
			Launch: packit.LaunchMetadata{
				Processes: processes,
				SBOM:      sbomFormatter,
			},
		}, nil
	}
}
