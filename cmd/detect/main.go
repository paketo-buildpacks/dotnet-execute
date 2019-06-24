package main

import (
	"fmt"
	"os"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/detect"
)

func main() {
	context, err := detect.DefaultDetect()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to create a default detection context: %s", err)
		os.Exit(100)
	}

	// Decide if the following is needed for this CNB. This is needed if your CNB depends on the results of previous CNB contributions to the buildplan.
	// Otherwise, don't use it as it negates parallelization.
	//if err := context.BuildPlan.Init(); err != nil {
	//	_, _ = fmt.Fprintf(os.Stderr, "Failed to initialize Build Plan: %s\n", err)
	//	os.Exit(101)
	//}

	code, err := runDetect(context)
	if err != nil {
		context.Logger.Info(err.Error())
	}

	os.Exit(code)
}

func runDetect(context detect.Detect) (int, error) {
	return context.Pass(buildplan.BuildPlan{}) // TODO implementation
}
