package dotnetexecute_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	dotnetexecute "github.com/paketo-buildpacks/dotnet-execute"
	"github.com/paketo-buildpacks/dotnet-execute/fakes"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/scribe"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		buffer     *bytes.Buffer
		workingDir string

		logger              scribe.Emitter
		projectParser       *fakes.ProjectParser
		runtimeConfigParser *fakes.ConfigParser

		detect packit.DetectFunc
	)

	it.Before(func() {
		var err error
		workingDir, err = os.MkdirTemp("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		runtimeConfigParser = &fakes.ConfigParser{}
		runtimeConfigParser.ParseCall.Returns.RuntimeConfig = dotnetexecute.RuntimeConfig{
			Path: filepath.Join(workingDir, "some-app.runtimeconfig.json"),
		}
		projectParser = &fakes.ProjectParser{}

		buffer = bytes.NewBuffer(nil)
		logger = scribe.NewEmitter(buffer)

		detect = dotnetexecute.Detect(dotnetexecute.Configuration{}, logger, runtimeConfigParser, projectParser)
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
						Metadata: dotnetexecute.BuildPlanMetadata{
							Launch: true,
						},
					},
				},
			}))

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

			it("requires dotnet-core-aspnet-runtime", func() {
				result, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Plan).To(Equal(packit.BuildPlan{
					Requires: []packit.BuildPlanRequirement{
						{
							Name: "dotnet-core-aspnet-runtime",
							Metadata: dotnetexecute.BuildPlanMetadata{
								Launch: true,
							},
						},
						{
							Name: "icu",
							Metadata: dotnetexecute.BuildPlanMetadata{
								Launch: true,
							},
						},
					},
				}))

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

			it("requires dotnet-core-aspnet-runtime", func() {
				result, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Plan).To(Equal(packit.BuildPlan{
					Requires: []packit.BuildPlanRequirement{
						{
							Name: "dotnet-core-aspnet-runtime",
							Metadata: dotnetexecute.BuildPlanMetadata{
								Launch: true,
							},
						},
						{
							Name: "icu",
							Metadata: dotnetexecute.BuildPlanMetadata{
								Launch: true,
							},
						},
					},
				}))

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

			it("requires dotnet-core-aspnet-runtime", func() {
				result, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Plan).To(Equal(packit.BuildPlan{
					Requires: []packit.BuildPlanRequirement{
						{
							Name: "dotnet-core-aspnet-runtime",
							Metadata: dotnetexecute.BuildPlanMetadata{
								Launch: true,
							},
						},
						{
							Name: "icu",
							Metadata: dotnetexecute.BuildPlanMetadata{
								Launch: true,
							},
						},
					},
				}))

				Expect(runtimeConfigParser.ParseCall.Receives.Glob).To(Equal(filepath.Join(workingDir, "*.runtimeconfig.json")))
				Expect(projectParser.FindProjectFileCall.Receives.Root).To(Equal(workingDir))
			})
		})
	})

	context("there is a proj file present (and no .runtimeconfig.json)", func() {
		it.Before(func() {
			projectParser.FindProjectFileCall.Returns.String = "/path/to/some-file.csproj"
		})

		it("detects successfully", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Requires: []packit.BuildPlanRequirement{
					{
						Name: "dotnet-application",
						Metadata: dotnetexecute.BuildPlanMetadata{
							Launch: true,
						},
					},
					{
						Name: "dotnet-core-aspnet-runtime",
						Metadata: dotnetexecute.BuildPlanMetadata{
							Launch: true,
						},
					},
					{
						Name: "icu",
						Metadata: dotnetexecute.BuildPlanMetadata{
							Launch: true,
						},
					},
				},
			}))

			Expect(runtimeConfigParser.ParseCall.Receives.Glob).To(Equal(filepath.Join(workingDir, "*.runtimeconfig.json")))

			Expect(projectParser.FindProjectFileCall.Receives.Root).To(Equal(workingDir))
			Expect(projectParser.NodeIsRequiredCall.Receives.Path).To(Equal("/path/to/some-file.csproj"))
		})
	})

	context("the proj file specifies a version of dotnet-runtime", func() {
		it.Before(func() {
			projectParser.FindProjectFileCall.Returns.String = "/path/to/some-file.csproj"
		})

		it("requires that version for dotnet-core-aspnet-runtime, requires a 70.* version of ICU, and detects successfully", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Requires: []packit.BuildPlanRequirement{
					{
						Name: "dotnet-application",
						Metadata: dotnetexecute.BuildPlanMetadata{
							Launch: true,
						},
					},
					{
						Name: "dotnet-core-aspnet-runtime",
						Metadata: dotnetexecute.BuildPlanMetadata{
							Launch: true,
						},
					},
					{
						Name: "icu",
						Metadata: dotnetexecute.BuildPlanMetadata{
							Launch: true,
						},
					},
				},
			}))

			Expect(runtimeConfigParser.ParseCall.Receives.Glob).To(Equal(filepath.Join(workingDir, "*.runtimeconfig.json")))

			Expect(projectParser.FindProjectFileCall.Receives.Root).To(Equal(workingDir))
			Expect(projectParser.NodeIsRequiredCall.Receives.Path).To(Equal("/path/to/some-file.csproj"))
		})
	})

	context("the proj file requires ASPNet", func() {
		it.Before(func() {
			projectParser.FindProjectFileCall.Returns.String = "/path/to/some-file.csproj"
		})

		it("requires that version for dotnet-core-aspnet-runtime correctly", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Requires: []packit.BuildPlanRequirement{
					{
						Name: "dotnet-application",
						Metadata: dotnetexecute.BuildPlanMetadata{
							Launch: true,
						},
					},
					{
						Name: "dotnet-core-aspnet-runtime",
						Metadata: dotnetexecute.BuildPlanMetadata{
							Launch: true,
						},
					},
					{
						Name: "icu",
						Metadata: dotnetexecute.BuildPlanMetadata{
							Launch: true,
						},
					},
				},
			}))

			Expect(runtimeConfigParser.ParseCall.Receives.Glob).To(Equal(filepath.Join(workingDir, "*.runtimeconfig.json")))

			Expect(projectParser.FindProjectFileCall.Receives.Root).To(Equal(workingDir))
			Expect(projectParser.NodeIsRequiredCall.Receives.Path).To(Equal("/path/to/some-file.csproj"))
		})
	})

	context("the proj file requires Node", func() {
		it.Before(func() {
			projectParser.FindProjectFileCall.Returns.String = "/path/to/some-file.csproj"
			projectParser.NodeIsRequiredCall.Returns.Bool = true
		})

		it("requires that version for dotnet-core-aspnet-runtime and node", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Requires: []packit.BuildPlanRequirement{
					{
						Name: "dotnet-application",
						Metadata: dotnetexecute.BuildPlanMetadata{
							Launch: true,
						},
					},
					{
						Name: "dotnet-core-aspnet-runtime",
						Metadata: dotnetexecute.BuildPlanMetadata{
							Launch: true,
						},
					},
					{
						Name: "node",
						Metadata: dotnetexecute.BuildPlanMetadata{
							VersionSource: "some-file.csproj",
							Launch:        true,
						},
					},
					{
						Name: "icu",
						Metadata: dotnetexecute.BuildPlanMetadata{
							Launch: true,
						},
					},
				},
			}))

			Expect(runtimeConfigParser.ParseCall.Receives.Glob).To(Equal(filepath.Join(workingDir, "*.runtimeconfig.json")))

			Expect(projectParser.FindProjectFileCall.Receives.Root).To(Equal(workingDir))
			Expect(projectParser.NodeIsRequiredCall.Receives.Path).To(Equal("/path/to/some-file.csproj"))
		})
	})

	context("when BP_DOTNET_PROJECT_PATH sets a custom project-path", func() {
		it.Before(func() {
			detect = dotnetexecute.Detect(dotnetexecute.Configuration{
				ProjectPath: "src/proj1",
			}, logger, runtimeConfigParser, projectParser)
		})

		context("project-path directory contains a proj file", func() {
			it.Before(func() {
				projectParser.FindProjectFileCall.Returns.String = "/path/to/some-file.csproj"
			})

			it("detects successfully", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(runtimeConfigParser.ParseCall.Receives.Glob).To(Equal(filepath.Join(workingDir, "src/proj1", "*.runtimeconfig.json")))

				Expect(projectParser.FindProjectFileCall.Receives.Root).To(Equal(filepath.Join(workingDir, "src/proj1")))
				Expect(projectParser.NodeIsRequiredCall.Receives.Path).To(Equal("/path/to/some-file.csproj"))
			})
		})
	})

	context("when BP_LIVE_RELOAD_ENABLED is set to true", func() {
		it.Before(func() {
			detect = dotnetexecute.Detect(dotnetexecute.Configuration{
				LiveReloadEnabled: true,
			}, logger, runtimeConfigParser, projectParser)
		})

		it("requires watchexec at launch", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan.Requires).To(ContainElement(packit.BuildPlanRequirement{
				Name: "watchexec",
				Metadata: dotnetexecute.BuildPlanMetadata{
					Launch: true,
				},
			},
			))
		})
	})

	context("when BP_DEBUG_ENABLED is set to true", func() {
		it.Before(func() {
			detect = dotnetexecute.Detect(dotnetexecute.Configuration{
				DebugEnabled: true,
			}, logger, runtimeConfigParser, projectParser)
		})

		it("requires vsdbg at launch", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan.Requires).To(ContainElement(packit.BuildPlanRequirement{
				Name: "vsdbg",
				Metadata: dotnetexecute.BuildPlanMetadata{
					Launch: true,
				},
			},
			))
		})
	})

	context("failure cases", func() {
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
	})
}
