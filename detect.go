package dotnetexecute

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit"
)

//go:generate faux --interface BuildpackConfigParser --output fakes/buildpack_config_parser.go
type BuildpackConfigParser interface {
	ParseProjectPath(path string) (projectPath string, err error)
}

//go:generate faux --interface ConfigParser --output fakes/config_parser.go
type ConfigParser interface {
	Parse(glob string) (RuntimeConfig, error)
}

func Detect(buildpackYMLParser BuildpackConfigParser, configParser ConfigParser) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		root := context.WorkingDir

		path, err := buildpackYMLParser.ParseProjectPath(filepath.Join(context.WorkingDir, "buildpack.yml"))
		if err != nil {
			return packit.DetectResult{}, fmt.Errorf("failed to parse buildpack.yml: %w", err)
		}

		if path != "" {
			root = filepath.Join(root, path)
		}

		requirements := []packit.BuildPlanRequirement{
			{
				Name: "icu",
				Metadata: map[string]interface{}{
					"launch": true,
				},
			},
		}

		config, err := configParser.Parse(filepath.Join(root, "*.runtimeconfig.json"))
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return packit.DetectResult{}, err
		}

		if config.RuntimeVersion != "" {
			requirements = append(requirements, packit.BuildPlanRequirement{
				Name: "dotnet-runtime",
				Metadata: map[string]interface{}{
					"version":        config.RuntimeVersion,
					"version-source": filepath.Base(config.Path),
					"launch":         true,
				},
			})

			requirements = append(requirements, packit.BuildPlanRequirement{
				Name: "dotnet-sdk",
				Metadata: map[string]interface{}{
					"version":        config.SDKVersion,
					"version-source": filepath.Base(config.Path),
					"launch":         !config.Executable,
				},
			})
		}

		projectFiles, err := filepath.Glob(filepath.Join(root, "*.*sproj"))
		if err != nil {
			return packit.DetectResult{}, fmt.Errorf("failed checking pattern *.*sproj: %w", err)
		}

		if config.Path == "" && len(projectFiles) == 0 {
			return packit.DetectResult{}, packit.Fail.WithMessage("no *.runtimeconfig.json or project file found")
		}

		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Requires: requirements,
			},
		}, nil
	}
}
