package dotnetexecute

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ProjectFileParser struct{}

func NewProjectFileParser() ProjectFileParser {
	return ProjectFileParser{}
}

func (p ProjectFileParser) FindProjectFile(path string) (string, error) {
	projectFiles, err := filepath.Glob(filepath.Join(path, "*.csproj"))
	if err != nil {
		return "", err
	}

	fsProjFiles, err := filepath.Glob(filepath.Join(path, "*.fsproj"))
	if err != nil {
		return "", err
	}
	projectFiles = append(projectFiles, fsProjFiles...)

	vbProjFiles, err := filepath.Glob(filepath.Join(path, "*.vbproj"))
	if err != nil {
		return "", err
	}
	projectFiles = append(projectFiles, vbProjFiles...)

	if len(projectFiles) > 0 {
		return projectFiles[0], nil
	}

	return "", nil
}

func (p ProjectFileParser) NodeIsRequired(path string) (bool, error) {
	needsNode, err := findInFile("node ", path)
	if err != nil {
		return false, err
	}

	needsNPM, err := findInFile("npm ", path)
	if err != nil {
		return false, err
	}

	return needsNode || needsNPM, nil
}

func (p ProjectFileParser) NPMIsRequired(path string) (bool, error) {
	return findInFile("npm ", path)
}

func findInFile(str, path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("failed to open %s: %w", path, err)
	}
	defer func() {
		_ = file.Close()
	}()

	var project struct {
		Targets []struct {
			Execs []struct {
				Command string `xml:",attr"`
			} `xml:"Exec"`
		} `xml:"Target"`
	}

	err = xml.NewDecoder(file).Decode(&project)
	if err != nil {
		return false, fmt.Errorf("failed to decode %s: %w", path, err)
	}

	for _, target := range project.Targets {
		for _, exec := range target.Execs {
			if strings.HasPrefix(exec.Command, str) {
				return true, nil
			}
		}
	}

	return false, nil
}
