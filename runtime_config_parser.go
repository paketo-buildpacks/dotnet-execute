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
	Path           string
	RuntimeVersion string
	ASPNETVersion  string
	AppName        string
	Executable     bool
}

type framework struct {
	Name    string `json:"name"`
	Version string `json:"version"`
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
			Framework  framework   `json:"framework"`
			Frameworks []framework `json:"frameworks"`
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

	switch data.RuntimeOptions.Framework.Name {
	case "Microsoft.NETCore.App":
		config.RuntimeVersion = versionOrWildcard(data.RuntimeOptions.Framework.Version)
	case "Microsoft.AspNetCore.App":
		config.ASPNETVersion = versionOrWildcard(data.RuntimeOptions.Framework.Version)
		config.RuntimeVersion = config.ASPNETVersion
	default:
		config.RuntimeVersion = ""
		config.ASPNETVersion = ""
	}

	for _, f := range data.RuntimeOptions.Frameworks {
		switch f.Name {
		case "Microsoft.NETCore.App":
			if config.RuntimeVersion != "" {
				return RuntimeConfig{}, fmt.Errorf("malformed runtimeconfig.json: multiple '%s' frameworks specified", f.Name)
			}
			config.RuntimeVersion = versionOrWildcard(f.Version)
		case "Microsoft.AspNetCore.App":
			if config.ASPNETVersion != "" {
				return RuntimeConfig{}, fmt.Errorf("malformed runtimeconfig.json: multiple '%s' frameworks specified", f.Name)
			}
			config.ASPNETVersion = versionOrWildcard(f.Version)
		default:
			continue
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

func versionOrWildcard(version string) string {
	if version == "" {
		return "*"
	}
	return version
}
