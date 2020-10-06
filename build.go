package dotnetcoreconf

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/paketo-buildpacks/packit"
)

func Build(buildpackYMLParser Parser) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
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

		binaryName := getAppBinaryName(runtimeConfigPath)

		command := fmt.Sprintf("%s/%s --urls http://0.0.0.0:${PORT:-8080}", projRoot, binaryName)

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

func getAppBinaryName(runtimeConfigPath string) string {
	runtimeConfigFile := filepath.Base(runtimeConfigPath)
	executableFile := strings.ReplaceAll(runtimeConfigFile, ".runtimeconfig.json", "")
	return executableFile
}
