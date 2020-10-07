package dotnetcoreconf_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	dotnetcoreconf "github.com/paketo-buildpacks/dotnet-core-conf"
	"github.com/paketo-buildpacks/dotnet-core-conf/fakes"
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

		build     packit.BuildFunc
		ymlParser *fakes.Parser
	)

	it.Before(func() {
		var err error
		layersDir, err = ioutil.TempDir("", "layers")
		Expect(err).NotTo(HaveOccurred())

		cnbDir, err = ioutil.TempDir("", "cnb")
		Expect(err).NotTo(HaveOccurred())

		workingDir, err = ioutil.TempDir("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		ymlParser = &fakes.Parser{}
		buffer = bytes.NewBuffer(nil)
		logger := scribe.NewLogger(buffer)
		build = dotnetcoreconf.Build(ymlParser, logger)
	})

	it.After(func() {
		Expect(os.RemoveAll(layersDir)).To(Succeed())
		Expect(os.RemoveAll(cnbDir)).To(Succeed())
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	context("the app is a framework-dependent or self-contained executable", func() {
		it.Before(func() {
			err := os.MkdirAll(filepath.Join(workingDir, "my", "proj1"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())
			Expect(ioutil.WriteFile(filepath.Join(workingDir, "my", "proj1", "some-app.runtimeconfig.json"), nil, os.ModePerm)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(workingDir, "my", "proj1", "some-app"), nil, os.ModePerm)).To(Succeed())
			ymlParser.ParseProjectPathCall.Returns.ProjectPath = "my/proj1"
		})

		it.After(func() {
			Expect(os.RemoveAll(filepath.Join(workingDir, "my"))).To(Succeed())
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
				Processes: []packit.Process{
					{
						Type:    "web",
						Command: fmt.Sprintf("%s/my/proj1/some-app --urls http://0.0.0.0:${PORT:-8080}", workingDir),
					},
				},
			}))
		})
	})

	context("the app is a framework dependent deployment", func() {
		it.Before(func() {
			err := os.MkdirAll(filepath.Join(workingDir, "my", "proj1"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())
			Expect(ioutil.WriteFile(filepath.Join(workingDir, "my", "proj1", "some-app.runtimeconfig.json"), nil, os.ModePerm)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(workingDir, "my", "proj1", "some-app.dll"), nil, os.ModePerm)).To(Succeed())
			ymlParser.ParseProjectPathCall.Returns.ProjectPath = "my/proj1"
		})

		it.After(func() {
			Expect(os.RemoveAll(filepath.Join(workingDir, "my"))).To(Succeed())
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
				Processes: []packit.Process{
					{
						Type:    "web",
						Command: fmt.Sprintf("dotnet %s/my/proj1/some-app.dll --urls http://0.0.0.0:${PORT:-8080}", workingDir),
					},
				},
			}))
		})
	})

	context("failure cases", func() {
		context("buildpack yml can't be parsed", func() {
			it.Before(func() {
				ymlParser.ParseProjectPathCall.Returns.Err = fmt.Errorf("some-error")
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
				Expect(err).To(MatchError("failed to parse buildpack.yml: some-error"))
			})
		})

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
				Expect(err).To(MatchError(ContainSubstring("failed to find *.*runtimeconfig.json")))
			})
		})
	})
}
