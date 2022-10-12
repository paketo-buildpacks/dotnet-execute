package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testBackwardsCompatibility(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		pack   occam.Pack
		docker occam.Docker
	)

	it.Before(func() {
		pack = occam.NewPack().WithVerbose().WithNoColor()
		docker = occam.NewDocker()
	})

	if !strings.Contains(builder.Local.Stack.ID, "jammy") {
		var (
			image     occam.Image
			container occam.Container

			name   string
			source string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		context("when building a FDD app", func() {
			it("builds and runs successfully", func() {
				var err error
				source, err = occam.Source(filepath.Join("testdata", "fdd_aspnet"))
				Expect(err).NotTo(HaveOccurred())

				var logs fmt.Stringer
				image, logs, err = pack.Build.
					WithPullPolicy("never").
					WithBuildpacks(
						settings.Buildpacks.ICU.Online,
						settings.Buildpacks.DotnetCoreRuntime.Online,
						settings.Buildpacks.DotnetCoreASPNet.Online,
						settings.Buildpacks.DotnetCoreSDK.Online,
						settings.Buildpacks.DotnetExecute.Online,
					).
					Execute(name, source)
				Expect(err).ToNot(HaveOccurred(), logs.String)

				container, err = docker.Container.Run.
					WithEnv(map[string]string{"PORT": "9090"}).
					WithPublish("9090").
					WithPublishAll().
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				Eventually(func() string {
					logs, _ := docker.Container.Logs.Execute(container.ID)
					return logs.String()
				}).Should(ContainSubstring(`Setting ASPNETCORE_URLS=http://0.0.0.0:9090`))

				Eventually(container).Should(Serve(ContainSubstring("Hello World!")).OnPort(9090))
			})
		})

		context("when building a FDE app", func() {
			it("builds and runs successfully", func() {
				var err error
				source, err = occam.Source(filepath.Join("testdata", "framework_dependent_executable"))
				Expect(err).NotTo(HaveOccurred())

				var logs fmt.Stringer
				image, logs, err = pack.Build.
					WithPullPolicy("never").
					WithBuildpacks(
						settings.Buildpacks.ICU.Online,
						settings.Buildpacks.DotnetCoreRuntime.Online,
						settings.Buildpacks.DotnetExecute.Online,
					).
					Execute(name, source)
				Expect(err).ToNot(HaveOccurred(), logs.String)

				container, err = docker.Container.Run.Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				Eventually(func() string {
					logs, _ := docker.Container.Logs.Execute(container.ID)
					return logs.String()
				}).Should(Equal(`Setting ASPNETCORE_URLS=http://0.0.0.0:8080
Hello World!
`))
			})
		})

		context("when building a source app", func() {
			it("builds and runs successfully", func() {
				var err error
				source, err = occam.Source(filepath.Join("testdata", "source_app"))
				Expect(err).NotTo(HaveOccurred())

				var logs fmt.Stringer
				image, logs, err = pack.Build.
					WithPullPolicy("never").
					WithBuildpacks(
						settings.Buildpacks.ICU.Online,
						settings.Buildpacks.DotnetCoreRuntime.Online,
						settings.Buildpacks.DotnetCoreASPNet.Online,
						settings.Buildpacks.DotnetCoreSDK.Online,
						settings.Buildpacks.DotnetPublish.Online,
						settings.Buildpacks.DotnetExecute.Online,
					).
					Execute(name, source)
				Expect(err).ToNot(HaveOccurred(), logs.String)

				container, err = docker.Container.Run.
					WithEnv(map[string]string{"PORT": "8080"}).
					WithPublish("8080").
					WithPublishAll().
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				Eventually(container).Should(Serve(ContainSubstring("simple_3_0_app")).OnPort(8080))
			})
		})
	}
}
