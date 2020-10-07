package dotnetcoreconf

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/scribe"
)

func Build(buildpackYMLParser Parser, logger scribe.Logger) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)
		projRoot := context.WorkingDir

		bpYMLProjPath, err := buildpackYMLParser.ParseProjectPath(filepath.Join(context.WorkingDir, "buildpack.yml"))
		if err != nil {
			return packit.BuildResult{}, fmt.Errorf("failed to parse buildpack.yml: %w", err)
		}

		projRoot = filepath.Join(projRoot, bpYMLProjPath)

		runtimeConfigPath, err := getRuntimeConfigPath(projRoot)
		if err != nil {
			return packit.BuildResult{}, fmt.Errorf("failed to find *.*runtimeconfig.json: %w", err)
		}

		appName := getAppName(runtimeConfigPath)

		has, err := hasExecutable(projRoot, appName)
		if err != nil {
			return packit.BuildResult{}, fmt.Errorf("failed to stat app executable: %w", err)
		}

		command := fmt.Sprintf("%s/%s --urls http://0.0.0.0:${PORT:-8080}", projRoot, appName)
		if !has {
			// must check for the existence of <appName>.dll during rewrite
			command = fmt.Sprintf("dotnet %s/%s.dll --urls http://0.0.0.0:${PORT:-8080}", projRoot, appName)
		}

		logger.Process("Assigning launch processes")
		logger.Subprocess("web: %s", command)
		logger.Break()

		return packit.BuildResult{
			Processes: []packit.Process{
				{
					Type:    "web",
					Command: command,
				},
			},
		}, nil
	}
}

func getRuntimeConfigPath(appRoot string) (string, error) {
	configFiles, err := filepath.Glob(filepath.Join(appRoot, "*.runtimeconfig.json"))
	if err != nil {
		return "", err
	}

	if len(configFiles) > 1 {
		return "", fmt.Errorf("multiple *.runtimeconfig.json files present")
	}

	if len(configFiles) < 1 {
		return "", fmt.Errorf("no *.runtimeconfig.json files present")
	}

	return configFiles[0], nil
}

func getAppName(runtimeConfigPath string) string {
	runtimeConfigFile := filepath.Base(runtimeConfigPath)
	executableFile := strings.ReplaceAll(runtimeConfigFile, ".runtimeconfig.json", "")
	return executableFile
}

func hasExecutable(projRoot, appName string) (bool, error) {
	info, err := os.Stat(filepath.Join(projRoot, appName))
	if err != nil && !os.IsNotExist(err) {
		return false, err
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	if info.Mode()&0111 != 0 {
		return true, nil
	}

	return false, nil
}
