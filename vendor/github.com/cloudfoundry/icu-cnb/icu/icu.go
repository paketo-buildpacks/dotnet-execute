package icu

import (
	"os"
	"path/filepath"

	"github.com/buildpack/libbuildpack/buildpackplan"
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
)

type Contributor struct {
	context  build.Build
	icuLayer layers.DependencyLayer
	plan     buildpackplan.Plan
}

const Dependency = "icu"

func NewContributor(context build.Build) (Contributor, bool, error) {
	plan, wantDependency, err := context.Plans.GetShallowMerged(Dependency)
	if err != nil {
		return Contributor{}, false, err
	}

	if !wantDependency {
		return Contributor{}, false, nil
	}

	dep, err := context.Buildpack.RuntimeDependency(Dependency, plan.Version, context.Stack)
	if err != nil {
		return Contributor{}, false, err
	}

	return Contributor{
		context:  context,
		icuLayer: context.Layers.DependencyLayer(dep),
		plan:     plan,
	}, true, nil
}

func (c Contributor) Contribute() error {
	return c.icuLayer.Contribute(func(artifact string, layer layers.DependencyLayer) error {
		layer.Logger.Body("Expanding to %s", layer.Root)

		if err := helper.ExtractTarGz(artifact, layer.Root, 2); err != nil {
			return err
		}

		//TODO: This structure will most likely change when we build the dependency ourselves
		if err := helper.CopyDirectory(filepath.Join(layer.Root, "usr", "local", "bin"), filepath.Join(layer.Root, "bin")); err != nil {
			return err
		}

		if err := helper.CopyDirectory(filepath.Join(layer.Root, "usr", "local", "sbin"), filepath.Join(layer.Root, "bin")); err != nil {
			return err
		}

		if err := helper.CopyDirectory(filepath.Join(layer.Root, "usr", "local", "include"), filepath.Join(layer.Root, "include")); err != nil {
			return err
		}

		if err := helper.CopyDirectory(filepath.Join(layer.Root, "usr", "local", "lib"), filepath.Join(layer.Root, "lib")); err != nil {
			return err
		}

		if err := os.RemoveAll(filepath.Join(layer.Root, "usr")); err != nil {
			return err
		}

		return nil
	}, getFlags(c.plan.Metadata)...)
}

func getFlags(metadata buildpackplan.Metadata) []layers.Flag {
	flagsArray := []layers.Flag{}
	flagValueMap := map[string]layers.Flag{"build": layers.Build, "launch": layers.Launch, "cache": layers.Cache}
	for _, flagName := range []string{"build", "launch", "cache"} {
		flagPresent, _ := metadata[flagName].(bool)
		if flagPresent {
			flagsArray = append(flagsArray, flagValueMap[flagName])
		}
	}
	return flagsArray
}
