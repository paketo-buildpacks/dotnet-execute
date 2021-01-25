package dotnetexecute_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	dotnetexecute "github.com/paketo-buildpacks/dotnet-execute"
	"github.com/sclevine/spec"
)

func testProjectFileParser(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
		parser dotnetexecute.ProjectFileParser
	)

	it.Before(func() {
		parser = dotnetexecute.NewProjectFileParser()
	})

	context("FindProjectFile", func() {
		var path string
		it.Before(func() {
			var err error
			path, err = ioutil.TempDir("", "workingDir")
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(os.RemoveAll(path)).To(Succeed())
		})

		it("returns an empty string and no error", func() {
			projectFilePath, err := parser.FindProjectFile(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(projectFilePath).To(Equal(""))
		})

		context("when there is a csproj", func() {
			it.Before(func() {
				Expect(ioutil.WriteFile(filepath.Join(path, "app.csproj"), nil, 0600)).To(Succeed())
			})

			it("returns the path to it", func() {
				projectFilePath, err := parser.FindProjectFile(path)
				Expect(err).NotTo(HaveOccurred())
				Expect(projectFilePath).To(Equal(filepath.Join(path, "app.csproj")))
			})
		})

		context("when there is an fsproj", func() {
			it.Before(func() {
				Expect(ioutil.WriteFile(filepath.Join(path, "app.fsproj"), nil, 0600)).To(Succeed())
			})

			it("returns the path to it", func() {
				projectFilePath, err := parser.FindProjectFile(path)
				Expect(err).NotTo(HaveOccurred())
				Expect(projectFilePath).To(Equal(filepath.Join(path, "app.fsproj")))
			})
		})

		context("when there is a vbproj", func() {
			it.Before(func() {
				Expect(ioutil.WriteFile(filepath.Join(path, "app.vbproj"), nil, 0600)).To(Succeed())
			})

			it("returns the path to it", func() {
				projectFilePath, err := parser.FindProjectFile(path)
				Expect(err).NotTo(HaveOccurred())
				Expect(projectFilePath).To(Equal(filepath.Join(path, "app.vbproj")))
			})
		})

		context("failure cases", func() {
			context("when file pattern matching fails", func() {
				it("returns the error", func() {
					_, err := parser.FindProjectFile(`\`)
					Expect(err).To(MatchError("syntax error in pattern"))
				})
			})

		})
	})

	context("ParseVersion", func() {
		var path string

		it.Before(func() {
			file, err := ioutil.TempFile("", "app.csproj")
			Expect(err).NotTo(HaveOccurred())
			defer file.Close()

			_, err = file.WriteString(`
				<Project>
					<PropertyGroup>
						<Obfuscate>true</Obfuscate>
					</PropertyGroup>
					<PropertyGroup>
						<RuntimeFrameworkVersion>1.2.3</RuntimeFrameworkVersion>
					</PropertyGroup>
				</Project>
			`)
			Expect(err).NotTo(HaveOccurred())

			path = file.Name()
		})

		it.After(func() {
			Expect(os.Remove(path)).To(Succeed())
		})

		it("parses the dotnet runtime version from the ?sproj file", func() {
			version, err := parser.ParseVersion(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(version).To(Equal("1.2.3"))
		})

		context("when the RuntimeFrameworkVersion is not specified", func() {
			context("when TargetFramework is of the syntax netcoreapp<x>.<y>", func() {
				it.Before(func() {
					err := ioutil.WriteFile(path, []byte(`
					<Project>
						<PropertyGroup>
							<TargetFramework>netcoreapp1.2</TargetFramework>
						</PropertyGroup>
					</Project>
				`), 0600)
					Expect(err).NotTo(HaveOccurred())
				})

				it("falls back to using the TargetFramework version", func() {
					version, err := parser.ParseVersion(path)
					Expect(err).NotTo(HaveOccurred())
					Expect(version).To(Equal("1.2.0"))
				})
			})

			context("when TargetFramework is of the syntax net<x>.<y>", func() {
				it.Before(func() {
					err := ioutil.WriteFile(path, []byte(`
					<Project>
						<PropertyGroup>
							<TargetFramework>net5.0</TargetFramework>
						</PropertyGroup>
					</Project>
				`), 0600)
					Expect(err).NotTo(HaveOccurred())
				})

				it("falls back to using the TargetFramework version", func() {
					version, err := parser.ParseVersion(path)
					Expect(err).NotTo(HaveOccurred())
					Expect(version).To(Equal("5.0.0"))
				})
			})

			context("when TargetFramework is of the syntax net<x>.<y>-<platform>", func() {
				it.Before(func() {
					err := ioutil.WriteFile(path, []byte(`
					<Project>
						<PropertyGroup>
							<TargetFramework>net5.0-someplatform</TargetFramework>
						</PropertyGroup>
					</Project>
				`), 0600)
					Expect(err).NotTo(HaveOccurred())
				})

				it("falls back to using the TargetFramework version", func() {
					version, err := parser.ParseVersion(path)
					Expect(err).NotTo(HaveOccurred())
					Expect(version).To(Equal("5.0.0"))
				})
			})
		})

		context("failure cases", func() {
			context("when the project file does not exist", func() {
				it("returns an error", func() {
					_, err := parser.ParseVersion("no-such-file")
					Expect(err).To(MatchError(MatchRegexp(`failed to read project file: .* no such file or directory`)))
				})
			})

			context("when the project file is malformed", func() {
				it.Before(func() {
					err := ioutil.WriteFile(path, []byte(`<<< %%%`), 0600)
					Expect(err).NotTo(HaveOccurred())
				})

				it("returns an error", func() {
					_, err := parser.ParseVersion(path)
					Expect(err).To(MatchError(MatchRegexp(`failed to parse project file: XML syntax error .*`)))
				})
			})

			context("when the project file does not contain a version", func() {
				it.Before(func() {
					err := ioutil.WriteFile(path, []byte(`<Project></Project>`), 0600)
					Expect(err).NotTo(HaveOccurred())
				})

				it("returns an error", func() {
					_, err := parser.ParseVersion(path)
					Expect(err).To(MatchError("failed to find version in project file: missing TargetFramework property"))
				})
			})
		})
	})

	context("ASPNetIsRequired", func() {
		var path string

		it.Before(func() {
			file, err := ioutil.TempFile("", "app.csproj")
			Expect(err).NotTo(HaveOccurred())
			defer file.Close()

			path = file.Name()
		})

		it.After(func() {
			Expect(os.RemoveAll(path)).To(Succeed())
		})

		context("when project SDK is Microsoft.NET.Sdk.Web", func() {
			it.Before(func() {
				Expect(ioutil.WriteFile(path, []byte(`<Project Sdk="Microsoft.NET.Sdk.Web"></Project>`), 0600)).To(Succeed())
			})

			it("returns true", func() {
				needsAspnt, err := parser.ASPNetIsRequired(path)
				Expect(err).NotTo(HaveOccurred())

				Expect(needsAspnt).To(BeTrue())
			})
		})

		context("when project PackageReference is Microsoft.AspNetCore.App", func() {
			it.Before(func() {
				Expect(ioutil.WriteFile(path, []byte(`
<Project Sdk="Microsoft.NET.Sdk">
<ItemGroup>
	<PackageReference Include="Microsoft.AspNetCore.App"/>
</ItemGroup>
</Project>
`), 0600)).To(Succeed())
			})

			it("returns true", func() {
				needsAspnt, err := parser.ASPNetIsRequired(path)
				Expect(err).NotTo(HaveOccurred())

				Expect(needsAspnt).To(BeTrue())
			})
		})

		context("when project PackageReference is Microsoft.AspNetCore.All", func() {
			it.Before(func() {
				Expect(ioutil.WriteFile(path, []byte(`
<Project Sdk="Microsoft.NET.Sdk">
<ItemGroup>
	<PackageReference Include="Microsoft.AspNetCore.All"/>
</ItemGroup>
</Project>
`), 0600)).To(Succeed())
			})

			it("returns true", func() {
				needsAspnt, err := parser.ASPNetIsRequired(path)
				Expect(err).NotTo(HaveOccurred())
				Expect(needsAspnt).To(BeTrue())
			})
		})

		context("failure cases", func() {
			context("when the file can not be opened", func() {
				it.Before(func() {
					Expect(os.RemoveAll(path)).To(Succeed())
				})

				it("errors", func() {
					_, err := parser.ASPNetIsRequired(path)
					Expect(err.Error()).To(ContainSubstring("failed to open"))
				})
			})

			context("when the file can not be decoded", func() {
				it.Before(func() {
					Expect(ioutil.WriteFile(path, []byte("%%%"), 0644)).To(Succeed())
				})

				it("errors", func() {
					_, err := parser.ASPNetIsRequired(path)
					Expect(err.Error()).To(ContainSubstring("failed to decode"))
				})
			})
		})
	})

	context("NodeIsRequired", func() {
		var path string

		it.Before(func() {
			file, err := ioutil.TempFile("", "app.csproj")
			Expect(err).NotTo(HaveOccurred())
			defer file.Close()

			path = file.Name()
		})

		it.After(func() {
			Expect(os.RemoveAll(path)).To(Succeed())
		})

		context("when project includes target commands that invoke node", func() {
			it.Before(func() {
				Expect(ioutil.WriteFile(path, []byte(`
					<Project>
						<Target Name="first-target">
							<Exec Command="echo hello" />
						</Target>
						<Target Name="second-target">
							<Exec Command="node --version" />
						</Target>
					</Project>
				`), 0600)).To(Succeed())
			})

			it("returns true", func() {
				needNode, err := parser.NodeIsRequired(path)
				Expect(err).NotTo(HaveOccurred())

				Expect(needNode).To(BeTrue())
			})
		})

		context("when project does NOT include target commands that invoke node", func() {
			it.Before(func() {
				Expect(ioutil.WriteFile(path, []byte(`
					<Project>
						<Target Name="first-target">
							<Exec Command="echo hello" />
						</Target>
						<Target Name="second-target">
							<Exec Command="echo goodbye" />
						</Target>
					</Project>
				`), 0600)).To(Succeed())
			})

			it("returns false", func() {
				needNode, err := parser.NodeIsRequired(path)
				Expect(err).NotTo(HaveOccurred())

				Expect(needNode).To(BeFalse())
			})
		})

		context("when project includes target commands that invoke npm", func() {
			it.Before(func() {
				Expect(ioutil.WriteFile(path, []byte(`
					<Project>
						<Target Name="first-target">
							<Exec Command="echo hello" />
						</Target>
						<Target Name="second-target">
							<Exec Command="npm install" />
						</Target>
					</Project>
				`), 0600)).To(Succeed())
			})

			it("returns true", func() {
				needNode, err := parser.NodeIsRequired(path)
				Expect(err).NotTo(HaveOccurred())

				Expect(needNode).To(BeTrue())
			})
		})

		context("failure cases", func() {
			context("when the file can not be opened", func() {
				it.Before(func() {
					Expect(os.RemoveAll(path)).To(Succeed())
				})

				it("errors", func() {
					_, err := parser.NodeIsRequired(path)
					Expect(err.Error()).To(ContainSubstring("failed to open"))
				})
			})

			context("when the file can not be decoded", func() {
				it.Before(func() {
					Expect(ioutil.WriteFile(path, []byte("%%%"), 0644)).To(Succeed())
				})

				it("errors", func() {
					_, err := parser.NodeIsRequired(path)
					Expect(err.Error()).To(ContainSubstring("failed to decode"))
				})
			})

		})
	})

	context("NPMIsRequired", func() {
		var path string

		it.Before(func() {
			file, err := ioutil.TempFile("", "app.csproj")
			Expect(err).NotTo(HaveOccurred())
			defer file.Close()

			path = file.Name()
		})

		it.After(func() {
			Expect(os.RemoveAll(path)).To(Succeed())
		})

		context("when project includes target commands that invoke npm", func() {
			it.Before(func() {
				Expect(ioutil.WriteFile(path, []byte(`
					<Project>
						<Target Name="first-target">
							<Exec Command="echo hello" />
						</Target>
						<Target Name="second-target">
							<Exec Command="npm install" />
						</Target>
					</Project>
				`), 0600)).To(Succeed())
			})

			it("returns true", func() {
				needNode, err := parser.NPMIsRequired(path)
				Expect(err).NotTo(HaveOccurred())

				Expect(needNode).To(BeTrue())
			})
		})

		context("when project does NOT include target commands that invoke node", func() {
			it.Before(func() {
				Expect(ioutil.WriteFile(path, []byte(`
					<Project>
						<Target Name="first-target">
							<Exec Command="echo hello" />
						</Target>
						<Target Name="second-target">
							<Exec Command="echo goodbye" />
						</Target>
					</Project>
				`), 0600)).To(Succeed())
			})

			it("returns false", func() {
				needNode, err := parser.NPMIsRequired(path)
				Expect(err).NotTo(HaveOccurred())

				Expect(needNode).To(BeFalse())
			})
		})
	})
}
