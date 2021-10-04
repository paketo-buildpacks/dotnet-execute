package dotnetexecute_test

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	dotnetexecute "github.com/paketo-buildpacks/dotnet-execute"
	"github.com/paketo-buildpacks/dotnet-execute/fakes"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/scribe"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layersDir          string
		workingDir         string
		cnbDir             string
		buffer             *bytes.Buffer
		configParser       *fakes.ConfigParser
		buildpackYMLParser *fakes.BuildpackConfigParser

		build packit.BuildFunc
	)

	it.Before(func() {
		var err error
		layersDir, err = ioutil.TempDir("", "layers")
		Expect(err).NotTo(HaveOccurred())

		cnbDir, err = ioutil.TempDir("", "cnb")
		Expect(err).NotTo(HaveOccurred())

		workingDir, err = ioutil.TempDir("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		configParser = &fakes.ConfigParser{}

		buildpackYMLParser = &fakes.BuildpackConfigParser{}

		buffer = bytes.NewBuffer(nil)
		logger := scribe.NewLogger(buffer)

		build = dotnetexecute.Build(buildpackYMLParser, configParser, logger)
	})

	it.After(func() {
		Expect(os.RemoveAll(layersDir)).To(Succeed())
		Expect(os.RemoveAll(cnbDir)).To(Succeed())
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	context("The project path is set via buildpack.yml", func() {
		it.Before(func() {
			buildpackYMLParser.ParseProjectPathCall.Returns.ProjectPath = "src/proj1"
			configParser.ParseCall.Returns.RuntimeConfig = dotnetexecute.RuntimeConfig{
				Path:       filepath.Join(workingDir, "myapp.runtimeconfig.json"),
				AppName:    "myapp",
				Executable: true,
			}
		})

		it("Logs a deprecation warning to the user", func() {
			_, err := build(packit.BuildContext{
				WorkingDir: workingDir,
				CNBPath:    cnbDir,
				Stack:      "some-stack",
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "1.2.3",
				},
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{},
				},
				Layers: packit.Layers{Path: layersDir},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(buildpackYMLParser.ParseProjectPathCall.Receives.Path).To(Equal(filepath.Join(workingDir, "buildpack.yml")))
			Expect(buffer.String()).To(ContainSubstring("WARNING: Setting the project path through buildpack.yml will be deprecated soon in Dotnet Execute Buildpack v2.0.0"))
			Expect(buffer.String()).To(ContainSubstring("Please specify the project path through the $BP_DOTNET_PROJECT_PATH environment variable instead. See README.md or the documentation on paketo.io for more information."))
		})
	})

	context("the app is a framework-dependent or self-contained executable", func() {
		it.Before(func() {
			configParser.ParseCall.Returns.RuntimeConfig = dotnetexecute.RuntimeConfig{
				Path:       filepath.Join(workingDir, "myapp.runtimeconfig.json"),
				AppName:    "myapp",
				Executable: true,
			}
		})

		it("returns a result that builds correctly", func() {
			result, err := build(packit.BuildContext{
				WorkingDir: workingDir,
				CNBPath:    cnbDir,
				Stack:      "some-stack",
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "some-version",
				},
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{},
				},
				Layers: packit.Layers{Path: layersDir},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(Equal(packit.BuildResult{
				Plan: packit.BuildpackPlan{
					Entries: nil,
				},
				Layers: nil,
				Launch: packit.LaunchMetadata{
					Processes: []packit.Process{
						{
							Type:    "web",
							Command: fmt.Sprintf("%s --urls http://0.0.0.0:${PORT:-8080}", filepath.Join(workingDir, "myapp")),
						},
					},
				},
			}))
		})
	})

	context("the app is a framework dependent deployment", func() {
		it.Before(func() {
			configParser.ParseCall.Returns.RuntimeConfig = dotnetexecute.RuntimeConfig{
				Path:       filepath.Join(workingDir, "myapp.runtimeconfig.json"),
				AppName:    "myapp",
				Executable: false,
			}
			Expect(ioutil.WriteFile(filepath.Join(workingDir, "myapp.dll"), nil, os.ModePerm)).To(Succeed())
		})

		it.After(func() {
			Expect(os.RemoveAll(filepath.Join(workingDir, "myapp.dll"))).To(Succeed())
		})

		it("returns a result that builds correctly", func() {
			result, err := build(packit.BuildContext{
				WorkingDir: workingDir,
				CNBPath:    cnbDir,
				Stack:      "some-stack",
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "some-version",
				},
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{},
				},
				Layers: packit.Layers{Path: layersDir},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(Equal(packit.BuildResult{
				Plan: packit.BuildpackPlan{
					Entries: nil,
				},
				Layers: nil,
				Launch: packit.LaunchMetadata{
					Processes: []packit.Process{
						{
							Type:    "web",
							Command: fmt.Sprintf("dotnet %s --urls http://0.0.0.0:${PORT:-8080}", filepath.Join(workingDir, "myapp.dll")),
						},
					},
				},
			}))
		})
	})

	context("when BP_LIVE_RELOAD_ENABLED=true", func() {
		it.Before(func() {
			Expect(os.Setenv("BP_LIVE_RELOAD_ENABLED", "true")).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(workingDir, "myapp.dll"), nil, os.ModePerm)).To(Succeed())
			configParser.ParseCall.Returns.RuntimeConfig = dotnetexecute.RuntimeConfig{
				Path:       filepath.Join(workingDir, "myapp.runtimeconfig.json"),
				AppName:    "myapp",
				Executable: false,
			}
		})

		it.After(func() {
			Expect(os.Unsetenv("BP_LIVE_RELOAD_ENABLED")).To(Succeed())
			Expect(os.RemoveAll(filepath.Join(workingDir, "myapp.dll"))).To(Succeed())
		})

		it("wraps the start command with watchexec", func() {
			result, err := build(packit.BuildContext{
				WorkingDir: workingDir,
				CNBPath:    cnbDir,
				Stack:      "some-stack",
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "some-version",
				},
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{},
				},
				Layers: packit.Layers{Path: layersDir},
			})
			Expect(err).NotTo(HaveOccurred())

			startCommand := fmt.Sprintf("dotnet %s --urls http://0.0.0.0:${PORT:-8080}", filepath.Join(workingDir, "myapp.dll"))
			Expect(result).To(Equal(packit.BuildResult{
				Plan: packit.BuildpackPlan{
					Entries: nil,
				},
				Layers: nil,
				Launch: packit.LaunchMetadata{
					Processes: []packit.Process{
						{
							Type:    "web",
							Command: fmt.Sprintf(`watchexec --restart --watch %s "%s"`, workingDir, startCommand),
						},
					},
				},
			}))
		})
	})

	context("failure cases", func() {
		context("buildpack.yml parsing fails", func() {
			it.Before(func() {
				buildpackYMLParser.ParseProjectPathCall.Returns.Err = errors.New("error parsing buildpack.yml")
			})

			it("logs a warning", func() {
				_, err := build(packit.BuildContext{
					WorkingDir: workingDir,
					CNBPath:    cnbDir,
					Stack:      "some-stack",
					BuildpackInfo: packit.BuildpackInfo{
						Name:    "Some Buildpack",
						Version: "some-version",
					},
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{},
					},
					Layers: packit.Layers{Path: layersDir},
				})
				Expect(err).To(MatchError(ContainSubstring("error parsing buildpack.yml")))
			})
		})

		context("runtime config parsing fails", func() {
			it.Before(func() {
				configParser.ParseCall.Returns.Error = errors.New("error parsing runtimeconfig.json")
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					WorkingDir: workingDir,
					CNBPath:    cnbDir,
					Stack:      "some-stack",
					BuildpackInfo: packit.BuildpackInfo{
						Name:    "Some Buildpack",
						Version: "some-version",
					},
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{},
					},
					Layers: packit.Layers{Path: layersDir},
				})
				Expect(err).To(MatchError(ContainSubstring("error parsing runtimeconfig.json")))
			})
		})

		context("error when checking for existence of dll file", func() {
			it.Before(func() {
				configParser.ParseCall.Returns.RuntimeConfig = dotnetexecute.RuntimeConfig{
					Path:       filepath.Join(workingDir, "myapp.runtimeconfig.json"),
					AppName:    "myapp",
					Executable: false,
				}
				Expect(os.Chmod(workingDir, 0000)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(workingDir, os.ModePerm)).To(Succeed())
			})
			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					WorkingDir: workingDir,
					CNBPath:    cnbDir,
					Stack:      "some-stack",
					BuildpackInfo: packit.BuildpackInfo{
						Name:    "Some Buildpack",
						Version: "some-version",
					},
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{},
					},
					Layers: packit.Layers{Path: layersDir},
				})
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})

		context("neither executable nor dll file are present (no entrypoint is found)", func() {
			it.Before(func() {
				configParser.ParseCall.Returns.RuntimeConfig = dotnetexecute.RuntimeConfig{
					Path:       filepath.Join(workingDir, "myapp.runtimeconfig.json"),
					AppName:    "myapp",
					Executable: false,
				}
				files, err := filepath.Glob(filepath.Join(workingDir, "*.dll"))
				Expect(err).NotTo(HaveOccurred())
				Expect(files).To(BeEmpty())
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					WorkingDir: workingDir,
					CNBPath:    cnbDir,
					Stack:      "some-stack",
					BuildpackInfo: packit.BuildpackInfo{
						Name:    "Some Buildpack",
						Version: "some-version",
					},
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{},
					},
					Layers: packit.Layers{Path: layersDir},
				})
				Expect(err).To(MatchError(ContainSubstring("no entrypoint [myapp.dll] found")))
			})

		})

		context("parsing the value of BP_LIVE_RELOAD_ENABLED fails", func() {
			it.Before(func() {
				Expect(ioutil.WriteFile(filepath.Join(workingDir, "myapp.dll"), nil, os.ModePerm)).To(Succeed())
				configParser.ParseCall.Returns.RuntimeConfig = dotnetexecute.RuntimeConfig{
					Path:       filepath.Join(workingDir, "myapp.runtimeconfig.json"),
					AppName:    "myapp",
					Executable: false,
				}
				Expect(os.Setenv("BP_LIVE_RELOAD_ENABLED", "%%%")).To(Succeed())
			})

			it.After(func() {
				Expect(os.Unsetenv("BP_LIVE_RELOAD_ENABLED")).To(Succeed())
				Expect(os.RemoveAll(filepath.Join(workingDir, "myapp.dll"))).To(Succeed())
			})

			it("fails", func() {
				_, err := build(packit.BuildContext{
					WorkingDir: workingDir,
				})
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(ContainSubstring("invalid syntax")))
			})
		})
	})
}
