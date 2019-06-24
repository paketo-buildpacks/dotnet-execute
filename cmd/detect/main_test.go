package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/dotnet-core-conf-cnb/conf"

	specLogger "github.com/buildpack/libbuildpack/logger"

	"github.com/buildpack/libbuildpack/detect"

	"github.com/cloudfoundry/libcfbuildpack/test"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitDetect(t *testing.T) {
	spec.Run(t, "Detect", testDetect, spec.Report(report.Terminal{}))
}

func testDetect(t *testing.T, when spec.G, it spec.S) {
	var factory *test.DetectFactory

	it.Before(func() {
		RegisterTestingT(t)
		factory = test.NewDetectFactory(t)
	})

	when("there is a file with suffix runtimeconfig.json", func() {
		it("passes detect and adds dotnet-core-conf to the buildplan", func() {
			runtimeConfigPath := filepath.Join(factory.Detect.Application.Root, "test.runtimeconfig.json")
			test.TouchFile(t, runtimeConfigPath)
			defer os.RemoveAll(runtimeConfigPath)

			code, err := runDetect(factory.Detect)
			Expect(err).NotTo(HaveOccurred())
			Expect(code).To(Equal(detect.PassStatusCode))
			Expect(factory.Output).To(HaveKey(conf.Layer))
			Expect(factory.Output[conf.Layer].Metadata).To(HaveKeyWithValue("build", true))
		})
	})

	when("there are multiple files with suffix runtimeconfig.json", func() {
		it("fails detect with error about multiple runtime configs", func() {
			runtimeConfigPath1 := filepath.Join(factory.Detect.Application.Root, "test1.runtimeconfig.json")
			test.TouchFile(t, runtimeConfigPath1)
			defer os.RemoveAll(runtimeConfigPath1)

			runtimeConfigPath2 := filepath.Join(factory.Detect.Application.Root, "test2.runtimeconfig.json")
			test.TouchFile(t, runtimeConfigPath2)
			defer os.RemoveAll(runtimeConfigPath2)

			code, err := runDetect(factory.Detect)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(TooManyRuntimeConfigs))
			Expect(code).To(Equal(detect.FailStatusCode))
		})
	})

	when("there is NOT a file with suffix runtimeconfig.json", func() {
		it("fails detect", func() {
			buf := bytes.Buffer{}
			factory.Detect.Logger.Logger = specLogger.NewLogger(&buf, &buf)

			code, err := runDetect(factory.Detect)
			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(ContainSubstring(MissingRuntimeConfig))
			Expect(code).To(Equal(detect.FailStatusCode))
		})
	})

}
