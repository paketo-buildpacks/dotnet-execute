package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/paketo-buildpacks/dotnet-execute/cmd/port-chooser/internal"
)

// main will invoke the port chooser, and write all provided environment variables to FD 3.
// See https://github.com/buildpacks/rfcs/blob/main/text/0093-remove-shell-processes.md.
func main() {
	execdWriter := os.NewFile(3, "/dev/fd/3")

	envVars, err := internal.ChoosePort()
	if err != nil {
		return
	}

	for k, v := range envVars {
		if _, err := fmt.Fprintf(execdWriter, "%s=%s\n", k, strconv.Quote(v)); err != nil {
			return
		}
	}
}
