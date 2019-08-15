package conf

import (
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
)

const Layer = "dotnet-core-conf"

type Contributor struct {
	context build.Build
}

func NewContributor(context build.Build) (Contributor, bool, error) {

	_, wantLayer, err := context.Plans.GetShallowMerged(Layer)
	if err != nil {
		return Contributor{}, false, nil
	}

	if !wantLayer {
		return Contributor{}, false, nil
	}

	return Contributor{context: context}, true, nil
}

func (c Contributor) Contribute() error {
	runtimeConfigRe := regexp.MustCompile(`\.(runtimeconfig\.json)$`)
	runtimeConfigMatches, err := helper.FindFiles(c.context.Application.Root, runtimeConfigRe)
	if err != nil {
		return err
	}
	runtimeConfigFile := filepath.Base(runtimeConfigMatches[0])
	executableFile := runtimeConfigRe.ReplaceAllString(runtimeConfigFile, "")

	startCmd := fmt.Sprintf("cd %s && ./%s --server.urls http://0.0.0.0:${PORT}", c.context.Application.Root, executableFile)

	return c.context.Layers.WriteApplicationMetadata(layers.Metadata{Processes: []layers.Process{{"web", startCmd}}})
}
