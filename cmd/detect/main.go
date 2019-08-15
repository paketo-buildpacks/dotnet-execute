package main

import (
	"fmt"
	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/dotnet-core-conf-cnb/conf"
	"os"
	"regexp"

	"github.com/cloudfoundry/libcfbuildpack/helper"

	"github.com/cloudfoundry/libcfbuildpack/detect"
)

const (
	MissingRuntimeConfig  = "*.runtimeconfig.json file not found"
	TooManyRuntimeConfigs = "multiple *.runtimeconfig.json files present"
)

func main() {
	context, err := detect.DefaultDetect()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to create a default detection context: %s", err)
		os.Exit(100)
	}

	// TODO
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
	runtimeConfigRe := regexp.MustCompile(`\.(runtimeconfig\.json)$`)
	runtimeConfigMatches, err := helper.FindFiles(context.Application.Root, runtimeConfigRe)
	if err != nil {
		return context.Fail(), err
	}

	if len(runtimeConfigMatches) < 1 {
		context.Logger.Info(MissingRuntimeConfig)
		return context.Fail(), nil
	} else if len(runtimeConfigMatches) > 1 {
		return context.Fail(), fmt.Errorf(TooManyRuntimeConfigs)
	} else {
		return context.Pass(buildplan.Plan{
			Provides: []buildplan.Provided{{Name: conf.Layer}},
			Requires: []buildplan.Required{{
				Name: conf.Layer,
				Metadata: buildplan.Metadata{
					"build": true,
				}}},
		})
	}
}
