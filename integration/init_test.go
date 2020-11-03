package integration_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

var settings struct {
	BuildpackInfo struct {
		ID   string
		Name string
	}
	Config struct {
		ICU               string `json:"icu"`
		DotnetCoreRuntime string `json:"dotnet-core-runtime"`
		DotnetCoreSDK     string `json:"dotnet-core-sdk"`
		DotnetCoreASPNet  string `json:"dotnet-core-aspnet"`
		DotnetPublish     string `json:"dotnet-publish"`
		NodeEngine        string `json:"node-engine"`
	}
	Buildpacks struct {
		DotnetExecute struct {
			Online string
		}
		DotnetCoreRuntime struct {
			Online string
		}
		DotnetCoreSDK struct {
			Online string
		}
		DotnetCoreASPNet struct {
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
	}
}

func TestIntegration(t *testing.T) {
	Expect := NewWithT(t).Expect

	file, err := os.Open("../integration.json")
	Expect(err).NotTo(HaveOccurred())

	Expect(json.NewDecoder(file).Decode(&settings.Config)).To(Succeed())
	Expect(file.Close()).To(Succeed())

	file, err = os.Open("../buildpack.toml")
	Expect(err).NotTo(HaveOccurred())

	_, err = toml.DecodeReader(file, &settings.BuildpackInfo)
	Expect(err).NotTo(HaveOccurred())
	Expect(file.Close()).To(Succeed())

	root, err := filepath.Abs("./..")
	Expect(err).ToNot(HaveOccurred())

	buildpackStore := occam.NewBuildpackStore()

	settings.Buildpacks.DotnetExecute.Online, err = buildpackStore.Get.
		WithVersion("1.2.3").
		Execute(root)
	Expect(err).ToNot(HaveOccurred())

	settings.Buildpacks.ICU.Online, err = buildpackStore.Get.
		Execute(settings.Config.ICU)
	Expect(err).ToNot(HaveOccurred())

	settings.Buildpacks.DotnetCoreRuntime.Online, err = buildpackStore.Get.
		Execute(settings.Config.DotnetCoreRuntime)
	Expect(err).ToNot(HaveOccurred())

	settings.Buildpacks.DotnetCoreSDK.Online, err = buildpackStore.Get.
		Execute(settings.Config.DotnetCoreSDK)
	Expect(err).ToNot(HaveOccurred())

	settings.Buildpacks.DotnetCoreASPNet.Online, err = buildpackStore.Get.
		Execute(settings.Config.DotnetCoreASPNet)
	Expect(err).ToNot(HaveOccurred())

	settings.Buildpacks.DotnetPublish.Online, err = buildpackStore.Get.
		Execute(settings.Config.DotnetPublish)
	Expect(err).ToNot(HaveOccurred())

	settings.Buildpacks.NodeEngine.Online, err = buildpackStore.Get.
		Execute(settings.Config.NodeEngine)
	Expect(err).ToNot(HaveOccurred())

	SetDefaultEventuallyTimeout(10 * time.Second)

	suite := spec.New("Integration", spec.Report(report.Terminal{}), spec.Parallel())
	suite("SelfContainedExecutable", testSelfContainedExecutable)
	suite("FrameworkDependentDeployment", testFrameworkDependentDeployment)
	suite("FrameworkDependentExecutable", testFrameworkDependentExecutable)
	suite("FdeASPNet", testFdeASPNet)
	suite("FddASPNet", testFddASPNet)
	suite("SourceApp", testSourceApp)
	suite("NodeApp", testNodeApp)
	suite.Run(t)
}
