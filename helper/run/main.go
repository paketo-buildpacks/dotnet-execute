package main

import (
	"fmt"
	"github.com/paketo-buildpacks/dotnet-execute/helper"
	"os"
	"strconv"
)

func main() {
	execdWriter := os.NewFile(3, "/dev/fd/3")

	envVars, err := helper.Execute()
	if err != nil {
		return
	}

	for k, v := range envVars {
		if _, err := fmt.Fprintf(execdWriter, "%s=%s\n", k, strconv.Quote(v)); err != nil {
			return
		}
	}
}
