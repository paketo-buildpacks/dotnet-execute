package dotnetexecute_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitDotnetExecute(t *testing.T) {
	suite := spec.New("dotnet-execute", spec.Report(report.Terminal{}), spec.Parallel())
	suite("BuildpackYMLParser", testBuildpackYMLParser)
	suite("Detect", testDetect)
	suite("Build", testBuild)
	suite.Run(t)
}
