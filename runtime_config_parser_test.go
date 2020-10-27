package dotnetexecute_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	dotnetexecute "github.com/paketo-buildpacks/dotnet-execute"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testRuntimeConfigParser(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir string
		parser     dotnetexecute.RuntimeConfigParser
	)

	it.Before(func() {
		var err error
		workingDir, err = ioutil.TempDir("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		err = ioutil.WriteFile(filepath.Join(workingDir, "some-app.runtimeconfig.json"), []byte(`{}`), 0600)
		Expect(err).NotTo(HaveOccurred())

		parser = dotnetexecute.NewRuntimeConfigParser()
	})

	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	context("Parse", func() {
		it("parses the path of the file", func() {
			config, err := parser.Parse(filepath.Join(workingDir, "*.runtimeconfig.json"))
			Expect(err).NotTo(HaveOccurred())
			Expect(config.Path).To(Equal(filepath.Join(workingDir, "some-app.runtimeconfig.json")))
		})

		it("parses the runtime version", func() {
			config, err := parser.Parse(filepath.Join(workingDir, "*.runtimeconfig.json"))
			Expect(err).NotTo(HaveOccurred())
			Expect(config.RuntimeVersion).To(Equal(""))
		})

		it("parses the name of the app", func() {
			config, err := parser.Parse(filepath.Join(workingDir, "*.runtimeconfig.json"))
			Expect(err).NotTo(HaveOccurred())
			Expect(config.AppName).To(Equal("some-app"))
		})

		context("when the app includes an executable", func() {
			it.Before(func() {
				Expect(ioutil.WriteFile(filepath.Join(workingDir, "some-app"), nil, 0755)).To(Succeed())
			})

			it("reports that the app includes an executable", func() {
				config, err := parser.Parse(filepath.Join(workingDir, "*.runtimeconfig.json"))
				Expect(err).NotTo(HaveOccurred())
				Expect(config.Executable).To(BeTrue())
			})
		})

		context("when the app does not include an executable", func() {
			it("reports that the app does not include an executable", func() {
				config, err := parser.Parse(filepath.Join(workingDir, "*.runtimeconfig.json"))
				Expect(err).NotTo(HaveOccurred())
				Expect(config.Executable).To(BeFalse())
			})
		})

		context("when the runtime framework version is specified", func() {
			it.Before(func() {
				Expect(ioutil.WriteFile(filepath.Join(workingDir, "some-app.runtimeconfig.json"), []byte(`{
					"runtimeOptions": {
						"framework": {
							"version": "2.1.3"
						}
					}
				}`), 0600)).To(Succeed())
			})

			it("returns that version", func() {
				config, err := parser.Parse(filepath.Join(workingDir, "*.runtimeconfig.json"))
				Expect(err).NotTo(HaveOccurred())
				Expect(config.RuntimeVersion).To(Equal("2.1.3"))
			})
		})

		context("when the runtime framework is specified with no version", func() {
			it.Before(func() {
				Expect(ioutil.WriteFile(filepath.Join(workingDir, "some-app.runtimeconfig.json"), []byte(`{
					"runtimeOptions": {
						"framework": {
							"name": "Microsoft.NETCore.App"
						}
					}
				}`), 0600)).To(Succeed())
			})

			it("returns that version", func() {
				config, err := parser.Parse(filepath.Join(workingDir, "*.runtimeconfig.json"))
				Expect(err).NotTo(HaveOccurred())
				Expect(config.RuntimeVersion).To(Equal("*"))
			})
		})

		context("failure cases", func() {
			context("when given an invalid glob", func() {
				it("returns an error", func() {
					_, err := parser.Parse("[-]")
					Expect(err).To(MatchError(`failed to find *.runtimeconfig.json: syntax error in pattern: "[-]"`))
				})
			})
			context("when fileinfo for the executable cannot be retrieved", func() {
				it("returns an error", func() {
					_, err := parser.Parse("[-]")
					Expect(err).To(MatchError(`failed to find *.runtimeconfig.json: syntax error in pattern: "[-]"`))
				})
			})

			context("the runtimeconfig.json file cannot be opened", func() {
				it.Before(func() {
					Expect(ioutil.WriteFile(filepath.Join(workingDir, "other-app.runtimeconfig.json"), []byte(`{}`), 0600)).To(Succeed())
				})

				it("returns an error", func() {
					_, err := parser.Parse(filepath.Join(workingDir, "*.runtimeconfig.json"))
					Expect(err).To(MatchError(ContainSubstring("multiple *.runtimeconfig.json files present")))
					Expect(err).To(MatchError(ContainSubstring("some-app.runtimeconfig.json")))
					Expect(err).To(MatchError(ContainSubstring("other-app.runtimeconfig.json")))
				})
			})

			context("when there are multiple runtimeconfig.json files", func() {
				it.Before(func() {
					Expect(os.RemoveAll(filepath.Join(workingDir, "some-app.runtimeconfig.json"))).To(Succeed())
				})

				it("returns an error", func() {
					_, err := parser.Parse(filepath.Join(workingDir, "*.runtimeconfig.json"))
					Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
				})
			})

			context("the runtimeconfig.json file cannot be parsed", func() {
				it.Before(func() {
					Expect(ioutil.WriteFile(filepath.Join(workingDir, "some-app.runtimeconfig.json"), []byte(`%%%`), 0600)).To(Succeed())
				})

				it("returns an error", func() {
					_, err := parser.Parse(filepath.Join(workingDir, "*.runtimeconfig.json"))
					Expect(err).To(MatchError(ContainSubstring("invalid character")))
				})
			})
		})
	})
}
