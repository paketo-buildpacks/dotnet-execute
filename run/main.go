package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Netflix/go-env"
	dotnetexecute "github.com/paketo-buildpacks/dotnet-execute"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/sbom"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

type Generator struct{}

func (f Generator) Generate(path string) (sbom.SBOM, error) {
	return sbom.Generate(path)
}

func main() {
	var config dotnetexecute.Configuration
	_, err := env.UnmarshalFromEnviron(&config)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to parse build configuration: %w", err))
	}

	logger := scribe.NewEmitter(os.Stdout).WithLevel(config.LogLevel)
	buildpackYMLParser := dotnetexecute.NewBuildpackYMLParser()
	configParser := dotnetexecute.NewRuntimeConfigParser()
	projectParser := dotnetexecute.NewProjectFileParser()

	packit.Run(
		dotnetexecute.Detect(
			config,
			logger,
			buildpackYMLParser,
			configParser,
			projectParser,
		),
		dotnetexecute.Build(
			config,
			buildpackYMLParser,
			configParser,
			Generator{},
			logger,
			chronos.DefaultClock,
		),
	)
}
