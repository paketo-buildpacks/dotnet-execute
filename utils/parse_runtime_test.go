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

func TestUnitDetect(t *testing.T) {
	spec.Run(t, "Detect", testDetect, spec.Report(report.Terminal{}))
}

func testDetect(t *testing.T, when spec.G, it spec.S) {
	it.Before(func(){
		RegisterTestingT(t)
	})

	when("when there is a valid runtimeconfig.json", func() {
		it("parses", func() {
			appRoot, err := ioutil.TempDir("", "")
			Expect(err).ToNot(HaveOccurred())
			runtimeConfigJSONPath := filepath.Join(appRoot, "appName.runtimeconfig.json")
			Expect(ioutil.WriteFile(runtimeConfigJSONPath, []byte(`
{
  "runtimeOptions": {
    "tfm": "netcoreapp2.2",
    "framework": {
      "name": "Microsoft.NETCore.App",
      "version": "2.2.5"
    }
  }
}
`), os.ModePerm)).To(Succeed())
			config, err := CreateRuntimeConfig(appRoot)
			parsedConfig := ConfigJSON{}
			parsedConfig.RuntimeOptions.Framework.Name = "Microsoft.NETCore.App"
			parsedConfig.RuntimeOptions.Framework.Version = "2.2.5"
			Expect(err).ToNot(HaveOccurred())
			Expect(config).To(Equal(parsedConfig))
		})

		it("parses when comments are in runtime.json", func() {
			appRoot, err := ioutil.TempDir("", "")
			Expect(err).ToNot(HaveOccurred())
			runtimeConfigJSONPath := filepath.Join(appRoot, "appName.runtimeconfig.json")

			Expect(ioutil.WriteFile(runtimeConfigJSONPath, []byte(`
{
  "runtimeOptions": {
    /*
    Multi line
    Comment
    */
    "tfm": "netcoreapp2.2",
    "framework": {
	  "name": "Microsoft.NETCore.App",
	  "version": "2.2.5"
    },
    // comment here ok?
    "configProperties": {
	  "System.GC.Server": true
    }
  }
}
		`), os.ModePerm)).To(Succeed())
			config, err := CreateRuntimeConfig(appRoot)
			parsedConfig := ConfigJSON{}
			parsedConfig.RuntimeOptions.Framework.Name = "Microsoft.NETCore.App"
			parsedConfig.RuntimeOptions.Framework.Version = "2.2.5"
			Expect(err).ToNot(HaveOccurred())
			Expect(config).To(Equal(parsedConfig))
		})
	})

	when("when there are multiple runtimeconfig.json", func() {
		it("fails fast", func() {
			appRoot, err := ioutil.TempDir("", "")
			Expect(err).ToNot(HaveOccurred())
			runtimeConfigJSONPath := filepath.Join(appRoot, "appName.runtimeconfig.json")
			anotherRuntimeConfigJSONPath := filepath.Join(appRoot, "another.runtimeconfig.json")
			Expect(ioutil.WriteFile(runtimeConfigJSONPath, []byte(`{}`), os.ModePerm)).To(Succeed())
			Expect(ioutil.WriteFile(anotherRuntimeConfigJSONPath, []byte(`{}`), os.ModePerm)).To(Succeed())

			_, err = CreateRuntimeConfig(appRoot)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("multiple *.runtimeconfig.json files present"))
		})
	})

	when("there is not runtimeconfig.json at the given root", func(){
		it("returns an empty config", func(){
			appRoot, err := ioutil.TempDir("", "")
			Expect(err).ToNot(HaveOccurred())

			config, err := CreateRuntimeConfig(appRoot)
			Expect(err).ToNot(HaveOccurred())
			Expect(config).To(Equal(ConfigJSON{}))
		})

	})

}