package dotnetexecute

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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

func Build(
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

		config, err := configParser.Parse(filepath.Join(context.WorkingDir, "*.runtimeconfig.json"))
		if err != nil {
			return packit.BuildResult{}, fmt.Errorf("failed to find *.runtimeconfig.json: %w", err)
		}

		command := filepath.Join(context.WorkingDir, config.AppName)
		var args []string
		if !config.Executable {
			_, err := os.Stat(filepath.Join(context.WorkingDir, fmt.Sprintf("%s.dll", config.AppName)))
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return packit.BuildResult{}, err
			}
			if errors.Is(err, os.ErrNotExist) {
				return packit.BuildResult{}, fmt.Errorf("no entrypoint [%s.dll] found: %w ", config.AppName, err)
			}

			command = "dotnet"
			args = append(args, fmt.Sprintf("%s.dll", filepath.Join(context.WorkingDir, config.AppName)))
		}

		processes := []packit.Process{
			{
				Type:    "web",
				Command: command,
				Args:    args,
				Default: true,
				Direct:  true,
			},
		}

		if reload, ok := os.LookupEnv("BP_LIVE_RELOAD_ENABLED"); ok {
			shouldEnableReload, err := strconv.ParseBool(reload)
			if err != nil {
				return packit.BuildResult{}, fmt.Errorf("failed to parse BP_LIVE_RELOAD_ENABLED: %w", err)
			}

			if shouldEnableReload {
				processes = []packit.Process{
					{
						Type:    "web",
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
						Type:    "no-reload",
						Command: command,
						Args:    args,
						Direct:  true,
					},
				}
			}
		}

		logger.LaunchProcesses(processes)

		portChooserLayer, err := context.Layers.Get("port-chooser")
		if err != nil {
			return packit.BuildResult{}, err
		}
		portChooserLayer.Launch = true
		portChooserLayer.ExecD = []string{filepath.Join(context.CNBPath, "bin", "port-chooser")}

		sbomLayer, err := context.Layers.Get("sbom")
		if err != nil {
			return packit.BuildResult{}, err
		}
		sbomLayer, err = sbomLayer.Reset()
		if err != nil {
			return packit.BuildResult{}, err
		}
		sbomLayer.Launch = true

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
		sbomLayer.SBOM, err = sbomContent.InFormats(context.BuildpackInfo.SBOMFormats...)
		if err != nil {
			return packit.BuildResult{}, err
		}

		return packit.BuildResult{
			Layers: []packit.Layer{
				portChooserLayer,
				sbomLayer,
			},
			Launch: packit.LaunchMetadata{
				Processes: processes,
			},
		}, nil
	}
}
