package dotnetexecute

import (
	"errors"
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/fs"
	"github.com/paketo-buildpacks/packit/scribe"
	"os"
	"path/filepath"
	"strconv"
)

func Build(buildpackYMLParser BuildpackConfigParser, configParser ConfigParser, logger scribe.Emitter) packit.BuildFunc {
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

		helperLayer, err := buildHelperLayer(context)
		if err != nil {
			return packit.BuildResult{}, err
		}

		return packit.BuildResult{
			Layers: []packit.Layer{
				helperLayer,
			},
			Launch: packit.LaunchMetadata{
				Processes: processes,
			},
		}, nil
	}
}

func buildHelperLayer(context packit.BuildContext) (helperLayer packit.Layer, err error) {
	helperLayer, err = context.Layers.Get("dotnetexecute_helper")
	helperLayer.Launch = true
	if err != nil {
		return packit.Layer{}, err
	}

	in := filepath.Join(context.CNBPath, "bin", "helper")
	execdDir := filepath.Join(helperLayer.Path, "exec.d")
	err = os.MkdirAll(execdDir, os.ModePerm)
	if err != nil {
		return packit.Layer{}, err
	}
	out := filepath.Join(execdDir, "helper")

	err = fs.Copy(in, out)
	if err != nil {
		return packit.Layer{}, err
	}

	return
}
