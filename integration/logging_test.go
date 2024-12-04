package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testLogging(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		pack   occam.Pack
		docker occam.Docker
	)

	it.Before(func() {
		pack = occam.NewPack().WithNoColor()
		docker = occam.NewDocker()
	})

	context("when building an app with BP_LOG_LEVEL=DEBUG", func() {
		var (
			image occam.Image

			name   string
			source string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		it("presents the expected log output", func() {
			var err error
			source, err = occam.Source(filepath.Join("testdata", "self_contained_8"))
			Expect(err).NotTo(HaveOccurred())

			var logs fmt.Stringer
			image, logs, err = pack.Build.
				WithPullPolicy("never").
				WithBuildpacks(
					settings.Buildpacks.ICU.Online,
					settings.Buildpacks.DotnetExecute.Online,
				).
				WithEnv(map[string]string{
					"BP_LOG_LEVEL": "DEBUG",
				}).
				Execute(name, source)
			Expect(err).ToNot(HaveOccurred(), logs.String)

			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, settings.BuildpackInfo.Buildpack.Name)),
				"  Build configuration:",
				// General matcher since env vars are extracted from map in different order each time
				MatchRegexp("    BP_.*: .*"),
			))
			Expect(logs).To(ContainLines(
				"  Generating SBOM for /workspace",
				MatchRegexp(`      Completed in ([0-9]*(\.[0-9]*)?[a-z]+)+`),
				"",
				"  Writing SBOM in the following format(s):",
				"    application/vnd.cyclonedx+json",
				"    application/spdx+json",
				"    application/vnd.syft+json",
				"",
				"  Assigning launch processes:",
				`    webapp_8 (default): /workspace/webapp_8`,
				"",
				"  Setting up layer 'port-chooser'",
				"    Available at app launch: true",
				"    Available to other buildpacks: false",
				"    Cached for rebuilds: false",
				"",
			))
		})
	})
}
