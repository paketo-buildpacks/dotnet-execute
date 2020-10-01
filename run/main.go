package main

import (
	dotnetcoreconf "github.com/paketo-buildpacks/dotnet-core-conf"
	"github.com/paketo-buildpacks/packit"
)

func main() {
	// logger := scribe.NewLogger(os.Stdout)
	packit.Run(dotnetcoreconf.Detect(), dotnetcoreconf.Build())
}
