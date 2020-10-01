package dotnetcoreconf_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitDotnetCoreConf(t *testing.T) {
	suite := spec.New("dotnet-core-conf", spec.Report(report.Terminal{}), spec.Parallel())
	// suite("Build", testBuild)
	suite("BuildpackYMLParser", testBuildpackYMLParser)
	suite("Detect", testDetect)
	suite.Run(t)
}
