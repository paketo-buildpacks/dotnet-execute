package integration

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/dagger"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

var (
	dotnetCoreConfURI, builder string
	bpList                     []string
)

const testBuildpack = "test-buildpack"

var suite = spec.New("Integration", spec.Report(report.Terminal{}), spec.Parallel())

func init() {
	suite("Integration", testIntegration)
}

func BeforeSuite() {
	root, err := dagger.FindBPRoot()
	Expect(err).ToNot(HaveOccurred())
	dotnetCoreConfURI, err = dagger.PackageBuildpack(root)
	Expect(err).NotTo(HaveOccurred())

	config, err := dagger.ParseConfig("config.json")
	Expect(err).NotTo(HaveOccurred())

	builder = config.Builder

	for _, bp := range config.BuildpackOrder[builder] {
		var bpURI string
		if bp == testBuildpack {
			bpList = append(bpList, dotnetCoreConfURI)
			continue
		}
		bpURI, err = dagger.GetLatestBuildpack(bp)
		Expect(err).NotTo(HaveOccurred())
		bpList = append(bpList, bpURI)
	}
}

func AfterSuite() {
	for _, bp := range bpList {
		Expect(dagger.DeleteBuildpack(bp)).To(Succeed())
	}
}

func TestIntegration(t *testing.T) {
	RegisterTestingT(t)
	BeforeSuite()
	suite.Run(t)
	AfterSuite()
}

func testIntegration(t *testing.T, when spec.G, it spec.S) {
	var (
		Expect     func(interface{}, ...interface{}) Assertion
		Eventually func(interface{}, ...interface{}) AsyncAssertion
		app        *dagger.App
		err        error
	)

	it.Before(func() {
		Expect = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually
	})

	it.After(func() {
		if app != nil {
			Expect(app.Destroy()).To(Succeed())
		}
	})

	when("the app is self contained", func() {
		it("builds successfully", func() {
			appRoot := filepath.Join("testdata", "self_contained_2.1")

			app, err = dagger.NewPack(
				appRoot,
				dagger.RandomImage(),
				dagger.SetBuildpacks(bpList...),
				dagger.SetBuilder(builder),
			).Build()
			Expect(err).NotTo(HaveOccurred())

			if builder == "bionic" {
				app.SetHealthCheck("stat /workspace", "2s", "15s")
			}

			Expect(app.Start()).To(Succeed())

			Eventually(func() string {
				body, _, _ := app.HTTPGet("/")
				return body
			}).Should(ContainSubstring("Hello World"))
		})
	})
}
