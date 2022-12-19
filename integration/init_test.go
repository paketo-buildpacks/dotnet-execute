package integration_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/paketo-buildpacks/occam"
	"github.com/paketo-buildpacks/occam/packagers"
	"github.com/paketo-buildpacks/packit/v2/pexec"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
)

var settings struct {
	BuildpackInfo struct {
		Buildpack struct {
			ID   string
			Name string
		}
	}
	Config struct {
		ICU                     string `json:"icu"`
		DotnetCoreSDK           string `json:"dotnet-core-sdk"`
		DotnetCoreASPNetRuntime string `json:"dotnet-core-aspnet-runtime"`
		DotnetPublish           string `json:"dotnet-publish"`
		NodeEngine              string `json:"node-engine"`
		Watchexec               string `json:"watchexec"`
		Vsdbg                   string `json:"vsdbg"`
		// For backwards compatibility tests
		DotnetCoreRuntime string `json:"dotnet-core-runtime"`
		DotnetCoreASPNet  string `json:"dotnet-core-aspnet"`
	}
	Buildpacks struct {
		DotnetExecute struct {
			Online string
		}
		DotnetCoreSDK struct {
			Online string
		}
		DotnetCoreASPNetRuntime struct {
			Online string
		}
		DotnetPublish struct {
			Online string
		}
		ICU struct {
			Online string
		}
		NodeEngine struct {
			Online string
		}
		Watchexec struct {
			Online string
		}
		Vsdbg struct {
			Online string
		}
		// For backwards compatibility tests
		DotnetCoreRuntime struct {
			Online string
		}
		DotnetCoreASPNet struct {
			Online string
		}
	}
}
var builder struct {
	Local struct {
		Stack struct {
			ID string `json:"id"`
		} `json:"stack"`
	} `json:"local_info"`
}

func TestIntegration(t *testing.T) {
	Expect := NewWithT(t).Expect
	format.MaxLength = 0

	file, err := os.Open("../integration.json")
	Expect(err).NotTo(HaveOccurred())

	Expect(json.NewDecoder(file).Decode(&settings.Config)).To(Succeed())
	Expect(file.Close()).To(Succeed())

	file, err = os.Open("../buildpack.toml")
	Expect(err).NotTo(HaveOccurred())

	_, err = toml.NewDecoder(file).Decode(&settings.BuildpackInfo)
	Expect(err).NotTo(HaveOccurred())
	Expect(file.Close()).To(Succeed())

	root, err := filepath.Abs("./..")
	Expect(err).ToNot(HaveOccurred())

	buildpackStore := occam.NewBuildpackStore()

	libpakBuildpackStore := occam.NewBuildpackStore().WithPackager(packagers.NewLibpak())

	settings.Buildpacks.DotnetExecute.Online, err = buildpackStore.Get.
		WithVersion("1.2.3").
		Execute(root)
	Expect(err).ToNot(HaveOccurred())

	settings.Buildpacks.ICU.Online, err = buildpackStore.Get.
		Execute(settings.Config.ICU)
	Expect(err).ToNot(HaveOccurred())

	settings.Buildpacks.DotnetCoreSDK.Online, err = buildpackStore.Get.
		Execute(settings.Config.DotnetCoreSDK)
	Expect(err).ToNot(HaveOccurred())

	settings.Buildpacks.DotnetCoreASPNetRuntime.Online, err = buildpackStore.Get.
		Execute(settings.Config.DotnetCoreASPNetRuntime)
	Expect(err).ToNot(HaveOccurred())

	settings.Buildpacks.DotnetPublish.Online, err = buildpackStore.Get.
		Execute(settings.Config.DotnetPublish)
	Expect(err).ToNot(HaveOccurred())

	settings.Buildpacks.NodeEngine.Online, err = buildpackStore.Get.
		Execute(settings.Config.NodeEngine)
	Expect(err).ToNot(HaveOccurred())

	settings.Buildpacks.Watchexec.Online, err = libpakBuildpackStore.Get.
		Execute(settings.Config.Watchexec)
	Expect(err).ToNot(HaveOccurred())

	settings.Buildpacks.Vsdbg.Online, err = buildpackStore.Get.
		Execute(settings.Config.Vsdbg)
	Expect(err).ToNot(HaveOccurred())

	// For backwards compatibility test
	settings.Buildpacks.DotnetCoreRuntime.Online, err = buildpackStore.Get.
		Execute(settings.Config.DotnetCoreRuntime)
	Expect(err).ToNot(HaveOccurred())

	settings.Buildpacks.DotnetCoreASPNet.Online, err = buildpackStore.Get.
		Execute(settings.Config.DotnetCoreASPNet)
	Expect(err).ToNot(HaveOccurred())

	SetDefaultEventuallyTimeout(10 * time.Second)

	buf := bytes.NewBuffer(nil)
	cmd := pexec.NewExecutable("pack")
	Expect(cmd.Execute(pexec.Execution{
		Args:   []string{"builder", "inspect", "--output", "json"},
		Stdout: buf,
		Stderr: buf,
	})).To(Succeed(), buf.String())

	Expect(json.Unmarshal(buf.Bytes(), &builder)).To(Succeed(), buf.String())

	suite := spec.New("Integration", spec.Report(report.Terminal{}), spec.Parallel())
	suite("BackwardsCompatibility", testBackwardsCompatibility)
	suite("FrameworkDependentDeployment", testFrameworkDependentDeployment)
	suite("FrameworkDependentExecutable", testFrameworkDependentExecutable)
	suite("Logging", testLogging)
	suite("NodeApp", testNodeApp)
	suite("SelfContainedExecutable", testSelfContainedExecutable)
	suite("SourceApp", testSourceApp)
	suite.Run(t)
}
