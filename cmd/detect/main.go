package main

import (
	"fmt"
	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/dotnet-core-conf-cnb/conf"
	"github.com/cloudfoundry/dotnet-core-conf-cnb/utils"
	"github.com/cloudfoundry/libcfbuildpack/detect"
	"os"
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

	code, err := runDetect(context)
	if err != nil {
		context.Logger.Info(err.Error())
	}

	os.Exit(code)
}

func runDetect(context detect.Detect) (int, error) {
	runtimeConfig, err := utils.NewRuntimeConfig(context.Application.Root)
	if err != nil {
		return context.Fail(), err
	}

	if !runtimeConfig.IsPresent() {
		context.Logger.Info(MissingRuntimeConfig)
		return context.Fail(), nil
	}

	return context.Pass(buildplan.Plan{
		Provides: []buildplan.Provided{{Name: conf.Layer}},
		Requires: []buildplan.Required{{
			Name: conf.Layer,
			Metadata: buildplan.Metadata{
				"build": true,
			}}},
	})

}
