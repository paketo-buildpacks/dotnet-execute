package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/buildpack/libbuildpack/buildplan"

	"github.com/cloudfoundry/dotnet-core-conf-cnb/conf"

	specLogger "github.com/buildpack/libbuildpack/logger"

	"github.com/cloudfoundry/libcfbuildpack/detect"

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
			test.WriteFile(t, runtimeConfigPath, `
{
  "runtimeOptions": {
    "tfm": "netcoreapp2.2",
    "framework": {
      "name": "Microsoft.NETCore.App",
      "version": "2.2.5"
    }
  }
}`)
			defer os.RemoveAll(runtimeConfigPath)

			code, err := runDetect(factory.Detect)
			Expect(err).NotTo(HaveOccurred())
			Expect(code).To(Equal(detect.PassStatusCode))
			Expect(factory.Plans.Plan.Provides).To(Equal([]buildplan.Provided{{Name: conf.Layer}}))
			Expect(factory.Plans.Plan.Requires).To(Equal([]buildplan.Required{{
				Name: conf.Layer,
				Metadata: buildplan.Metadata{
					"build": true,
				}}}))
		})
	})

	when("there is a file with suffix runtimeconfig.json", func() {
		it("passes detect and adds dotnet-core-conf and icu to the buildplan when the stack is bionic", func() {
			runtimeConfigPath := filepath.Join(factory.Detect.Application.Root, "test.runtimeconfig.json")
			test.WriteFile(t, runtimeConfigPath, `
{
  "runtimeOptions": {
    "tfm": "netcoreapp2.2",
    "framework": {
      "name": "Microsoft.NETCore.App",
      "version": "2.2.5"
    }
  }
}`)
			defer os.RemoveAll(runtimeConfigPath)

			factory.Detect.Stack = "io.buildpacks.stacks.bionic"
			code, err := runDetect(factory.Detect)
			Expect(err).NotTo(HaveOccurred())
			Expect(code).To(Equal(detect.PassStatusCode))
			Expect(factory.Plans.Plan.Provides).To(Equal([]buildplan.Provided{{Name: conf.Layer}}))
			Expect(factory.Plans.Plan.Requires).To(Equal([]buildplan.Required{{
				Name: conf.Layer,
				Metadata: buildplan.Metadata{
					"build": true,
				}},
				{
					Name: "icu",
					Metadata: buildplan.Metadata{
						"launch": true,
					}},
			}))
		})
	})

	when("there is not runtimecongif.json", func() {
		when("when there is a *.csproj", func() {
			it("passes detect and add dotnet-core-conf to the buildplan", func() {
				csprojPath := filepath.Join(factory.Detect.Application.Root, "test.csproj")
				test.WriteFile(t, csprojPath, `test proj`)

				defer os.RemoveAll(csprojPath)

				code, err := runDetect(factory.Detect)
				Expect(err).NotTo(HaveOccurred())
				Expect(code).To(Equal(detect.PassStatusCode))
				Expect(factory.Plans.Plan.Provides).To(Equal([]buildplan.Provided{{Name: conf.Layer}}))
				Expect(factory.Plans.Plan.Requires).To(Equal([]buildplan.Required{{
					Name: conf.Layer,
					Metadata: buildplan.Metadata{
						"build": true,
					}}}))
			})
		})

		when("when there is a *.fsproj", func() {
			it("passes detect and add dotnet-core-conf to the buildplan", func() {
				csprojPath := filepath.Join(factory.Detect.Application.Root, "test.fsproj")
				test.WriteFile(t, csprojPath, `test proj`)

				defer os.RemoveAll(csprojPath)

				code, err := runDetect(factory.Detect)
				Expect(err).NotTo(HaveOccurred())
				Expect(code).To(Equal(detect.PassStatusCode))
				Expect(factory.Plans.Plan.Provides).To(Equal([]buildplan.Provided{{Name: conf.Layer}}))
				Expect(factory.Plans.Plan.Requires).To(Equal([]buildplan.Required{{
					Name: conf.Layer,
					Metadata: buildplan.Metadata{
						"build": true,
					}}}))
			})
		})
	})

	when("there are multiple files with suffix runtimeconfig.json", func() {
		it("fails detect with error about multiple runtime configs", func() {
			runtimeConfigPath1 := filepath.Join(factory.Detect.Application.Root, "test1.runtimeconfig.json")
			test.WriteFile(t, runtimeConfigPath1, "{}")
			defer os.RemoveAll(runtimeConfigPath1)

			runtimeConfigPath2 := filepath.Join(factory.Detect.Application.Root, "test2.runtimeconfig.json")
			test.WriteFile(t, runtimeConfigPath2, "{}")
			defer os.RemoveAll(runtimeConfigPath2)

			code, err := runDetect(factory.Detect)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("multiple *.runtimeconfig.json files present"))
			Expect(code).To(Equal(detect.FailStatusCode))
		})
	})

	when("there is NOT a file with suffix runtimeconfig.json", func() {
		it("fails detect", func() {
			buf := bytes.Buffer{}
			factory.Detect.Logger.Logger = specLogger.NewLogger(&buf, &buf)

			code, err := runDetect(factory.Detect)
			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(ContainSubstring("*.runtimeconfig.json file not found and expecting only a single *.csproj file in the app directory"))
			Expect(code).To(Equal(detect.FailStatusCode))
		})
	})

}
