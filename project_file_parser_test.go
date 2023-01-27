package dotnetexecute_test

import (
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
			path, err = os.MkdirTemp("", "workingDir")
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
				Expect(os.WriteFile(filepath.Join(path, "app.csproj"), nil, 0600)).To(Succeed())
			})

			it("returns the path to it", func() {
				projectFilePath, err := parser.FindProjectFile(path)
				Expect(err).NotTo(HaveOccurred())
				Expect(projectFilePath).To(Equal(filepath.Join(path, "app.csproj")))
			})
		})

		context("when there is an fsproj", func() {
			it.Before(func() {
				Expect(os.WriteFile(filepath.Join(path, "app.fsproj"), nil, 0600)).To(Succeed())
			})

			it("returns the path to it", func() {
				projectFilePath, err := parser.FindProjectFile(path)
				Expect(err).NotTo(HaveOccurred())
				Expect(projectFilePath).To(Equal(filepath.Join(path, "app.fsproj")))
			})
		})

		context("when there is a vbproj", func() {
			it.Before(func() {
				Expect(os.WriteFile(filepath.Join(path, "app.vbproj"), nil, 0600)).To(Succeed())
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

	context("NodeIsRequired", func() {
		var path string

		it.Before(func() {
			file, err := os.CreateTemp("", "app.csproj")
			Expect(err).NotTo(HaveOccurred())
			defer file.Close()

			path = file.Name()
		})

		it.After(func() {
			Expect(os.RemoveAll(path)).To(Succeed())
		})

		context("when project includes target commands that invoke node", func() {
			it.Before(func() {
				Expect(os.WriteFile(path, []byte(`
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
				Expect(os.WriteFile(path, []byte(`
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
				Expect(os.WriteFile(path, []byte(`
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
					Expect(os.WriteFile(path, []byte("%%%"), 0644)).To(Succeed())
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
			file, err := os.CreateTemp("", "app.csproj")
			Expect(err).NotTo(HaveOccurred())
			defer file.Close()

			path = file.Name()
		})

		it.After(func() {
			Expect(os.RemoveAll(path)).To(Succeed())
		})

		context("when project includes target commands that invoke npm", func() {
			it.Before(func() {
				Expect(os.WriteFile(path, []byte(`
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
				Expect(os.WriteFile(path, []byte(`
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
