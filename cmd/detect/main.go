package main

import (
	"fmt"
	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/dotnet-core-conf-cnb/conf"
	"github.com/cloudfoundry/dotnet-core-conf-cnb/utils"
	"github.com/cloudfoundry/libcfbuildpack/detect"
	"os"
	"path/filepath"
)

const (
	MissingRuntimeConfig  = "*.runtimeconfig.json file not found"
	NotSingleProjFile = "expecting only a single *.csproj file in the app directory"
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

	projFiles, err := filepath.Glob(filepath.Join(context.Application.Root, "*.csproj"))
	if err != nil {
		return context.Fail(), err
	}

	if  !runtimeConfig.IsPresent() && len(projFiles) != 1 {
		context.Logger.Info("%s and %s", MissingRuntimeConfig, NotSingleProjFile)
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
