package main

import (
	"os"

	dotnetcoreexecute "github.com/paketo-buildpacks/dotnet-core-execute"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/scribe"
)

func main() {
	logger := scribe.NewLogger(os.Stdout)
	buildpackYMLParser := dotnetcoreexecute.NewBuildpackYMLParser()
	packit.Run(
		dotnetcoreexecute.Detect(buildpackYMLParser),
		dotnetcoreexecute.Build(buildpackYMLParser, logger),
	)
}
