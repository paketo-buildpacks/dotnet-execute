package main

import (
	"fmt"
	"github.com/paketo-buildpacks/dotnet-execute/cmd/port-chooser/internal"
	"os"
	"strconv"
)

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
