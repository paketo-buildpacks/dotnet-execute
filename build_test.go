package dotnetexecute_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	dotnetexecute "github.com/paketo-buildpacks/dotnet-execute"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/scribe"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layersDir  string
		workingDir string
		cnbDir     string
		buffer     *bytes.Buffer

		build packit.BuildFunc
	)

	it.Before(func() {
		var err error
		layersDir, err = ioutil.TempDir("", "layers")
		Expect(err).NotTo(HaveOccurred())

		cnbDir, err = ioutil.TempDir("", "cnb")
		Expect(err).NotTo(HaveOccurred())

		workingDir, err = ioutil.TempDir("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		buffer = bytes.NewBuffer(nil)
		logger := scribe.NewLogger(buffer)

		build = dotnetexecute.Build(logger)
	})

	it.After(func() {
		Expect(os.RemoveAll(layersDir)).To(Succeed())
		Expect(os.RemoveAll(cnbDir)).To(Succeed())
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	context("the app is a framework-dependent or self-contained executable", func() {
		it.Before(func() {
			Expect(ioutil.WriteFile(filepath.Join(workingDir, "some-app.runtimeconfig.json"), []byte(`{}`), os.ModePerm)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(workingDir, "some-app"), nil, os.ModePerm)).To(Succeed())
		})

		it("returns a result that builds correctly", func() {
			result, err := build(packit.BuildContext{
				WorkingDir: workingDir,
				CNBPath:    cnbDir,
				Stack:      "some-stack",
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "some-version",
				},
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{},
				},
				Layers: packit.Layers{Path: layersDir},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(Equal(packit.BuildResult{
				Plan: packit.BuildpackPlan{
					Entries: nil,
				},
				Layers: nil,
				Launch: packit.LaunchMetadata{
					Processes: []packit.Process{
						{
							Type:    "web",
							Command: fmt.Sprintf("%s --urls http://0.0.0.0:${PORT:-8080}", filepath.Join(workingDir, "some-app")),
						},
					},
				},
			}))
		})
	})

	context("the app is a framework dependent deployment", func() {
		it.Before(func() {
			Expect(ioutil.WriteFile(filepath.Join(workingDir, "some-app.runtimeconfig.json"), []byte(`{}`), os.ModePerm)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(workingDir, "some-app.dll"), nil, os.ModePerm)).To(Succeed())
		})

		it("returns a result that builds correctly", func() {
			result, err := build(packit.BuildContext{
				WorkingDir: workingDir,
				CNBPath:    cnbDir,
				Stack:      "some-stack",
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "some-version",
				},
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{},
				},
				Layers: packit.Layers{Path: layersDir},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(Equal(packit.BuildResult{
				Plan: packit.BuildpackPlan{
					Entries: nil,
				},
				Layers: nil,
				Launch: packit.LaunchMetadata{
					Processes: []packit.Process{
						{
							Type:    "web",
							Command: fmt.Sprintf("dotnet %s --urls http://0.0.0.0:${PORT:-8080}", filepath.Join(workingDir, "some-app.dll")),
						},
					},
				},
			}))
		})
	})

	context("the app is source code", func() {
		context("the built app in the layers dir is an FDE", func() {
			it.Before(func() {
				Expect(os.Mkdir(filepath.Join(layersDir, "publish-output-location"), os.ModePerm)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(layersDir, "publish-output-location", "some-app.runtimeconfig.json"), []byte(`{}`), os.ModePerm)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(layersDir, "publish-output-location", "some-app"), nil, os.ModePerm)).To(Succeed())
				os.Setenv("PUBLISH_OUTPUT_LOCATION", filepath.Join(layersDir, "publish-output-location"))
			})
			it.After(func() {
				os.Unsetenv("PUBLISH_OUTPUT_LOCATION")
			})
			it("returns a result that builds correctly", func() {
				result, err := build(packit.BuildContext{
					WorkingDir: workingDir,
					CNBPath:    cnbDir,
					Stack:      "some-stack",
					BuildpackInfo: packit.BuildpackInfo{
						Name:    "Some Buildpack",
						Version: "some-version",
					},
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{},
					},
					Layers: packit.Layers{Path: layersDir},
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(result).To(Equal(packit.BuildResult{
					Plan: packit.BuildpackPlan{
						Entries: nil,
					},
					Layers: nil,
					Launch: packit.LaunchMetadata{
						Processes: []packit.Process{
							{
								Type:    "web",
								Command: fmt.Sprintf("%s --urls http://0.0.0.0:${PORT:-8080}", filepath.Join(layersDir, "publish-output-location", "some-app")),
							},
						},
					},
				}))
			})
		})

		context("the built app in the layers dir is an FDD", func() {
			it.Before(func() {
				Expect(os.Mkdir(filepath.Join(layersDir, "publish-output-location"), os.ModePerm)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(layersDir, "publish-output-location", "some-app.runtimeconfig.json"), []byte(`{}`), os.ModePerm)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(layersDir, "publish-output-location", "some-app.dll"), nil, os.ModePerm)).To(Succeed())
				os.Setenv("PUBLISH_OUTPUT_LOCATION", filepath.Join(layersDir, "publish-output-location"))
			})
			it.After(func() {
				os.Unsetenv("PUBLISH_OUTPUT_LOCATION")
			})
			it("returns a result that builds correctly", func() {
				result, err := build(packit.BuildContext{
					WorkingDir: workingDir,
					CNBPath:    cnbDir,
					Stack:      "some-stack",
					BuildpackInfo: packit.BuildpackInfo{
						Name:    "Some Buildpack",
						Version: "some-version",
					},
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{},
					},
					Layers: packit.Layers{Path: layersDir},
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(result).To(Equal(packit.BuildResult{
					Plan: packit.BuildpackPlan{
						Entries: nil,
					},
					Layers: nil,
					Launch: packit.LaunchMetadata{
						Processes: []packit.Process{
							{
								Type:    "web",
								Command: fmt.Sprintf("dotnet %s --urls http://0.0.0.0:${PORT:-8080}", filepath.Join(layersDir, "publish-output-location", "some-app.dll")),
							},
						},
					},
				}))
			})
		})
	})

	context("failure cases", func() {
		context("runtime config not present", func() {
			it.Before(func() {
				files, err := filepath.Glob(filepath.Join(workingDir, "*.runtimeconfig.json"))
				Expect(err).NotTo(HaveOccurred())
				Expect(files).To(BeEmpty())
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					WorkingDir: workingDir,
					CNBPath:    cnbDir,
					Stack:      "some-stack",
					BuildpackInfo: packit.BuildpackInfo{
						Name:    "Some Buildpack",
						Version: "some-version",
					},
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{},
					},
					Layers: packit.Layers{Path: layersDir},
				})
				Expect(err).To(MatchError(ContainSubstring("failed to find *.runtimeconfig.json")))
			})
		})
	})
}
