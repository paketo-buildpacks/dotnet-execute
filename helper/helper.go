package helper

import (
	"fmt"
	"os"
	"strconv"
)

const (
	ENV_VAR = "ASPNETCORE_URLS"
)

type Helper struct{}

func (h Helper) Execute() (map[string]string, error) {
	if _, hasUrl := os.LookupEnv(ENV_VAR); hasUrl {
		return map[string]string{}, nil
	}

	portForDotNet := 8080

	if port, hasPort := os.LookupEnv("PORT"); hasPort {
		if port, err := strconv.Atoi(port); err == nil {
			portForDotNet = port
		}
	}

	url := fmt.Sprintf("http://0.0.0.0:%d", portForDotNet)

	fmt.Printf("Setting ASPNETCORE_URLS=%s\n", url)

	envVars := map[string]string{
		ENV_VAR: url,
	}
	return envVars, nil
}
