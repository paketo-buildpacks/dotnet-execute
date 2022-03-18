package main

import (
	"os"

	dotnetexecute "github.com/paketo-buildpacks/dotnet-execute"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

func main() {
	logger := scribe.NewEmitter(os.Stdout)
	buildpackYMLParser := dotnetexecute.NewBuildpackYMLParser()
	configParser := dotnetexecute.NewRuntimeConfigParser()
	projectParser := dotnetexecute.NewProjectFileParser()

	packit.Run(
		dotnetexecute.Detect(
			buildpackYMLParser,
			configParser,
			projectParser,
		),
		dotnetexecute.Build(buildpackYMLParser, configParser, logger),
	)
}
