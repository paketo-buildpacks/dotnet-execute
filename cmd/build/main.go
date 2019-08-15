package main

import (
	"fmt"
	"github.com/buildpack/libbuildpack/buildpackplan"
	"os"

	"github.com/cloudfoundry/dotnet-core-conf-cnb/conf"

	"github.com/cloudfoundry/libcfbuildpack/build"
)

func main() {
	context, err := build.DefaultBuild()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to create a default build context: %s", err)
		os.Exit(101)
	}

	code, err := runBuild(context)
	if err != nil {
		context.Logger.Info(err.Error())
	}

	os.Exit(code)

}

func runBuild(context build.Build) (int, error) {
	contributor, willContribute, err := conf.NewContributor(context)
	if err != nil {
		return 102, err
	}

	if willContribute {
		if err := contributor.Contribute(); err != nil {
			return 103, err
		}
	}

	return context.Success(buildpackplan.Plan{})
}
