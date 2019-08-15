package conf_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/libcfbuildpack/layers"

	"github.com/cloudfoundry/dotnet-core-conf-cnb/conf"

	"github.com/cloudfoundry/libcfbuildpack/test"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/libcfbuildpack/buildpackplan"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitConf(t *testing.T) {
	spec.Run(t, "Conf", testConf, spec.Report(report.Terminal{}))
}

func testConf(t *testing.T, when spec.G, it spec.S) {
	var (
		f *test.BuildFactory
	)

	it.Before(func() {
		RegisterTestingT(t)
		f = test.NewBuildFactory(t)

		f.AddPlan(buildpackplan.Plan{Name: conf.Layer})

	})

	when("conf.NewContributor", func() {
		it("returns true if dotnet-core-conf is in the buildplan", func() {
			_, willContribute, err := conf.NewContributor(f.Build)
			Expect(err).NotTo(HaveOccurred())
			Expect(willContribute).To(BeTrue())
		})

		it("returns false if dotnet-core-conf is not in the buildplan", func() {
			f.Build.Plans = buildpackplan.Plans{}

			_, willContribute, err := conf.NewContributor(f.Build)
			Expect(err).NotTo(HaveOccurred())
			Expect(willContribute).To(BeFalse())
		})
	})

	when("Contribute", func() {
		it("sets the start command", func() {
			executable := "test-executable"
			executableFilePath := filepath.Join(f.Build.Application.Root, executable)
			test.TouchFile(t, executableFilePath)
			defer os.RemoveAll(executableFilePath)

			runtimeConfigFile := fmt.Sprintf("%s.runtimeconfig.json", executable)
			runtimeConfigFilePath := filepath.Join(f.Build.Application.Root, runtimeConfigFile)
			test.TouchFile(t, runtimeConfigFilePath)
			defer os.RemoveAll(runtimeConfigFilePath)

			startCmd := fmt.Sprintf("cd %s && ./%s --server.urls http://0.0.0.0:${PORT}", f.Build.Application.Root, executable)

			contributor, _, err := conf.NewContributor(f.Build)
			Expect(err).NotTo(HaveOccurred())
			Expect(contributor.Contribute()).To(Succeed())
			Expect(f.Build.Layers).To(test.HaveApplicationMetadata(layers.Metadata{Processes: []layers.Process{{"web", startCmd}}}))
		})
	})

}
