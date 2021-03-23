package dotnetexecute

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/parsers"
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

func Detect(buildpackYMLParser BuildpackConfigParser, configParser ConfigParser, projectParser ProjectParser, projectPathParser parsers.ProjectPathParser) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		root := context.WorkingDir

		path, err := projectPathParser.Get(context.WorkingDir, "BP_DOTNET_PROJECT_PATH")
		if err != nil {
			return packit.DetectResult{}, err
		}

		if path == "" {
			path, err = buildpackYMLParser.ParseProjectPath(filepath.Join(context.WorkingDir, "buildpack.yml"))
			if err != nil {
				return packit.DetectResult{}, fmt.Errorf("failed to parse buildpack.yml: %w", err)
			}
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

		// FDE + FDD cases
		if config.Version != "" {
			requirements = append(requirements, packit.BuildPlanRequirement{
				Name: "dotnet-runtime",
				Metadata: map[string]interface{}{
					"version":        config.Version,
					"version-source": "runtimeconfig.json",
					"launch":         true,
				},
			})

			// Only make SDK available at launch if there is no executable (FDD case only)
			if !config.Executable {
				requirements = append(requirements, packit.BuildPlanRequirement{
					Name: "dotnet-sdk",
					Metadata: map[string]interface{}{
						"version":        getSDKVersion(config.Version),
						"version-source": "runtimeconfig.json",
					},
				})
			}

			if config.UsesASPNet {
				requirements = append(requirements, packit.BuildPlanRequirement{
					// When aspnet buildpack is rewritten per RFC0001, change to "dotnet-aspnet"
					Name: "dotnet-aspnetcore",
					Metadata: map[string]interface{}{
						"version":        config.Version,
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

		if projectFile != "" {
			version, err := projectParser.ParseVersion(projectFile)
			if err != nil {
				return packit.DetectResult{}, err
			}

			requirements = append(requirements, packit.BuildPlanRequirement{
				Name: "dotnet-application",
				Metadata: map[string]interface{}{
					"launch": true,
				},
			})

			requirements = append(requirements, packit.BuildPlanRequirement{
				Name: "dotnet-runtime",
				Metadata: map[string]interface{}{
					"version":        version,
					"version-source": filepath.Base(projectFile),
					"launch":         true,
				},
			})

			requirements = append(requirements, packit.BuildPlanRequirement{
				Name: "dotnet-sdk",
				Metadata: map[string]interface{}{
					"version":        getSDKVersion(version),
					"version-source": filepath.Base(projectFile),
				},
			})

			aspNetIsRequired, err := projectParser.ASPNetIsRequired(projectFile)
			if err != nil {
				return packit.DetectResult{}, err
			}

			if aspNetIsRequired {
				requirements = append(requirements, packit.BuildPlanRequirement{
					Name: "dotnet-aspnetcore",
					Metadata: map[string]interface{}{
						"version":        version,
						"version-source": filepath.Base(projectFile),
						"launch":         true,
					},
				})
			}

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
	pieces := strings.SplitN(version, ".", 3)
	if len(pieces) < 3 {
		pieces = append(pieces, "*")
	}

	var parts []string
	for i, part := range pieces {
		if i+1 == len(pieces) {
			part = "*"
		}

		parts = append(parts, part)

		if part == "*" {
			break
		}
	}

	return strings.Join(parts, ".")
}
