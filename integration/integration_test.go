package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/dagger"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

var (
	bpDir, dotnetCoreConfURI string
)

func TestIntegration(t *testing.T) {
	var err error
	Expect := NewWithT(t).Expect
	bpDir, err = dagger.FindBPRoot()
	Expect(err).NotTo(HaveOccurred())
	dotnetCoreConfURI, err = dagger.PackageBuildpack(bpDir)
	Expect(err).ToNot(HaveOccurred())
	defer os.RemoveAll(dotnetCoreConfURI)

	spec.Run(t, "Integration", testIntegration, spec.Report(report.Terminal{}))
}

func testIntegration(t *testing.T, when spec.G, it spec.S) {
	var Expect func(interface{}, ...interface{}) GomegaAssertion
	it.Before(func() {
		Expect = NewWithT(t).Expect
	})

	when("the app is self contained aka vendored", func() {
		it("builds successfully", func() {
			appRoot := filepath.Join("testdata", "self_contained_2.1")

			app, err := dagger.PackBuild(appRoot, dotnetCoreConfURI)
			Expect(err).NotTo(HaveOccurred())
			defer app.Destroy()

			Expect(app.Start()).To(Succeed())
			body, err := app.HTTPGetBody("/")
			Expect(err).NotTo(HaveOccurred())
			Expect(body).To(ContainSubstring("Hello World!"))
		})
	})
}
