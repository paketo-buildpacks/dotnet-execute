package conf

import (
	"fmt"

	"github.com/cloudfoundry/dotnet-core-conf-cnb/utils"
	"github.com/cloudfoundry/libcfbuildpack/build"
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
	runtimeConfig, err := utils.NewRuntimeConfig(c.context.Application.Root)
	if err != nil {
		return err
	}

	hasFDE, err := runtimeConfig.HasFDE()
	if err != nil {
		return err
	}

	startCmdPrefix := fmt.Sprintf("dotnet %s.dll", runtimeConfig.BinaryName)
	if hasFDE {
		startCmdPrefix = fmt.Sprintf("./%s", runtimeConfig.BinaryName)
	}

	args := startCmdPrefix
	if !runtimeConfig.HasRuntimeDependency() {
		args = fmt.Sprintf("%s --urls http://0.0.0.0:${PORT}", startCmdPrefix)
	}

	startCmd := fmt.Sprintf("cd %s && %s", c.context.Application.Root, args)

	if c.context.Build.Stack == "io.buildpacks.stacks.bionic" {
		err := c.context.Layers.Layer(Layer).Contribute(c.context.Buildpack, func(layer layers.Layer) error {
			if err := layer.OverrideLaunchEnv("DOTNET_SYSTEM_GLOBALIZATION_INVARIANT", "true"); err != nil {
				return err
			}
			return err
		}, []layers.Flag{layers.Launch}...)

		if err != nil {
			return err
		}
	}

	return c.context.Layers.WriteApplicationMetadata(layers.Metadata{Processes: []layers.Process{{"web", startCmd, false}}})
}
