package dotnetexecute_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	dotnetexecute "github.com/paketo-buildpacks/dotnet-execute"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testRuntimeConfigParser(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir string
		parser     dotnetexecute.RuntimeConfigParser
	)

	it.Before(func() {
		var err error
		workingDir, err = ioutil.TempDir("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		err = ioutil.WriteFile(filepath.Join(workingDir, "some-app.runtimeconfig.json"), []byte(`{}`), 0600)
		Expect(err).NotTo(HaveOccurred())

		parser = dotnetexecute.NewRuntimeConfigParser()
	})

	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	context("Parse", func() {
		it("parses the runtime config", func() {
			config, err := parser.Parse(filepath.Join(workingDir, "*.runtimeconfig.json"))
			Expect(err).NotTo(HaveOccurred())
			Expect(config).To(Equal(dotnetexecute.RuntimeConfig{
				Path:    filepath.Join(workingDir, "some-app.runtimeconfig.json"),
				AppName: "some-app",
			}))
		})

		context("when the runtime config includes comments", func() {
			it.Before(func() {
				Expect(ioutil.WriteFile(filepath.Join(workingDir, "some-app.runtimeconfig.json"), []byte(`{
					"runtimeOptions": {
						/*
						Multi line
						Comment
						*/
						"configProperties": {
							"System.GC.Server": true
						}
						// comment here ok?
					}
				}`), 0600)).To(Succeed())
			})

			it("parses the runtime config", func() {
				config, err := parser.Parse(filepath.Join(workingDir, "*.runtimeconfig.json"))
				Expect(err).NotTo(HaveOccurred())
				Expect(config).To(Equal(dotnetexecute.RuntimeConfig{
					Path:    filepath.Join(workingDir, "some-app.runtimeconfig.json"),
					AppName: "some-app",
				}))
			})
		})

		context("when the app includes an executable", func() {
			it.Before(func() {
				Expect(ioutil.WriteFile(filepath.Join(workingDir, "some-app"), nil, 0700)).To(Succeed())
			})

			it("reports that the app includes an executable", func() {
				config, err := parser.Parse(filepath.Join(workingDir, "*.runtimeconfig.json"))
				Expect(err).NotTo(HaveOccurred())
				Expect(config.Executable).To(BeTrue())
			})
		})

		context("when the app does not include an executable", func() {
			it("reports that the app does not include an executable", func() {
				config, err := parser.Parse(filepath.Join(workingDir, "*.runtimeconfig.json"))
				Expect(err).NotTo(HaveOccurred())
				Expect(config.Executable).To(BeFalse())
			})
		})

		context("when the runtime framework version is specified", func() {
			it.Before(func() {
				Expect(ioutil.WriteFile(filepath.Join(workingDir, "some-app.runtimeconfig.json"), []byte(`{
					"runtimeOptions": {
						"framework": {
							"name": "Microsoft.NETCore.App",
							"version": "2.1.3"
						}
					}
				}`), 0600)).To(Succeed())
			})

			it("returns the runtime version", func() {
				config, err := parser.Parse(filepath.Join(workingDir, "*.runtimeconfig.json"))
				Expect(err).NotTo(HaveOccurred())
				Expect(config.Version).To(Equal("2.1.3"))
			})
		})

		context("when runtime frameworks are specified", func() {
			it.Before(func() {
				Expect(ioutil.WriteFile(filepath.Join(workingDir, "some-app.runtimeconfig.json"), []byte(`{
  "runtimeOptions": {
    "frameworks": [
      {
        "name": "Microsoft.NETCore.App",
        "version": "2.1.3"
      }
    ]
  }
}`), 0600)).To(Succeed())
			})

			it("returns the runtime version", func() {
				config, err := parser.Parse(filepath.Join(workingDir, "*.runtimeconfig.json"))
				Expect(err).NotTo(HaveOccurred())
				Expect(config.Version).To(Equal("2.1.3"))
				Expect(config.UsesASPNet).To(BeFalse())
			})
		})

		context("when runtime frameworks include AspNetCore", func() {
			it.Before(func() {
				Expect(ioutil.WriteFile(filepath.Join(workingDir, "some-app.runtimeconfig.json"), []byte(`{
  "runtimeOptions": {
    "frameworks": [
      {
        "name": "Microsoft.NETCore.App",
        "version": "2.1.3"
      },
      {
        "name": "Microsoft.AspNetCore.App",
        "version": "2.1.3"
      }
    ]
  }
}`), 0600)).To(Succeed())
			})

			it("returns the runtime version and detects that ASP.NET is required", func() {
				config, err := parser.Parse(filepath.Join(workingDir, "*.runtimeconfig.json"))
				Expect(err).NotTo(HaveOccurred())
				Expect(config.Version).To(Equal("2.1.3"))
				Expect(config.UsesASPNet).To(BeTrue())
			})
		})

		context("when the runtime framework is specified with no version", func() {
			it.Before(func() {
				Expect(ioutil.WriteFile(filepath.Join(workingDir, "some-app.runtimeconfig.json"), []byte(`{
					"runtimeOptions": {
						"framework": {
							"name": "Microsoft.NETCore.App"
						}
					}
				}`), 0600)).To(Succeed())
			})

			it("returns that version", func() {
				config, err := parser.Parse(filepath.Join(workingDir, "*.runtimeconfig.json"))
				Expect(err).NotTo(HaveOccurred())
				Expect(config.Version).To(Equal("*"))
			})
		})

		context("when the app requires ASP.Net via Microsoft.AspNetCore.App", func() {
			it.Before(func() {
				Expect(ioutil.WriteFile(filepath.Join(workingDir, "some-app.runtimeconfig.json"), []byte(`{
					"runtimeOptions": {
						"framework": {
							"name": "Microsoft.AspNetCore.App",
							"version": "2.1.0"
						}
					}
				}`), 0600)).To(Succeed())
			})

			it("reports that the app requires ASP.Net", func() {
				config, err := parser.Parse(filepath.Join(workingDir, "*.runtimeconfig.json"))
				Expect(err).NotTo(HaveOccurred())
				Expect(config.Version).To(Equal("2.1.0"))
				Expect(config.UsesASPNet).To(BeTrue())
			})
		})

		context("the runtimeconfig.json does not exist", func() {
			it.Before(func() {
				Expect(os.RemoveAll(filepath.Join(workingDir, "some-app.runtimeconfig.json"))).NotTo(HaveOccurred())
			})

			it("returns the os.ErrNotExist", func() {
				_, err := parser.Parse(filepath.Join(workingDir, "*.runtimeconfig.json"))
				Expect(errors.Is(err, os.ErrNotExist)).To(BeTrue())
			})
		})

		context("failure cases", func() {
			context("when given an invalid glob", func() {
				it("returns an error", func() {
					_, err := parser.Parse("[-]")
					Expect(err).To(MatchError(`failed to find *.runtimeconfig.json: syntax error in pattern: "[-]"`))
				})
			})

			context("when frameworks array is specified in runtimeconfig.json", func() {
				context("when ASP.NET and .NET runtime versions do not match", func() {
					it.Before(func() {
						Expect(ioutil.WriteFile(filepath.Join(workingDir, "some-app.runtimeconfig.json"), []byte(`{
  "runtimeOptions": {
    "frameworks": [
      {
        "name": "Microsoft.NETCore.App",
        "version": "2.1.3"
      },
      {
        "name": "Microsoft.AspNetCore.App",
        "version": "2.0.0"
      }
    ]
  }
}`), 0600)).To(Succeed())
					})

					it("returns an error", func() {
						_, err := parser.Parse(filepath.Join(workingDir, "*.runtimeconfig.json"))
						Expect(err).To(MatchError(ContainSubstring("cannot satisfy mismatched runtimeconfig.json version requirements ('2.1.3' and '2.0.0')")))
					})
				})

				context("when there are multiple ASP.NET framework entries", func() {
					it.Before(func() {
						Expect(ioutil.WriteFile(filepath.Join(workingDir, "some-app.runtimeconfig.json"), []byte(`{
  "runtimeOptions": {
    "frameworks": [
      {
        "name": "Microsoft.AspNetCore.App",
        "version": "2.1.3"
      },
      {
        "name": "Microsoft.AspNetCore.App",
        "version": "2.0.0"
      }
    ]
  }
}`), 0600)).To(Succeed())
					})

					it("returns an error", func() {
						_, err := parser.Parse(filepath.Join(workingDir, "*.runtimeconfig.json"))
						Expect(err).To(MatchError(ContainSubstring("malformed runtimeconfig.json: multiple 'Microsoft.AspNetCore.App' frameworks specified")))
					})
				})

				context("and there are multiple NETCore framework entries", func() {
					it.Before(func() {
						Expect(ioutil.WriteFile(filepath.Join(workingDir, "some-app.runtimeconfig.json"), []byte(`{
  "runtimeOptions": {
    "frameworks": [
      {
        "name": "Microsoft.NETCore.App",
        "version": "2.1.3"
      },
      {
        "name": "Microsoft.NETCore.App",
        "version": "2.0.0"
      }
    ]
  }
}`), 0600)).To(Succeed())
					})

					it("returns an error", func() {
						_, err := parser.Parse(filepath.Join(workingDir, "*.runtimeconfig.json"))
						Expect(err).To(MatchError(ContainSubstring("malformed runtimeconfig.json: multiple 'Microsoft.NETCore.App' frameworks specified")))
					})
				})
			})

			context("when there are multiple runtimeconfig.json files", func() {
				it.Before(func() {
					Expect(ioutil.WriteFile(filepath.Join(workingDir, "other-app.runtimeconfig.json"), []byte(`{}`), 0600)).To(Succeed())
				})

				it("returns an error", func() {
					_, err := parser.Parse(filepath.Join(workingDir, "*.runtimeconfig.json"))
					Expect(err).To(MatchError(ContainSubstring("multiple *.runtimeconfig.json files present")))
					Expect(err).To(MatchError(ContainSubstring("some-app.runtimeconfig.json")))
					Expect(err).To(MatchError(ContainSubstring("other-app.runtimeconfig.json")))
				})
			})

			context("the runtimeconfig.json file cannot be minimized", func() {
				it.Before(func() {
					Expect(ioutil.WriteFile(filepath.Join(workingDir, "some-app.runtimeconfig.json"), []byte("var x = /hello"), 0600)).To(Succeed())
				})

				it("returns an error", func() {
					_, err := parser.Parse(filepath.Join(workingDir, "*.runtimeconfig.json"))
					Expect(err).To(MatchError(ContainSubstring("unterminated regular expression literal")))
				})
			})

			context("the runtimeconfig.json file cannot be parsed", func() {
				it.Before(func() {
					Expect(ioutil.WriteFile(filepath.Join(workingDir, "some-app.runtimeconfig.json"), []byte(`%%%`), 0600)).To(Succeed())
				})

				it("returns an error", func() {
					_, err := parser.Parse(filepath.Join(workingDir, "*.runtimeconfig.json"))
					Expect(err).To(MatchError(ContainSubstring("invalid character")))
				})
			})
		})
	})
}
