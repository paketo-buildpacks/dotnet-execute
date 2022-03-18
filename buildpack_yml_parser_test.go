package dotnetexecute_test

import (
	"os"
	"testing"

	dotnetexecute "github.com/paketo-buildpacks/dotnet-execute"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuildpackYMLParser(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		path   string
		parser dotnetexecute.BuildpackYMLParser
	)

	it.Before(func() {
		file, err := os.CreateTemp("", "buildpack.yml")
		Expect(err).NotTo(HaveOccurred())
		defer file.Close()

		_, err = file.WriteString(`---
dotnet-build:
  project-path: "src/proj1"
`)
		Expect(err).NotTo(HaveOccurred())

		path = file.Name()

		parser = dotnetexecute.NewBuildpackYMLParser()
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
				err := os.WriteFile(path, []byte("%%%"), 0644)
				Expect(err).NotTo(HaveOccurred())
			})

			it("returns an error", func() {
				_, err := parser.ParseProjectPath(path)
				Expect(err).To(MatchError(ContainSubstring("could not find expected directive name")))
			})
		})
	})
}
