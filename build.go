package dotnetexecute

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/Masterminds/semver"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/scribe"
)

func Build(buildpackYMLParser BuildpackConfigParser, configParser ConfigParser, logger scribe.Logger) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		projectPath, err := buildpackYMLParser.ParseProjectPath(filepath.Join(context.WorkingDir, "buildpack.yml"))
		if err != nil {
			return packit.BuildResult{}, fmt.Errorf("error parsing buildpack.yml: %w", err)
		}

		if projectPath != "" {
			nextMajorVersion := semver.MustParse(context.BuildpackInfo.Version).IncMajor()
			logger.Subprocess("WARNING: Setting the project path through buildpack.yml will be deprecated soon in Dotnet Execute Buildpack v%s.", nextMajorVersion.String())
			logger.Subprocess("Please specify the project path through the $BP_DOTNET_PROJECT_PATH environment variable instead. See README.md or the documentation on paketo.io for more information.")
		}

		config, err := configParser.Parse(filepath.Join(context.WorkingDir, "*.runtimeconfig.json"))

		if err != nil {
			return packit.BuildResult{}, fmt.Errorf("failed to find *.runtimeconfig.json: %w", err)
		}

		command := fmt.Sprintf("%s --urls http://0.0.0.0:${PORT:-8080}", filepath.Join(context.WorkingDir, config.AppName))
		if !config.Executable {

			_, err := os.Stat(filepath.Join(context.WorkingDir, fmt.Sprintf("%s.dll", config.AppName)))
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return packit.BuildResult{}, err
			}
			if errors.Is(err, os.ErrNotExist) {
				return packit.BuildResult{}, fmt.Errorf("no entrypoint [%s.dll] found: %w ", config.AppName, err)
			}

			command = fmt.Sprintf("dotnet %s.dll --urls http://0.0.0.0:${PORT:-8080}", filepath.Join(context.WorkingDir, config.AppName))
		}

		if reload, ok := os.LookupEnv("BP_LIVE_RELOAD_ENABLED"); ok {
			shouldEnableReload, err := strconv.ParseBool(reload)
			if err != nil {
				return packit.BuildResult{}, fmt.Errorf("failed to parse BP_LIVE_RELOAD_ENABLED: %w", err)
			}
			if shouldEnableReload {

				command = fmt.Sprintf(`watchexec --restart --watch %s "%s"`, context.WorkingDir, command)
			}
		}

		logger.Process("Assigning launch processes")
		logger.Subprocess("web: %s", command)
		logger.Break()

		return packit.BuildResult{
			Launch: packit.LaunchMetadata{
				Processes: []packit.Process{
					{
						Type:    "web",
						Command: command,
					},
				},
			},
		}, nil
	}
}
