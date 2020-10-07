package main

import (
	"os"

	dotnetcoreconf "github.com/paketo-buildpacks/dotnet-core-conf"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/scribe"
)

func main() {
	logger := scribe.NewLogger(os.Stdout)
	buildpackYMLParser := dotnetcoreconf.NewBuildpackYMLParser()
	packit.Run(
		dotnetcoreconf.Detect(buildpackYMLParser),
		dotnetcoreconf.Build(buildpackYMLParser, logger),
	)
}
