package main

import (
	"os"

	dotnetexecute "github.com/paketo-buildpacks/dotnet-execute"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/parsers"
	"github.com/paketo-buildpacks/packit/scribe"
)

func main() {
	logger := scribe.NewLogger(os.Stdout)
	buildpackYMLParser := dotnetexecute.NewBuildpackYMLParser(logger)
	configParser := dotnetexecute.NewRuntimeConfigParser()
	projectParser := dotnetexecute.NewProjectFileParser()
	projectPathParser := parsers.NewProjectPathParser()

	packit.Run(
		dotnetexecute.Detect(
			buildpackYMLParser,
			configParser,
			projectParser,
			projectPathParser,
		),
		dotnetexecute.Build(logger),
	)
}
