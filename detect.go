package dotnetexecute

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/Masterminds/semver"
	"github.com/paketo-buildpacks/packit/v2"
)

//go:generate faux --interface BuildpackConfigParser --output fakes/buildpack_config_parser.go
type BuildpackConfigParser interface {
	ParseProjectPath(path string) (projectPath string, err error)
}

//go:generate faux --interface ConfigParser --output fakes/config_parser.go
type ConfigParser interface {
	Parse(glob string) (RuntimeConfig, error)
}

//go:generate faux --interface ProjectParser --output fakes/project_parser.go
type ProjectParser interface {
	FindProjectFile(root string) (string, error)
	ParseVersion(path string) (string, error)
	ASPNetIsRequired(path string) (bool, error)
	NodeIsRequired(path string) (bool, error)
}

func Detect(buildpackYMLParser BuildpackConfigParser, configParser ConfigParser, projectParser ProjectParser) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		var projectPath string
		var ok bool
		var err error

		if projectPath, ok = os.LookupEnv("BP_DOTNET_PROJECT_PATH"); !ok {
			projectPath, err = buildpackYMLParser.ParseProjectPath(filepath.Join(context.WorkingDir, "buildpack.yml"))
			if err != nil {
				return packit.DetectResult{}, fmt.Errorf("failed to parse buildpack.yml: %w", err)
			}
		}

		root := context.WorkingDir

		if projectPath != "" {
			root = filepath.Join(root, projectPath)
		}

		requirements := []packit.BuildPlanRequirement{
			{
				Name: "icu",
				Metadata: map[string]interface{}{
					"launch": true,
				},
			},
			// Require dotnet aspnetcore at launch since app is probably an FDD or FDE when compiled;
			// but don't require a specific version
			{
				Name: "dotnet-aspnetcore",
				Metadata: map[string]interface{}{
					"launch": true,
				},
			},
		}

		if reload, ok := os.LookupEnv("BP_LIVE_RELOAD_ENABLED"); ok {
			shouldEnableReload, err := strconv.ParseBool(reload)
			if err != nil {
				return packit.DetectResult{}, fmt.Errorf("failed to parse BP_LIVE_RELOAD_ENABLED: %w", err)
			}
			if shouldEnableReload {
				requirements = append(requirements, packit.BuildPlanRequirement{
					Name: "watchexec",
					Metadata: map[string]interface{}{
						"launch": true,
					},
				})
			}
		}

		config, err := configParser.Parse(filepath.Join(root, "*.runtimeconfig.json"))
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return packit.DetectResult{}, err
		}

		// FDE + FDD cases
		if config.RuntimeVersion != "" {
			requirements = append(requirements, packit.BuildPlanRequirement{
				Name: "dotnet-aspnetcore",
				Metadata: map[string]interface{}{
					"version":        config.RuntimeVersion,
					"version-source": "runtimeconfig.json",
					"launch":         true,
				},
			})

			if config.ASPNETVersion != "" {
				requirements = append(requirements, packit.BuildPlanRequirement{
					// When aspnet buildpack is rewritten per RFC0001, change to "dotnet-aspnet"
					Name: "dotnet-aspnetcore",
					Metadata: map[string]interface{}{
						"version":        config.ASPNETVersion,
						"version-source": "runtimeconfig.json",
						"launch":         true,
					},
				})
			}
		}

		projectFile, err := projectParser.FindProjectFile(root)
		if err != nil {
			return packit.DetectResult{}, err
		}

		if config.Path == "" && projectFile == "" {
			return packit.DetectResult{}, packit.Fail.WithMessage("no *.runtimeconfig.json or project file found")
		}

		// TODO: Search for project files recursively?
		if projectFile != "" {
			requirements = append(requirements, packit.BuildPlanRequirement{
				Name: "dotnet-application",
				Metadata: map[string]interface{}{
					"launch": true,
				},
			})

			requirements = append(requirements, packit.BuildPlanRequirement{
				Name: "dotnet-aspnetcore",
				Metadata: map[string]interface{}{
					"launch":         true,
					"version-source": ".NET Execute Buildpack",
				},
			})

			nodeIsRequired, err := projectParser.NodeIsRequired(projectFile)
			if err != nil {
				return packit.DetectResult{}, err
			}

			if nodeIsRequired {
				requirements = append(requirements, packit.BuildPlanRequirement{
					Name: "node",
					Metadata: map[string]interface{}{
						"version-source": filepath.Base(projectFile),
						"launch":         true,
					},
				})
			}
		}

		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Requires: requirements,
			},
		}, nil
	}
}

func getSDKVersion(version string) string {
	if version == "" {
		return "*"
	}

	v := semver.MustParse(version)
	return fmt.Sprintf("%d.*", v.Major())
}
