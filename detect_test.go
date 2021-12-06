package dotnetexecute_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	dotnetexecute "github.com/paketo-buildpacks/dotnet-execute"
	"github.com/paketo-buildpacks/dotnet-execute/fakes"
	"github.com/paketo-buildpacks/packit"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		buildpackYMLParser  *fakes.BuildpackConfigParser
		runtimeConfigParser *fakes.ConfigParser
		projectParser       *fakes.ProjectParser

		workingDir string

		detect packit.DetectFunc
	)

	it.Before(func() {
		var err error
		workingDir, err = ioutil.TempDir("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		buildpackYMLParser = &fakes.BuildpackConfigParser{}
		runtimeConfigParser = &fakes.ConfigParser{}
		runtimeConfigParser.ParseCall.Returns.RuntimeConfig = dotnetexecute.RuntimeConfig{
			Path: filepath.Join(workingDir, "some-app.runtimeconfig.json"),
		}
		projectParser = &fakes.ProjectParser{}

		detect = dotnetexecute.Detect(buildpackYMLParser, runtimeConfigParser, projectParser)
	})

	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	context("there is a *.runtimeconfig.json file present", func() {
		it.Before(func() {
			runtimeConfigParser.ParseCall.Returns.RuntimeConfig = dotnetexecute.RuntimeConfig{
				Path:       filepath.Join(workingDir, "some-app.runtimeconfig.json"),
				Executable: true,
			}
		})

		it("detects successfully", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Requires: []packit.BuildPlanRequirement{
					{
						Name: "icu",
						Metadata: map[string]interface{}{
							"launch": true,
						},
					},
				},
			}))

			Expect(buildpackYMLParser.ParseProjectPathCall.Receives.Path).To(Equal(filepath.Join(workingDir, "buildpack.yml")))
			Expect(runtimeConfigParser.ParseCall.Receives.Glob).To(Equal(filepath.Join(workingDir, "*.runtimeconfig.json")))
			Expect(projectParser.FindProjectFileCall.Receives.Root).To(Equal(workingDir))
		})

		context("when the runtimeconfig.json specifies a runtime framework", func() {
			it.Before(func() {
				runtimeConfigParser.ParseCall.Returns.RuntimeConfig = dotnetexecute.RuntimeConfig{
					Path:           filepath.Join(workingDir, "some-app.runtimeconfig.json"),
					RuntimeVersion: "2.1.0",
					Executable:     true,
				}
			})

			it("requires dotnet-runtime and dotnet-sdk", func() {
				result, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Plan).To(Equal(packit.BuildPlan{
					Requires: []packit.BuildPlanRequirement{
						{
							Name: "icu",
							Metadata: map[string]interface{}{
								"launch": true,
							},
						},
						{
							Name: "dotnet-runtime",
							Metadata: map[string]interface{}{
								"version":        "2.1.0",
								"version-source": "runtimeconfig.json",
								"launch":         true,
							},
						},
					},
				}))

				Expect(buildpackYMLParser.ParseProjectPathCall.Receives.Path).To(Equal(filepath.Join(workingDir, "buildpack.yml")))
				Expect(runtimeConfigParser.ParseCall.Receives.Glob).To(Equal(filepath.Join(workingDir, "*.runtimeconfig.json")))
				Expect(projectParser.FindProjectFileCall.Receives.Root).To(Equal(workingDir))
			})
		})

		context("when there is no executable", func() {
			it.Before(func() {
				runtimeConfigParser.ParseCall.Returns.RuntimeConfig = dotnetexecute.RuntimeConfig{
					Path:           filepath.Join(workingDir, "some-app.runtimeconfig.json"),
					RuntimeVersion: "2.1.0",
					Executable:     false,
				}
			})

			it("requires dotnet-sdk", func() {
				result, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Plan).To(Equal(packit.BuildPlan{
					Requires: []packit.BuildPlanRequirement{
						{
							Name: "icu",
							Metadata: map[string]interface{}{
								"launch": true,
							},
						},
						{
							Name: "dotnet-runtime",
							Metadata: map[string]interface{}{
								"version":        "2.1.0",
								"version-source": "runtimeconfig.json",
								"launch":         true,
							},
						},
						{
							Name: "dotnet-sdk",
							Metadata: map[string]interface{}{
								"version":        "2.1.*",
								"version-source": "runtimeconfig.json",
							},
						},
					},
				}))

				Expect(buildpackYMLParser.ParseProjectPathCall.Receives.Path).To(Equal(filepath.Join(workingDir, "buildpack.yml")))
				Expect(runtimeConfigParser.ParseCall.Receives.Glob).To(Equal(filepath.Join(workingDir, "*.runtimeconfig.json")))
				Expect(projectParser.FindProjectFileCall.Receives.Root).To(Equal(workingDir))
			})
		})

		context("when the runtimeconfig.json specifies an ASP.NET framework", func() {
			it.Before(func() {
				runtimeConfigParser.ParseCall.Returns.RuntimeConfig = dotnetexecute.RuntimeConfig{
					Path:           filepath.Join(workingDir, "some-app.runtimeconfig.json"),
					RuntimeVersion: "2.1.0",
					ASPNETVersion:  "2.1.0",
					Executable:     true,
				}
			})

			it("requires dotnet-runtime, dotnet-sdk (launch = false), and dotnet-aspnetcore", func() {
				result, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Plan).To(Equal(packit.BuildPlan{
					Requires: []packit.BuildPlanRequirement{
						{
							Name: "icu",
							Metadata: map[string]interface{}{
								"launch": true,
							},
						},
						{
							Name: "dotnet-runtime",
							Metadata: map[string]interface{}{
								"version":        "2.1.0",
								"version-source": "runtimeconfig.json",
								"launch":         true,
							},
						},
						{
							Name: "dotnet-aspnetcore",
							Metadata: map[string]interface{}{
								"version":        "2.1.0",
								"version-source": "runtimeconfig.json",
								"launch":         true,
							},
						},
					},
				}))

				Expect(buildpackYMLParser.ParseProjectPathCall.Receives.Path).To(Equal(filepath.Join(workingDir, "buildpack.yml")))
				Expect(runtimeConfigParser.ParseCall.Receives.Glob).To(Equal(filepath.Join(workingDir, "*.runtimeconfig.json")))
				Expect(projectParser.FindProjectFileCall.Receives.Root).To(Equal(workingDir))
			})
		})
	})

	context("there is a proj file present (and no .runtimeconfig.json)", func() {
		it.Before(func() {
			projectParser.FindProjectFileCall.Returns.String = "/path/to/some-file.csproj"
			projectParser.ParseVersionCall.Returns.String = "*"
		})

		it("detects successfully", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Requires: []packit.BuildPlanRequirement{
					{
						Name: "icu",
						Metadata: map[string]interface{}{
							"launch": true,
						},
					},
					{
						Name: "dotnet-application",
						Metadata: map[string]interface{}{
							"launch": true,
						},
					},
					{
						Name: "dotnet-runtime",
						Metadata: map[string]interface{}{
							"version":        "*",
							"version-source": "some-file.csproj",
							"launch":         true,
						},
					},
					{
						Name: "dotnet-sdk",
						Metadata: map[string]interface{}{
							"version":        "*",
							"version-source": "some-file.csproj",
						},
					},
				},
			}))

			Expect(buildpackYMLParser.ParseProjectPathCall.Receives.Path).To(Equal(filepath.Join(workingDir, "buildpack.yml")))
			Expect(runtimeConfigParser.ParseCall.Receives.Glob).To(Equal(filepath.Join(workingDir, "*.runtimeconfig.json")))

			Expect(projectParser.FindProjectFileCall.Receives.Root).To(Equal(workingDir))
			Expect(projectParser.ParseVersionCall.Receives.Path).To(Equal("/path/to/some-file.csproj"))
			Expect(projectParser.ASPNetIsRequiredCall.Receives.Path).To(Equal("/path/to/some-file.csproj"))
			Expect(projectParser.NodeIsRequiredCall.Receives.Path).To(Equal("/path/to/some-file.csproj"))
		})
	})

	context("the proj file specifies a version of dotnet-runtime", func() {
		it.Before(func() {
			projectParser.FindProjectFileCall.Returns.String = "/path/to/some-file.csproj"
			projectParser.ParseVersionCall.Returns.String = "3.1.*"
		})

		it("requires that version for dotnet-runtime and dotnet-sdk and detects successfully", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Requires: []packit.BuildPlanRequirement{
					{
						Name: "icu",
						Metadata: map[string]interface{}{
							"launch": true,
						},
					},
					{
						Name: "dotnet-application",
						Metadata: map[string]interface{}{
							"launch": true,
						},
					},
					{
						Name: "dotnet-runtime",
						Metadata: map[string]interface{}{
							"version":        "3.1.*",
							"version-source": "some-file.csproj",
							"launch":         true,
						},
					},
					{
						Name: "dotnet-sdk",
						Metadata: map[string]interface{}{
							"version":        "3.1.*",
							"version-source": "some-file.csproj",
						},
					},
				},
			}))

			Expect(buildpackYMLParser.ParseProjectPathCall.Receives.Path).To(Equal(filepath.Join(workingDir, "buildpack.yml")))
			Expect(runtimeConfigParser.ParseCall.Receives.Glob).To(Equal(filepath.Join(workingDir, "*.runtimeconfig.json")))

			Expect(projectParser.FindProjectFileCall.Receives.Root).To(Equal(workingDir))
			Expect(projectParser.ParseVersionCall.Receives.Path).To(Equal("/path/to/some-file.csproj"))
			Expect(projectParser.ASPNetIsRequiredCall.Receives.Path).To(Equal("/path/to/some-file.csproj"))
			Expect(projectParser.NodeIsRequiredCall.Receives.Path).To(Equal("/path/to/some-file.csproj"))
		})
	})

	context("the proj file requires ASPNet", func() {
		it.Before(func() {
			projectParser.FindProjectFileCall.Returns.String = "/path/to/some-file.csproj"
			projectParser.ParseVersionCall.Returns.String = "3.1.*"
			projectParser.ASPNetIsRequiredCall.Returns.Bool = true
		})

		it("requires that version for dotnet-runtime and dotnet-sdk and dotnet-aspnet correctly", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Requires: []packit.BuildPlanRequirement{
					{
						Name: "icu",
						Metadata: map[string]interface{}{
							"launch": true,
						},
					},
					{
						Name: "dotnet-application",
						Metadata: map[string]interface{}{
							"launch": true,
						},
					},
					{
						Name: "dotnet-runtime",
						Metadata: map[string]interface{}{
							"version":        "3.1.*",
							"version-source": "some-file.csproj",
							"launch":         true,
						},
					},
					{
						Name: "dotnet-sdk",
						Metadata: map[string]interface{}{
							"version":        "3.1.*",
							"version-source": "some-file.csproj",
						},
					},
					{
						Name: "dotnet-aspnetcore",
						Metadata: map[string]interface{}{
							"version":        "3.1.*",
							"version-source": "some-file.csproj",
							"launch":         true,
						},
					},
				},
			}))

			Expect(buildpackYMLParser.ParseProjectPathCall.Receives.Path).To(Equal(filepath.Join(workingDir, "buildpack.yml")))
			Expect(runtimeConfigParser.ParseCall.Receives.Glob).To(Equal(filepath.Join(workingDir, "*.runtimeconfig.json")))

			Expect(projectParser.FindProjectFileCall.Receives.Root).To(Equal(workingDir))
			Expect(projectParser.ParseVersionCall.Receives.Path).To(Equal("/path/to/some-file.csproj"))
			Expect(projectParser.ASPNetIsRequiredCall.Receives.Path).To(Equal("/path/to/some-file.csproj"))
			Expect(projectParser.NodeIsRequiredCall.Receives.Path).To(Equal("/path/to/some-file.csproj"))
		})
	})

	context("the proj file requires Node", func() {
		it.Before(func() {
			projectParser.FindProjectFileCall.Returns.String = "/path/to/some-file.csproj"
			projectParser.ParseVersionCall.Returns.String = "3.1.*"
			projectParser.NodeIsRequiredCall.Returns.Bool = true
		})

		it("requires that version for dotnet-runtime and dotnet-sdk and dotnet-aspnet correctly", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Requires: []packit.BuildPlanRequirement{
					{
						Name: "icu",
						Metadata: map[string]interface{}{
							"launch": true,
						},
					},
					{
						Name: "dotnet-application",
						Metadata: map[string]interface{}{
							"launch": true,
						},
					},
					{
						Name: "dotnet-runtime",
						Metadata: map[string]interface{}{
							"version":        "3.1.*",
							"version-source": "some-file.csproj",
							"launch":         true,
						},
					},
					{
						Name: "dotnet-sdk",
						Metadata: map[string]interface{}{
							"version":        "3.1.*",
							"version-source": "some-file.csproj",
						},
					},
					{
						Name: "node",
						Metadata: map[string]interface{}{
							"version-source": "some-file.csproj",
							"launch":         true,
						},
					},
				},
			}))

			Expect(buildpackYMLParser.ParseProjectPathCall.Receives.Path).To(Equal(filepath.Join(workingDir, "buildpack.yml")))
			Expect(runtimeConfigParser.ParseCall.Receives.Glob).To(Equal(filepath.Join(workingDir, "*.runtimeconfig.json")))

			Expect(projectParser.FindProjectFileCall.Receives.Root).To(Equal(workingDir))
			Expect(projectParser.ParseVersionCall.Receives.Path).To(Equal("/path/to/some-file.csproj"))
			Expect(projectParser.ASPNetIsRequiredCall.Receives.Path).To(Equal("/path/to/some-file.csproj"))
			Expect(projectParser.NodeIsRequiredCall.Receives.Path).To(Equal("/path/to/some-file.csproj"))
		})
	})

	context("there is a buildpack.yml which sets a custom project-path", func() {
		it.Before(func() {
			buildpackYMLParser.ParseProjectPathCall.Returns.ProjectPath = "src/proj1"
		})

		context("project-path directory contains a proj file", func() {
			it.Before(func() {
				projectParser.FindProjectFileCall.Returns.String = "/path/to/some-file.csproj"
				projectParser.ParseVersionCall.Returns.String = "*"
			})

			it("detects successfully", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
					BuildpackInfo: packit.BuildpackInfo{
						Version: "0.0.1",
					},
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(buildpackYMLParser.ParseProjectPathCall.Receives.Path).To(Equal(filepath.Join(workingDir, "buildpack.yml")))
				Expect(runtimeConfigParser.ParseCall.Receives.Glob).To(Equal(filepath.Join(workingDir, "src/proj1", "*.runtimeconfig.json")))

				Expect(projectParser.FindProjectFileCall.Receives.Root).To(Equal(filepath.Join(workingDir, "src/proj1")))
				Expect(projectParser.ParseVersionCall.Receives.Path).To(Equal("/path/to/some-file.csproj"))
				Expect(projectParser.ASPNetIsRequiredCall.Receives.Path).To(Equal("/path/to/some-file.csproj"))
				Expect(projectParser.NodeIsRequiredCall.Receives.Path).To(Equal("/path/to/some-file.csproj"))
			})
		})
	})

	context("when BP_DOTNET_PROJECT_PATH sets a custom project-path", func() {
		it.Before(func() {
			Expect(os.Setenv("BP_DOTNET_PROJECT_PATH", "src/proj1")).To(Succeed())
		})

		it.After(func() {
			Expect(os.Unsetenv("BP_DOTNET_PROJECT_PATH")).To(Succeed())
		})

		context("project-path directory contains a proj file", func() {
			it.Before(func() {
				projectParser.FindProjectFileCall.Returns.String = "/path/to/some-file.csproj"
				projectParser.ParseVersionCall.Returns.String = "*"
			})

			it("detects successfully", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(buildpackYMLParser.ParseProjectPathCall.CallCount).To(Equal(0))
				Expect(runtimeConfigParser.ParseCall.Receives.Glob).To(Equal(filepath.Join(workingDir, "src/proj1", "*.runtimeconfig.json")))

				Expect(projectParser.FindProjectFileCall.Receives.Root).To(Equal(filepath.Join(workingDir, "src/proj1")))
				Expect(projectParser.ParseVersionCall.Receives.Path).To(Equal("/path/to/some-file.csproj"))
				Expect(projectParser.ASPNetIsRequiredCall.Receives.Path).To(Equal("/path/to/some-file.csproj"))
				Expect(projectParser.NodeIsRequiredCall.Receives.Path).To(Equal("/path/to/some-file.csproj"))
			})
		})
	})

	context("when BP_LIVE_RELOAD_ENABLED is set to true", func() {
		it.Before(func() {
			Expect(os.Setenv("BP_LIVE_RELOAD_ENABLED", "true")).To(Succeed())
		})

		it.After(func() {
			Expect(os.Unsetenv("BP_LIVE_RELOAD_ENABLED")).To(Succeed())
		})

		it("requires watchexec at launch", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan.Requires).To(ContainElement(packit.BuildPlanRequirement{
				Name: "watchexec",
				Metadata: map[string]interface{}{
					"launch": true,
				},
			},
			))
		})
	})

	context("failure cases", func() {
		context("when the buildpack.yml parser fails", func() {
			it.Before(func() {
				buildpackYMLParser.ParseProjectPathCall.Returns.Err = errors.New("some-error")
			})

			it("returns an error", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: "/working-dir",
				})
				Expect(err).To(MatchError("failed to parse buildpack.yml: some-error"))
			})
		})

		context("when the runtime config parsing fails", func() {
			it.Before(func() {
				runtimeConfigParser.ParseCall.Returns.Error = errors.New("failed to parse runtime config")
			})

			it("fails", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed to parse runtime config"))
			})
		})

		context("there is no *.runtimeconfig.json or project file present", func() {
			it.Before(func() {
				runtimeConfigParser.ParseCall.Returns.RuntimeConfig = dotnetexecute.RuntimeConfig{}
			})

			it("detection fails", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).To(MatchError(packit.Fail.WithMessage("no *.runtimeconfig.json or project file found")))
			})
		})

		context("there is an error finding the project file", func() {
			it.Before(func() {
				projectParser.FindProjectFileCall.Returns.Error = errors.New("some-error")
			})

			it("fails", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("some-error"))
			})
		})

		context("parsing the version from the project file fails", func() {
			it.Before(func() {
				projectParser.FindProjectFileCall.Returns.String = "/path/to/some-file.csproj"
				projectParser.ParseVersionCall.Returns.Error = errors.New("some-error")
			})

			it("fails", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("some-error"))
			})
		})

		context("parsing the aspnet requirement from the project file fails", func() {
			it.Before(func() {
				projectParser.FindProjectFileCall.Returns.String = "/path/to/some-file.csproj"
				projectParser.ASPNetIsRequiredCall.Returns.Error = errors.New("some-error")
			})

			it("fails", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("some-error"))
			})
		})

		context("parsing the node requirement from the project file fails", func() {
			it.Before(func() {
				projectParser.NodeIsRequiredCall.Returns.Error = errors.New("some-error")
				projectParser.FindProjectFileCall.Returns.String = "/path/to/some-file.csproj"
			})

			it("fails", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("some-error"))
			})
		})

		context("parsing the value of BP_LIVE_RELOAD_ENABLED fails", func() {
			it.Before(func() {
				Expect(os.Setenv("BP_LIVE_RELOAD_ENABLED", "%%%")).To(Succeed())
			})

			it.After(func() {
				Expect(os.Unsetenv("BP_LIVE_RELOAD_ENABLED")).To(Succeed())
			})

			it("fails", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(ContainSubstring("invalid syntax")))
			})
		})
	})
}
