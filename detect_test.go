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

	context("there is a *.runtimeconfig.json file present", func() {
		it.Before(func() {
			Expect(ioutil.WriteFile(filepath.Join(workingDir, "some-app.runtimeconfig.json"), []byte(""), os.ModePerm)).To(Succeed())
		})

		it("detects successfully", func() {
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

	context("there is a *.*sproj file present", func() {
		it.Before(func() {
			Expect(ioutil.WriteFile(filepath.Join(workingDir, "some-app.csproj"), []byte(""), os.ModePerm)).To(Succeed())
		})

		it("detects successfully", func() {
			_, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
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

			it("detects successfully", func() {
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
		context("project-path directory contains a *.runtimeconfig.json file", func() {
			it.Before(func() {
				err := ioutil.WriteFile(filepath.Join(workingDir, "src", "proj1", "some-app.runtimeconfig.json"), []byte(""), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
			})

			it("detects successfully", func() {
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
	context("failure cases", func() {
		context("there are multiple *.runtimeconfig.json files", func() {
			it.Before(func() {
				Expect(ioutil.WriteFile(filepath.Join(workingDir, "some-app.runtimeconfig.json"), []byte(""), os.ModePerm)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(workingDir, "another-app.runtimeconfig.json"), []byte(""), os.ModePerm)).To(Succeed())
			})

			it.After(func() {
				err := os.Remove(filepath.Join(workingDir, "some-app.runtimeconfig.json"))
				Expect(err).NotTo(HaveOccurred())
				err = os.Remove(filepath.Join(workingDir, "another-app.runtimeconfig.json"))
				Expect(err).NotTo(HaveOccurred())
			})

			it("detection fails", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(packit.Fail.WithMessage("multiple *.runtimeconfig.json files present")))
			})
		})

		context("there is no *.runtimeconfig.json or *.*sproj present", func() {
			it.Before(func() {
				files, err := filepath.Glob(filepath.Join(workingDir, "*.runtimeconfig.json"))
				Expect(err).NotTo(HaveOccurred())
				Expect(files).To(BeEmpty())

				files, err = filepath.Glob(filepath.Join(workingDir, "*.*sproj"))
				Expect(err).NotTo(HaveOccurred())
				Expect(files).To(BeEmpty())
			})

			it("detection fails", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).To(MatchError(packit.Fail))
			})
		})

	})
}
