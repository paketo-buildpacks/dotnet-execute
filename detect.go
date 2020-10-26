package dotnetexecute

import (
	"fmt"
	"path/filepath"

	"github.com/paketo-buildpacks/packit"
)

//go:generate faux --interface BuildpackConfigParser --output fakes/buildpack_config_parser.go
type BuildpackConfigParser interface {
	ParseProjectPath(path string) (projectPath string, err error)
}

//go:generate faux --interface ConfigParser --output fakes/config_parser.go
type ConfigParser interface {
	Parse(path string) (RuntimeConfig, error)
}

func Detect(buildpackYMLParser BuildpackConfigParser, configParser ConfigParser) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		projRoot := context.WorkingDir

		// Checking if there is a buildpack.yml file that contains a project path to use as the working dir
		bpYMLProjPath, err := buildpackYMLParser.ParseProjectPath(filepath.Join(context.WorkingDir, "buildpack.yml"))
		if err != nil {
			return packit.DetectResult{}, fmt.Errorf("failed to parse buildpack.yml: %w", err)
		}

		if bpYMLProjPath != "" {
			projRoot = filepath.Join(projRoot, bpYMLProjPath)
		}

		requirements := []packit.BuildPlanRequirement{
			{
				Name: "icu",
				Metadata: map[string]interface{}{
					"launch": true,
				},
			},
		}

		runtimeConfigFiles, err := filepath.Glob(filepath.Join(projRoot, "*.runtimeconfig.json"))
		if err != nil {
			return packit.DetectResult{}, fmt.Errorf("failed checking pattern *.runtimeconfig.json: %w", err)
		}
		if len(runtimeConfigFiles) > 1 {
			return packit.DetectResult{}, packit.Fail.WithMessage("multiple *.runtimeconfig.json files present")
		}

		if len(runtimeConfigFiles) == 1 {
			file := runtimeConfigFiles[0]
			config, err := configParser.Parse(file)
			if err != nil {
				return packit.DetectResult{}, fmt.Errorf("failed to parse %s: %w", file, err)
			}

			if config.RuntimeVersion != "" {
				requirements = append(requirements, packit.BuildPlanRequirement{
					Name: "dotnet-runtime",
					Metadata: map[string]interface{}{
						"version":        config.RuntimeVersion,
						"version-source": filepath.Base(file),
						"launch":         true,
					},
				})
			}
		}

		projFiles, err := filepath.Glob(filepath.Join(projRoot, "*.*sproj"))
		if err != nil {
			return packit.DetectResult{}, fmt.Errorf("failed checking pattern *.*sproj: %w", err)
		}

		if len(runtimeConfigFiles)+len(projFiles) == 0 {
			return packit.DetectResult{}, packit.Fail.WithMessage("no *.runtimeconfig.json or project file found")
		}

		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Requires: requirements,
			},
		}, nil
	}
}
