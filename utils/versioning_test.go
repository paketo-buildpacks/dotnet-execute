package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/libcfbuildpack/test"
	. "github.com/onsi/gomega"

	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitVersioning(t *testing.T) {
	spec.Run(t, "Detect", testVersioning, spec.Report(report.Terminal{}))
}

func testVersioning(t *testing.T, when spec.G, it spec.S) {
	var (
		factory                    *test.BuildFactory
		stubDotnetFrameworkFixture = filepath.Join("testdata", "stub-dotnet-framework.tar.xz")
		framework                  string
	)

	it.Before(func() {
		RegisterTestingT(t)
		framework = "dotnet-framework"
		factory = test.NewBuildFactory(t)
		factory.AddDependencyWithVersion(framework, "2.2.4", stubDotnetFrameworkFixture)
		factory.AddDependencyWithVersion(framework, "2.2.5", stubDotnetFrameworkFixture)
		factory.AddDependencyWithVersion(framework, "2.3.0", stubDotnetFrameworkFixture)
	})

	when("checking buildpack version compatiblity", func() {
		when("the buildpack.yml version is not a mask but is still compatible version with app runtime version", func() {
			it("does not return an error", func() {
				runtimeConfigVersion := "2.0.0"
				buildpackYamlVersion := "2.1.13"

				err := BuildpackYAMLVersionCheck(runtimeConfigVersion, buildpackYamlVersion)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		when("the buildpack.yml version mask is a compatible version with app runtime version", func() {
			it("does not return an error ", func() {
				runtimeConfigVersion := "2.0.0"
				buildpackYamlVersion := "2.1.*"

				err := BuildpackYAMLVersionCheck(runtimeConfigVersion, buildpackYamlVersion)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		when("the buildpack.yml major and runtime major are not equal", func() {
			it("returns an error ", func() {
				runtimeConfigVersion := "2.0.0"
				buildpackYamlVersion := "3.0.*"

				err := BuildpackYAMLVersionCheck(runtimeConfigVersion, buildpackYamlVersion)
				Expect(err).To(Equal(fmt.Errorf("major versions of runtimes do not match between buildpack.yml and runtimeconfig.json")))
			})
		})

		when("the buildpack.yml minor is less than runtime minor", func() {
			it("returns an error", func() {
				runtimeConfigVersion := "2.2.0"
				buildpackYamlVersion := "2.1.*"

				err := BuildpackYAMLVersionCheck(runtimeConfigVersion, buildpackYamlVersion)
				Expect(err).To(Equal(fmt.Errorf("the minor version of the runtimeconfig.json is greater than the minor version of the buildpack.yml")))
			})
		})
	})

	when("getting framework roll forward version", func() {
		when("applyPatches is false", func() {

			var appRoot string
			it.Before(func() {
				appRoot = factory.Build.Application.Root
				runtimeConfigJSONPath := filepath.Join(appRoot, "appName.runtimeconfig.json")
				Expect(ioutil.WriteFile(runtimeConfigJSONPath, []byte(`
	{
	  "runtimeOptions": {
		"tfm": "netcoreapp2.2",
		"framework": {
		  "name": "Microsoft.AspNetCore.App",
		  "version": "2.2.5"
		},
		"applyPatches": false	
	  }
	}
	`), os.ModePerm)).To(Succeed())

			})

			it.After(func() {
				os.RemoveAll(appRoot)
			})

			it("returns a version if rollForwardVersion is found in buildpack.toml", func() {
				rollVersion, err := FrameworkRollForward("2.2.4", framework, factory.Build)
				Expect(err).NotTo(HaveOccurred())
				Expect(rollVersion).To(Equal("2.2.4"))
			})

			it("returns a version if rollForwardVersion has a matching minor with a version found in buildpack.toml and the rollForwardVersion patch is lower", func() {
				rollVersion, err := FrameworkRollForward("2.2.0", framework, factory.Build)
				Expect(err).NotTo(HaveOccurred())
				Expect(rollVersion).To(Equal("2.2.5"))
			})

			// todo: this test will have to change to reflect Micorsoft's new default.
			it("returns a version if rollForwardVersion has a matching major with a version found in buildpack.toml and the rollForwardVersion minor is lower", func() {
				rollVersion, err := FrameworkRollForward("2.1.0", framework, factory.Build)
				Expect(err).NotTo(HaveOccurred())
				Expect(rollVersion).To(Equal("2.3.0"))
			})

			it("returns an version with matching major if rollForwardVersion has a matching minor with a version found in buildpack.toml and the rollForwardVersion patch is higher", func() {
				rollVersion, err := FrameworkRollForward("2.2.6", framework, factory.Build)
				Expect(err).NotTo(HaveOccurred())
				Expect(rollVersion).To(Equal("2.3.0"))
			})

			it("returns an empty string and an error with matching major if rollForwardVersion has a matching major with a version found in buildpack.toml and the rollForwardVersion minor is higher", func() {
				rollVersion, err := FrameworkRollForward("2.4.0", framework, factory.Build)
				Expect(err).To(Equal(fmt.Errorf("no compatible versions found")))
				Expect(rollVersion).To(Equal(""))
			})

			it("returns an empty string and an error when no matching major is found", func() {
				rollVersion, err := FrameworkRollForward("3.0.0", framework, factory.Build)
				Expect(err).To(Equal(fmt.Errorf("no compatible versions found")))
				Expect(rollVersion).To(Equal(""))
			})
		})

		when("applyPatches is true", func() {

			var appRoot string
			it.Before(func() {
				appRoot = factory.Build.Application.Root
				runtimeConfigJSONPath := filepath.Join(appRoot, "appName.runtimeconfig.json")
				Expect(ioutil.WriteFile(runtimeConfigJSONPath, []byte(`
	{
	  "runtimeOptions": {
		"tfm": "netcoreapp2.2",
		"framework": {
		  "name": "Microsoft.AspNetCore.App",
		  "version": "2.2.5"
		},
		"applyPatches": true	
	  }
	}
	`), os.ModePerm)).To(Succeed())

			})

			it.After(func() {
				os.RemoveAll(appRoot)
			})

			it("returns a the latest patch version if rollForwardVersion is found in buildpack.toml but apply patches is true", func() {
				rollVersion, err := FrameworkRollForward("2.2.4", framework, factory.Build)
				Expect(err).NotTo(HaveOccurred())
				Expect(rollVersion).To(Equal("2.2.5"))
			})
		})
	})
}
