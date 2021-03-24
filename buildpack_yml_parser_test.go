package dotnetexecute_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	dotnetexecute "github.com/paketo-buildpacks/dotnet-execute"
	"github.com/paketo-buildpacks/packit/scribe"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuildpackYMLParser(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		path   string
		logs   *bytes.Buffer
		parser dotnetexecute.BuildpackYMLParser
	)

	it.Before(func() {
		file, err := ioutil.TempFile("", "buildpack.yml")
		Expect(err).NotTo(HaveOccurred())
		defer file.Close()

		_, err = file.WriteString(`---
dotnet-build:
  project-path: "src/proj1"
`)
		Expect(err).NotTo(HaveOccurred())

		path = file.Name()

		logs = bytes.NewBuffer(nil)
		parser = dotnetexecute.NewBuildpackYMLParser(scribe.NewLogger(logs))
	})

	it.After(func() {
		Expect(os.RemoveAll(path)).To(Succeed())
	})

	context("Parse", func() {
		it("parses a buildpack.yml file", func() {
			configData, err := parser.Parse(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(configData.ProjectPath).To(Equal("src/proj1"))
		})
	})

	context("ParseProjectPath", func() {
		it("parses the project-path from a buildpack.yml file", func() {
			projectPath, err := parser.ParseProjectPath(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(projectPath).To(Equal("src/proj1"))
			Expect(logs.String()).To(ContainSubstring("WARNING: Setting the project path through buildpack.yml will be deprecated soon in Dotnet Execute Buildpack v1.0.0"))
			Expect(logs.String()).To(ContainSubstring("Please specify the project path through the $BP_DOTNET_PROJECT_PATH environment variable instead. See README.md or the documentation on paketo.io for more information."))
		})

		context("when the buildpack.yml file does not exist", func() {
			it.Before(func() {
				Expect(os.Remove(path)).To(Succeed())
			})

			it("returns an empty project-path", func() {
				projectPath, err := parser.ParseProjectPath(path)
				Expect(err).NotTo(HaveOccurred())
				Expect(projectPath).To(BeEmpty())
			})
		})
	})

	context("failure cases", func() {
		context("when the buildpack.yml file cannot be read", func() {
			it.Before(func() {
				Expect(os.Chmod(path, 0000)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(path, 0644)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := parser.ParseProjectPath(path)
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})

		context("when the contents of the buildpack.yml file are malformed", func() {
			it.Before(func() {
				err := ioutil.WriteFile(path, []byte("%%%"), 0644)
				Expect(err).NotTo(HaveOccurred())
			})

			it("returns an error", func() {
				_, err := parser.ParseProjectPath(path)
				Expect(err).To(MatchError(ContainSubstring("could not find expected directive name")))
			})
		})
	})
}
