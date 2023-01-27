package dotnetexecute

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Netflix/go-env"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

type BuildPlanMetadata struct {
	Version       string `toml:"version,omitempty"`
	VersionSource string `toml:"version-source,omitempty"`
	Launch        bool   `toml:"launch"`
}

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
	NodeIsRequired(path string) (bool, error)
}

// Detect will return a packit.DetectFunc that will be invoked during the
// detect phase of the buildpack lifecycle.
//
// Detection will contribute a Build Plan that requires different things
// depending on the type of app being built. See Configuration for details
// on how environment variable configuration influences detection.
//
// # Source Code Apps
//
// The buildpack will require .NET Core ASP.NET Runtime at launch-time. It will
// require ICU at launch time. It will require Nodejs at launch time if the app
// relies on JavaScript components.
//
// # Framework-dependent Deployments
//
// The buildpack will require the .NET Core ASP.NET Runtime at launch-time to
// run the framework-dependent app. It will require ICU at launch time. It will
// require Nodejs if the app relies on JavaScript components.
//
// # Framework-dependent Executables
//
// The buildpack will require the .NET Core ASP.NET Runtime at launch-time to
// run the framework-dependent app. It will require ICU at launch time. It will
// require Nodejs at launch time if the app relies on JavaScript components.
//
// Self-contained Executables
// The buildpack will require ICU at launch time. It will require Nodejs at
// launch time if the app relies on JavaScript components.
func Detect(
	config Configuration,
	logger scribe.Emitter,
	buildpackYMLParser BuildpackConfigParser,
	configParser ConfigParser,
	projectParser ProjectParser,
) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		logger.Debug.Process("Build configuration:")
		es, err := env.Marshal(&config)
		if err != nil {
			// not tested
			return packit.DetectResult{}, fmt.Errorf("parsing build configuration: %w", err)
		}
		for envVar := range es {
			// for bug https://github.com/Netflix/go-env/issues/23
			if !strings.Contains(envVar, "=") {
				logger.Debug.Subprocess("%s: %s", envVar, es[envVar])
			}
		}
		logger.Debug.Break()

		requirements := []packit.BuildPlanRequirement{}

		if config.LiveReloadEnabled {
			requirements = append(requirements, packit.BuildPlanRequirement{
				Name: "watchexec",
				Metadata: BuildPlanMetadata{
					Launch: true,
				},
			})
		}

		if config.DebugEnabled {
			requirements = append(requirements, packit.BuildPlanRequirement{
				Name: "vsdbg",
				Metadata: BuildPlanMetadata{
					Launch: true,
				},
			})
		}

		if config.ProjectPath == "" {
			var err error
			config.ProjectPath, err = buildpackYMLParser.ParseProjectPath(filepath.Join(context.WorkingDir, "buildpack.yml"))
			if err != nil {
				return packit.DetectResult{}, fmt.Errorf("failed to parse buildpack.yml: %w", err)
			}
		}

		root := context.WorkingDir
		if config.ProjectPath != "" {
			root = filepath.Join(root, config.ProjectPath)
		}

		logger.Debug.Process("Looking for .NET project files in '%s'", root)

		runtimeConfig, err := configParser.Parse(filepath.Join(root, "*.runtimeconfig.json"))
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return packit.DetectResult{}, err
		}
		// FDE + FDD cases
		if runtimeConfig.RuntimeVersion != "" {
			logger.Debug.Subprocess("Detected '%s'", filepath.Join(root, fmt.Sprintf("%s.runtimeconfig.json", runtimeConfig.AppName)))
			logger.Debug.Break()

			requirements = append(requirements, packit.BuildPlanRequirement{
				Name: "dotnet-core-aspnet-runtime",
				Metadata: BuildPlanMetadata{
					Launch: true,
				},
			})
		}

		projectFile, err := projectParser.FindProjectFile(root)
		if err != nil {
			return packit.DetectResult{}, err
		}

		if runtimeConfig.Path == "" && projectFile == "" {
			return packit.DetectResult{}, packit.Fail.WithMessage("no *.runtimeconfig.json or project file found")
		}

		if projectFile != "" {
			logger.Debug.Subprocess("Detected '%s'", projectFile)
			logger.Debug.Break()

			requirements = append(requirements, packit.BuildPlanRequirement{
				Name: "dotnet-application",
				Metadata: BuildPlanMetadata{
					Launch: true,
				},
			})

			requirements = append(requirements, packit.BuildPlanRequirement{
				Name: "dotnet-core-aspnet-runtime",
				Metadata: BuildPlanMetadata{
					Launch: true,
				},
			})

			nodeIsRequired, err := projectParser.NodeIsRequired(projectFile)
			if err != nil {
				return packit.DetectResult{}, err
			}

			if nodeIsRequired {
				requirements = append(requirements, packit.BuildPlanRequirement{
					Name: "node",
					Metadata: BuildPlanMetadata{
						VersionSource: filepath.Base(projectFile),
						Launch:        true,
					},
				})
			}
		}

		// ICU will always be append onto the build plan requirements
		requirements = append(requirements, packit.BuildPlanRequirement{
			Name: "icu",
			Metadata: BuildPlanMetadata{
				Launch: true,
			},
		})

		logger.Debug.Process("Returning build plan")
		logger.Debug.Subprocess("Requirements:")
		for _, req := range requirements {
			logger.Debug.Action(req.Name)
		}
		logger.Debug.Break()

		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Requires: requirements,
			},
		}, nil
	}
}
