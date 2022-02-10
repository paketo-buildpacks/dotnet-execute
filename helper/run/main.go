package main

import (
	"github.com/paketo-buildpacks/dotnet-execute/helper"
	"github.com/paketo-buildpacks/libpak/sherpa"
)

func main() {
	sherpa.Execute(func() error {
		return sherpa.Helpers(map[string]sherpa.ExecD{
			`helper`: helper.Helper{},
		})
	})
}
