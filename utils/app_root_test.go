package utils

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitAppRoot(t *testing.T) {
	spec.Run(t, "App Root", testAppRoot, spec.Report(report.Terminal{}))
}

func testAppRoot(t *testing.T, when spec.G, it spec.S) {
	it.Before(func() {
		RegisterTestingT(t)
	})

	when("GetAppRoot", func() {
		when("when there is project_path in buildpack.yml", func() {
			it("return the app root unchanged", func() {
				appRoot, err := ioutil.TempDir("", "")
				Expect(err).ToNot(HaveOccurred())
				defer os.RemoveAll(appRoot)

				newAppRoot, err := GetAppRoot(appRoot)
				Expect(err).ToNot(HaveOccurred())
				Expect(newAppRoot).To(Equal(appRoot))

			})
		})
		when("when there is project_path in buildpack.yml", func() {
			it("returns a modified app root", func() {
				appRoot, err := ioutil.TempDir("", "")
				Expect(err).ToNot(HaveOccurred())
				buildpackYamlPath := filepath.Join(appRoot, "buildpack.yml")

				Expect(ioutil.WriteFile(buildpackYamlPath, []byte(`---
dotnet-build:
  project-path: "src/proj1"
`), os.ModePerm)).To(Succeed())
				defer os.RemoveAll(appRoot)

				newAppRoot, err := GetAppRoot(appRoot)
				Expect(err).ToNot(HaveOccurred())
				Expect(newAppRoot).To(Equal(filepath.Join(appRoot, "src", "proj1")))
			})
		})
	})
}
