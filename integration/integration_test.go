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
	bpDir, dotnetCoreConfURI, builder string
	bpList                            []string
)

const testBP = "test-buildpack"

var suite = spec.New("Integration", spec.Report(report.Terminal{}))

func init() {
	suite("Integration", testIntegration)
}

func TestIntegration(t *testing.T) {
	var err error
	Expect := NewWithT(t).Expect
	bpDir, err = dagger.FindBPRoot()
	Expect(err).NotTo(HaveOccurred())

	dotnetCoreConfURI, err = dagger.PackageBuildpack(bpDir)
	Expect(err).ToNot(HaveOccurred())
	defer dagger.DeleteBuildpack(dotnetCoreConfURI)

	suite.Run(t)
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
		config, err := dagger.ParseConfig("config.json")
		Expect(err).ToNot(HaveOccurred())
		builder = config.Builder

		for _, bp := range config.BuildpackOrder[builder] {
			if bp == testBP {
				bpList = append(bpList, dotnetCoreConfURI)
				continue
			}
			bpURI, err := dagger.GetLatestBuildpack(bp)
			Expect(err).ToNot(HaveOccurred())
			bpList = append(bpList, bpURI)
		}
	})

	it.After(func() {
		if app != nil {
			app.Destroy()
		}

		for _, bp := range bpList {
			Expect(dagger.DeleteBuildpack(bp))
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
