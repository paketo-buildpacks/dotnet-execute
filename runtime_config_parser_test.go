package dotnetexecute_test

import (
	"io/ioutil"
	"os"
	"testing"

	dotnetexecute "github.com/paketo-buildpacks/dotnet-execute"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testRuntimeConfigParser(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		path   string
		parser dotnetexecute.RuntimeConfigParser
	)

	it.Before(func() {
		file, err := ioutil.TempFile("", "some-app.runtimeconfig.json")
		Expect(err).NotTo(HaveOccurred())
		defer file.Close()

		_, err = file.WriteString(`{}`)
		Expect(err).NotTo(HaveOccurred())

		path = file.Name()

		parser = dotnetexecute.NewRuntimeConfigParser()
	})

	it.After(func() {
		Expect(os.RemoveAll(path)).To(Succeed())
	})

	context("Parse", func() {
		it("parses a runtimeconfig.json file", func() {
			config, err := parser.Parse(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(config.RuntimeVersion).To(Equal(""))
		})

		context("when the runtime framework version is specified", func() {
			it.Before(func() {
				Expect(ioutil.WriteFile(path, []byte(`{
					"runtimeOptions": {
						"framework": {
							"version": "2.1.3"
						}
					}
				}`), 0600)).To(Succeed())
			})

			it("returns that version", func() {
				config, err := parser.Parse(path)
				Expect(err).NotTo(HaveOccurred())
				Expect(config.RuntimeVersion).To(Equal("2.1.3"))
			})
		})

		context("when the runtime framework is specified with no version", func() {
			it.Before(func() {
				Expect(ioutil.WriteFile(path, []byte(`{
					"runtimeOptions": {
						"framework": {
							"name": "Microsoft.NETCore.App"
						}
					}
				}`), 0600)).To(Succeed())
			})

			it("returns that version", func() {
				config, err := parser.Parse(path)
				Expect(err).NotTo(HaveOccurred())
				Expect(config.RuntimeVersion).To(Equal("*"))
			})
		})

		context("failure cases", func() {
			context("the runtimeconfig.json file cannot be opened", func() {
				it.Before(func() {
					Expect(os.RemoveAll(path)).To(Succeed())
				})

				it("returns an error", func() {
					_, err := parser.Parse(path)
					Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
				})
			})
		})

		context("the runtimeconfig.json file cannot be parsed", func() {
			it.Before(func() {
				Expect(ioutil.WriteFile(path, []byte(`%%%`), 0600)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := parser.Parse(path)
				Expect(err).To(MatchError(ContainSubstring("invalid character")))
			})
		})
	})
}
