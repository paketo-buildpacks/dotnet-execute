package dotnetcoreexecute_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitDotnetCoreExecute(t *testing.T) {
	suite := spec.New("dotnet-core-execute", spec.Report(report.Terminal{}), spec.Parallel())
	suite("BuildpackYMLParser", testBuildpackYMLParser)
	suite("Detect", testDetect)
	suite("Build", testBuild)
	suite.Run(t)
}
