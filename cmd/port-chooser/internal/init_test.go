package internal_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitDotnetExecute(t *testing.T) {
	suite := spec.New("dotnet-execute", spec.Report(report.Terminal{}), spec.Sequential())
	suite("portChooser", testPortChooser)
	suite.Run(t)
}
