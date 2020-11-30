package dotnetexecute

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gravityblast/go-jsmin"
)

type RuntimeConfig struct {
	Path       string
	Version    string
	AppName    string
	Executable bool
	UsesASPNet bool
}

type RuntimeConfigParser struct{}

func NewRuntimeConfigParser() RuntimeConfigParser {
	return RuntimeConfigParser{}
}

func (p RuntimeConfigParser) Parse(glob string) (RuntimeConfig, error) {
	files, err := filepath.Glob(glob)
	if err != nil {
		return RuntimeConfig{}, fmt.Errorf("failed to find *.runtimeconfig.json: %w: %q", err, glob)
	}

	if len(files) > 1 {
		return RuntimeConfig{}, fmt.Errorf("multiple *.runtimeconfig.json files present: %v", files)
	}

	if len(files) == 0 {
		return RuntimeConfig{}, fmt.Errorf("no *.runtimeconfig.json found: %w", os.ErrNotExist)
	}

	config := RuntimeConfig{
		Path: files[0],
	}

	var data struct {
		RuntimeOptions struct {
			Framework struct {
				Name    string `json:"name"`
				Version string `json:"version"`
			} `json:"framework"`
		} `json:"runtimeOptions"`
	}

	file, err := os.Open(config.Path)
	if err != nil {
		return RuntimeConfig{}, err
	}
	defer file.Close()

	buffer := bytes.NewBuffer(nil)
	err = jsmin.Min(file, buffer)
	if err != nil {
		return RuntimeConfig{}, err
	}

	err = json.NewDecoder(buffer).Decode(&data)
	if err != nil {
		return RuntimeConfig{}, err
	}

	config.Version = data.RuntimeOptions.Framework.Version
	if data.RuntimeOptions.Framework.Name == "Microsoft.NETCore.App" && config.Version == "" {
		config.Version = "*"
	}

	if data.RuntimeOptions.Framework.Name == "Microsoft.AspNetCore.App" ||
		data.RuntimeOptions.Framework.Name == "Microsoft.AspNetCore.All" {
		config.UsesASPNet = true
		if config.Version == "" {
			config.Version = "*"
		}
	}

	config.AppName = strings.TrimSuffix(filepath.Base(file.Name()), ".runtimeconfig.json")

	info, err := os.Stat(strings.TrimSuffix(file.Name(), ".runtimeconfig.json"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return RuntimeConfig{}, err
	}

	if info != nil && info.Mode()&0111 != 0 {
		config.Executable = true
	}

	return config, nil
}
