package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gravityblast/go-jsmin"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
)

type ConfigJSON struct {
	RuntimeOptions struct {
		Framework struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"framework"`
		ApplyPatches *bool `json:"applyPatches"`
	} `json:"runtimeOptions"`
}

func CreateRuntimeConfig(appRoot string) (ConfigJSON, error){
	path, err := runtimeConfigPath(appRoot)
	if err != nil {
		return ConfigJSON{}, err
	}

	runtimeJSON := ConfigJSON{}

	if path != "" {
		runtimeJSON, err = parseRuntimeConfig(path)
		if err != nil {
			return ConfigJSON{}, err
		}
	}

	return runtimeJSON, nil
}

func runtimeConfigPath(appRoot string) (string, error) {
	if configFiles, err := filepath.Glob(filepath.Join(appRoot, "*.runtimeconfig.json")); err != nil {
		return "", err
	} else if len(configFiles) == 1 {
		return configFiles[0], nil
	} else if len(configFiles) > 1 {
		return "", fmt.Errorf("multiple *.runtimeconfig.json files present")
	}
	return "", nil
}

func parseRuntimeConfig(runtimeConfigPath string) (ConfigJSON, error) {
	obj := ConfigJSON{}

	buf, err := sanitizeJsonConfig(runtimeConfigPath)
	if err != nil {
		return obj, err
	}

	if err := json.Unmarshal(buf, &obj); err != nil {
		return obj, errors.Wrap(err, "unable to parse runtime config")
	}

	return obj, nil
}

func sanitizeJsonConfig(runtimeConfigPath string) ([]byte, error) {
	input, err := os.Open(runtimeConfigPath)
	if err != nil {
		return nil, err
	}
	defer input.Close()

	output := &bytes.Buffer{}

	if err := jsmin.Min(input, output); err != nil {
		return nil, err
	}

	return output.Bytes(), nil
}