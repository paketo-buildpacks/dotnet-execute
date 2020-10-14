package main

import (
	"os"

	dotnetexecute "github.com/paketo-buildpacks/dotnet-execute"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/scribe"
)

func main() {
	logger := scribe.NewLogger(os.Stdout)
	buildpackYMLParser := dotnetexecute.NewBuildpackYMLParser()
	packit.Run(
		dotnetexecute.Detect(buildpackYMLParser),
		dotnetexecute.Build(logger),
	)
}
