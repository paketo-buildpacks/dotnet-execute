package dotnetexecute

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/scribe"
)

func Build(logger scribe.Logger) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		configParser := NewRuntimeConfigParser()
		var command string
		builtAppPath := context.WorkingDir

		_, publishOutputLocationSet := os.LookupEnv("PUBLISH_OUTPUT_LOCATION")

		if publishOutputLocationSet {
			builtAppPath, _ = os.LookupEnv("PUBLISH_OUTPUT_LOCATION")
		}

		config, err := configParser.Parse(filepath.Join(builtAppPath, "*.runtimeconfig.json"))

		if err != nil {
			return packit.BuildResult{}, fmt.Errorf("failed to find *.runtimeconfig.json: %w", err)
		}

		command = fmt.Sprintf("%s --urls http://0.0.0.0:${PORT:-8080}", filepath.Join(builtAppPath, config.AppName))
		if !config.Executable {
			// must check for the existence of <appName>.dll during rewrite
			command = fmt.Sprintf("dotnet %s.dll --urls http://0.0.0.0:${PORT:-8080}", filepath.Join(builtAppPath, config.AppName))
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
