package dotnetexecute

import (
	"os"

	"github.com/paketo-buildpacks/packit/scribe"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	ProjectPath string `yaml:"project-path"`
}

type BuildpackYMLParser struct {
	logger scribe.Logger
}

func NewBuildpackYMLParser(logger scribe.Logger) BuildpackYMLParser {
	return BuildpackYMLParser{
		logger: logger,
	}
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

	if config.ProjectPath != "" {
		p.logger.Subprocess("WARNING: Setting the project path through buildpack.yml will be deprecated soon in Dotnet Execute Buildpack v1.0.0")
		p.logger.Subprocess("Please specify the project path through the $BP_DOTNET_PROJECT_PATH environment variable instead. See README.md or the documentation on paketo.io for more information.")
	}

	return config.ProjectPath, nil
}
