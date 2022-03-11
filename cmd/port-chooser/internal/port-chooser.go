package internal

import (
	"fmt"
	"os"
	"strconv"
)

const (
	// AspNetCoreUrls is the canonical way to set the port for ASP.NET Core
	// 6.0: https://docs.microsoft.com/en-us/aspnet/core/fundamentals/host/web-host?view=aspnetcore-6.0#server-urls
	// 5.0: https://docs.microsoft.com/en-us/aspnet/core/fundamentals/host/web-host?view=aspnetcore-5.0#server-urls-1
	// 3.1: https://docs.microsoft.com/en-us/aspnet/core/fundamentals/host/web-host?view=aspnetcore-3.1#server-urls-2
	AspNetCoreUrls = "ASPNETCORE_URLS"
)

func ChoosePort() (map[string]string, error) {
	if _, hasUrl := os.LookupEnv(AspNetCoreUrls); hasUrl {
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
		AspNetCoreUrls: url,
	}
	return envVars, nil
}
