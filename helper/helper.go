package helper

import (
	"fmt"
	"os"
	"strconv"
)

const (
	EnvVar = "ASPNETCORE_URLS"
)

type Helper struct{}

func (h Helper) Execute() (map[string]string, error) {
	if _, hasUrl := os.LookupEnv(EnvVar); hasUrl {
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
		EnvVar: url,
	}
	return envVars, nil
}
