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
	configParser := dotnetexecute.NewRuntimeConfigParser()
	projectParser := dotnetexecute.NewProjectFileParser()

	packit.Run(
		dotnetexecute.Detect(
			buildpackYMLParser,
			configParser,
			projectParser,
		),
		dotnetexecute.Build(logger),
	)
}
