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
		workingDir          string
		detect              packit.DetectFunc
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

		detect = dotnetexecute.Detect(buildpackYMLParser, runtimeConfigParser)
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
		})

		context("when the runtimeconfig.json specifies a runtime framework", func() {
			it.Before(func() {
				runtimeConfigParser.ParseCall.Returns.RuntimeConfig = dotnetexecute.RuntimeConfig{
					Path:           filepath.Join(workingDir, "some-app.runtimeconfig.json"),
					RuntimeVersion: "2.1.0",
					SDKVersion:     "2.1.*",
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
								"version-source": "some-app.runtimeconfig.json",
								"launch":         true,
							},
						},
						{
							Name: "dotnet-sdk",
							Metadata: map[string]interface{}{
								"version":        "2.1.*",
								"version-source": "some-app.runtimeconfig.json",
								"launch":         false,
							},
						},
					},
				}))
			})
		})

		context("when there is no executable", func() {
			it.Before(func() {
				runtimeConfigParser.ParseCall.Returns.RuntimeConfig = dotnetexecute.RuntimeConfig{
					Path:           filepath.Join(workingDir, "some-app.runtimeconfig.json"),
					RuntimeVersion: "2.1.0",
					SDKVersion:     "2.1.*",
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
								"version-source": "some-app.runtimeconfig.json",
								"launch":         true,
							},
						},
						{
							Name: "dotnet-sdk",
							Metadata: map[string]interface{}{
								"version":        "2.1.*",
								"version-source": "some-app.runtimeconfig.json",
								"launch":         true,
							},
						},
					},
				}))
			})
		})
	})

	context("there is a *.*sproj file present (and no .runtimeconfig.json)", func() {
		it.Before(func() {
			Expect(ioutil.WriteFile(filepath.Join(workingDir, "some-app.csproj"), []byte(""), os.ModePerm)).To(Succeed())
		})

		it.After(func() {
			Expect(os.RemoveAll(filepath.Join(workingDir, "some-app.csproj"))).To(Succeed())
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
		})
	})

	context("there is a buildpack.yml sets a custom project-path", func() {
		it.Before(func() {
			buildpackYMLParser.ParseProjectPathCall.Returns.ProjectPath = "src/proj1"
			err := os.MkdirAll(filepath.Join(workingDir, "src", "proj1"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			err := os.RemoveAll(filepath.Join(workingDir, "src", "proj1"))
			Expect(err).NotTo(HaveOccurred())
		})

		context("project-path directory contains a *.*sproj file", func() {
			it.Before(func() {
				err := ioutil.WriteFile(filepath.Join(workingDir, "src", "proj1", "some-app.csproj"), []byte(""), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
			})

			it.After(func() {
				err := os.Remove(filepath.Join(workingDir, "src", "proj1", "some-app.csproj"))
				Expect(err).NotTo(HaveOccurred())
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
			})
		})

		context("project-path directory contains a *.runtimeconfig.json file", func() {
			it.Before(func() {
				buildpackYMLParser.ParseProjectPathCall.Returns.ProjectPath = "./sub-dir"
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

				Expect(runtimeConfigParser.ParseCall.Receives.Glob).To(Equal(filepath.Join(workingDir, "sub-dir", "*.runtimeconfig.json")))
			})
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

		context("there is no *.runtimeconfig.json or *.*sproj present", func() {
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
	})
}
