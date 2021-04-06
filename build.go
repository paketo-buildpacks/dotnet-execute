package dotnetexecute

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/scribe"
)

func Build(configParser ConfigParser, logger scribe.Logger) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

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
