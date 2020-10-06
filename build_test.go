package dotnetcoreconf_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	dotnetcoreconf "github.com/paketo-buildpacks/dotnet-core-conf"
	"github.com/paketo-buildpacks/dotnet-core-conf/fakes"
	"github.com/paketo-buildpacks/packit"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layersDir  string
		workingDir string
		cnbDir     string

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
		build = dotnetcoreconf.Build(ymlParser)
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
}
