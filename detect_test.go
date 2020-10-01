package dotnetcoreconf_test

import (
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

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		buildpackYMLParser *fakes.Parser
		workingDir         string
		detect             packit.DetectFunc
	)

	it.Before(func() {
		var err error
		workingDir, err = ioutil.TempDir("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		buildpackYMLParser = &fakes.Parser{}
		detect = dotnetcoreconf.Detect(buildpackYMLParser)
	})

	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	context("when *.runtimeconfig.json is present", func() {
		it.Before(func() {
			Expect(ioutil.WriteFile(filepath.Join(workingDir, "some-app.runtimeconfig.json"), []byte(""), os.ModePerm)).To(Succeed())
		})

		it("detects", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{
						Name: "dotnet-core-conf",
					},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name: "dotnet-core-conf",
						Metadata: map[string]interface{}{
							"build": true,
						},
					},
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

	context("when *.*sproj is present", func() {
		it.Before(func() {
			Expect(ioutil.WriteFile(filepath.Join(workingDir, "some-app.csproj"), []byte(""), os.ModePerm)).To(Succeed())
		})

		it("detects", func() {
			_, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})

	context("buildpack.yml sets a custom project-path", func() {
		it.Before(func() {
			buildpackYMLParser.ParseProjectPathCall.Returns.ProjectPath = "src/proj1"
			err := os.MkdirAll(filepath.Join(workingDir, "src", "proj1"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			err := os.RemoveAll(filepath.Join(workingDir, "src", "proj1"))
			Expect(err).NotTo(HaveOccurred())
		})

		context("project-path directory contains *.*sproj or *.runtimeconfig.json", func() {
			it.Before(func() {
				err := ioutil.WriteFile(filepath.Join(workingDir, "src", "proj1", "some-app.csproj"), []byte(""), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
			})

			it("detects", func() {
				result, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Plan).To(Equal(packit.BuildPlan{
					Provides: []packit.BuildPlanProvision{
						{
							Name: "dotnet-core-conf",
						},
					},
					Requires: []packit.BuildPlanRequirement{
						{
							Name: "dotnet-core-conf",
							Metadata: map[string]interface{}{
								"build": true,
							},
						},
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
	})

	context("when there is no *.runtimeconfig.json or *.*sproj present", func() {
		it.Before(func() {
			files, err := filepath.Glob(filepath.Join(workingDir, "*.runtimeconfig.json"))
			Expect(err).NotTo(HaveOccurred())
			Expect(files).To(BeEmpty())

			files, err = filepath.Glob(filepath.Join(workingDir, "*.*sproj"))
			Expect(err).NotTo(HaveOccurred())
			Expect(files).To(BeEmpty())
		})

		it("detection should fail", func() {
			_, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).To(MatchError(packit.Fail))
		})
	})
}
