package dotnetexecute

import (
	"encoding/json"
	"os"
)

type RuntimeConfig struct {
	RuntimeVersion string
}

type RuntimeConfigParser struct{}

func NewRuntimeConfigParser() RuntimeConfigParser {
	return RuntimeConfigParser{}
}

func (p RuntimeConfigParser) Parse(path string) (RuntimeConfig, error) {
	var data struct {
		RuntimeOptions struct {
			Framework struct {
				Name    string `json:"name"`
				Version string `json:"version"`
			} `json:"framework"`
		} `json:"runtimeOptions"`
	}

	file, err := os.Open(path)
	if err != nil {
		return RuntimeConfig{}, err
	}

	err = json.NewDecoder(file).Decode(&data)
	if err != nil {
		return RuntimeConfig{}, err
	}

	// If there is an expected framework set with no version, default the version to *
	if expectedFrameworkName(data.RuntimeOptions.Framework.Name) && data.RuntimeOptions.Framework.Version == "" {
		data.RuntimeOptions.Framework.Version = "*"
	}

	var config RuntimeConfig
	config.RuntimeVersion = data.RuntimeOptions.Framework.Version

	return config, nil
}

// This check if the name of the framework is one that we expect
func expectedFrameworkName(name string) bool {
	return name == "Microsoft.NETCore.App"
}
