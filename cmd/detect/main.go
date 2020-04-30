package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/paketo-buildpacks/dotnet-core-conf/conf"
	"github.com/paketo-buildpacks/dotnet-core-conf/utils"
)

const (
	MissingRuntimeConfig  = "*.runtimeconfig.json file not found"
	NotSingleProjFile     = "expecting only a single *.csproj file in the app directory"
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
	plan := buildplan.Plan{
		Provides: []buildplan.Provided{{Name: conf.Layer}},
		Requires: []buildplan.Required{{
			Name: conf.Layer,
			Metadata: buildplan.Metadata{
				"build": true,
			}}},
	}

	runtimeConfig, err := utils.NewRuntimeConfig(context.Application.Root)
	if err != nil {
		return context.Fail(), err
	}

	sourceAppRoot, err := utils.GetAppRoot(context.Application.Root)
	if err != nil {
		return context.Fail(), err
	}

	projFiles, err := filepath.Glob(filepath.Join(sourceAppRoot, "*.*sproj"))
	if err != nil {
		return context.Fail(), err
	}

	if !runtimeConfig.IsPresent() && len(projFiles) != 1 {
		context.Logger.Info("%s and %s", MissingRuntimeConfig, NotSingleProjFile)
		return context.Fail(), nil
	}

	if context.Stack == "io.buildpacks.stacks.bionic" {
		plan.Requires = append(plan.Requires, buildplan.Required{
			Name:     "icu",
			Metadata: buildplan.Metadata{"launch": true},
		})
	}

	return context.Pass(plan)
}
