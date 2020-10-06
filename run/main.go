package main

import (
	dotnetcoreconf "github.com/paketo-buildpacks/dotnet-core-conf"
	"github.com/paketo-buildpacks/packit"
)

func main() {
	// logger := scribe.NewLogger(os.Stdout)
	buildpackYMLParser := dotnetcoreconf.NewBuildpackYMLParser()
	packit.Run(
		dotnetcoreconf.Detect(buildpackYMLParser),
		dotnetcoreconf.Build(buildpackYMLParser),
	)
}
