package dotnetcoreexecute

import (
	"os"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	ProjectPath string `yaml:"project-path"`
}

type BuildpackYMLParser struct{}

func NewBuildpackYMLParser() BuildpackYMLParser {
	return BuildpackYMLParser{}
}

func (p BuildpackYMLParser) Parse(path string) (Config, error) {
	var buildpack struct {
		DotnetBuild Config `yaml:"dotnet-build"`
	}

	file, err := os.Open(path)
	if err != nil && !os.IsNotExist(err) {
		return Config{}, err
	}
	defer file.Close()

	if !os.IsNotExist(err) {
		err = yaml.NewDecoder(file).Decode(&buildpack)
		if err != nil {
			return Config{}, err
		}
	}

	return buildpack.DotnetBuild, nil
}

func (p BuildpackYMLParser) ParseProjectPath(path string) (string, error) {
	config, err := p.Parse(path)
	if err != nil {
		return "", err
	}

	return config.ProjectPath, nil
}
