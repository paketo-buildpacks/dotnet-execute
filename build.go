package dotnetexecute

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/Netflix/go-env"
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

	// When BP_DEBUG_ENABLED=TRUE, the buildpack will include the Visual Studio
	// Debugger in the app launch image Remote debuggers can invoke vsdbg inside
	// the running app container and attach to vsdbg's exposed port; also, the
	// buildpack will set ASPNETCORE_ENVIRONMENT=Development.
	DebugEnabled bool `env:"BP_DEBUG_ENABLED"`

	// When BP_LIVE_RELOAD_ENABLED=TRUE, the buildpack will make the app's entrypoint
	// process reload on changes to program files in the app container. It will
	// include watchexec in the app launch image and make the default container
	// entrypoint watchexec + <the usual app entrypoint>. See
	// https://github.com/watchexec/watchexec for more on watchexec as a
	// reloadable process manager.
	LiveReloadEnabled bool `env:"BP_LIVE_RELOAD_ENABLED"`

	// BP_LOG_LEVEL determines the amount of logs produced by the buildpack. Set
	// BP_LOG_LEVEL=DEBUG for more detailed logs.
	LogLevel string `env:"BP_LOG_LEVEL,default=INFO"`

	// When BP_DOTNET_PROJECT_PATH is set to a relative path, the buildpack
	// will look for project file(s) in that subdirectory to determine which
	// project to build into the app container.
	ProjectPath string `env:"BP_DOTNET_PROJECT_PATH"`
}

// Build will return a packit.BuildFunc that will be invoked during the build
// phase of the buildpack lifecycle.
//
// Build generates a SBOM of the .NET app's dependencies based on its compiled
// DLLs. It sets up the entrypoint for the app image and adds a helper that
// will determine at launch-time which container port the app should listen on.
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

		es, err := env.Marshal(&config)
		if err != nil {
			// not tested
			return packit.BuildResult{}, fmt.Errorf("parsing build configuration: %w", err)
		}

		logger.Debug.Process("Build configuration:")
		for envVar := range es {
			// for bug https://github.com/Netflix/go-env/issues/23
			if !strings.Contains(envVar, "=") {
				logger.Debug.Subprocess("%s: %s", envVar, es[envVar])
			}
		}
		logger.Debug.Break()

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

		// logger.LayerFlags(portChooserLayer)
		logger.Debug.Process("Setting up layer '%s'", portChooserLayer.Name)
		logger.Debug.Subprocess("Available at app launch: %t", portChooserLayer.Launch)
		logger.Debug.Subprocess("Available to other buildpacks: %t", portChooserLayer.Build)
		logger.Debug.Subprocess("Cached for rebuilds: %t", portChooserLayer.Cache)
		logger.Debug.Break()
		logger.EnvironmentVariables(portChooserLayer)

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
