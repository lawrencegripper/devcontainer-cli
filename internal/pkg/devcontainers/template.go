package devcontainers

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/stuartleeks/devcontainer-cli/internal/pkg/config"
	"github.com/stuartleeks/devcontainer-cli/internal/pkg/errors"
)

// DevcontainerTemplate holds info on templates for list/add etc
type DevcontainerTemplate struct {
	Name string
	// Path is the path including the .devcontainer folder
	Path string
}

// GetTemplateByName returns the template with the specified name or nil if not found
func GetTemplateByName(name string) (*DevcontainerTemplate, error) {
	// TODO - could possibly make this quicker by searching using the name rather than listing all and filtering
	templates, err := GetTemplates()
	if err != nil {
		return nil, err
	}
	for _, template := range templates {
		if template.Name == name {
			return &template, nil
		}
	}
	return nil, nil
}

// GetTemplates returns a list of discovered templates
func GetTemplates() ([]DevcontainerTemplate, error) {
	templates := []DevcontainerTemplate{}
	templateNames := map[string]bool{}

	folders := config.GetTemplateFolders()
	if len(folders) == 0 {
		return []DevcontainerTemplate{}, &errors.StatusError{Message: "No template folders configured - see https://github.com/stuartleeks/devcontainer-cli/#working-with-devcontainer-templates"}
	}
	for _, folder := range folders {
		folder := os.ExpandEnv(folder)
		newTemplates, err := getTemplatesFromFolder(folder)
		if err != nil {
			return []DevcontainerTemplate{}, err
		}
		for _, template := range newTemplates {
			if !templateNames[template.Name] {
				templateNames[template.Name] = true
				templates = append(templates, template)
			}
		}
	}
	sort.Slice(templates, func(i int, j int) bool { return templates[i].Name < templates[j].Name })
	return templates, nil
}

func getTemplatesFromFolder(folder string) ([]DevcontainerTemplate, error) {
	isDevcontainerFolder := func(parentPath string, fi os.FileInfo) bool {
		if !fi.IsDir() {
			return false
		}
		devcontainerJsonPath := filepath.Join(parentPath, fi.Name(), ".devcontainer/devcontainer.json")
		devContainerJsonInfo, err := os.Stat(devcontainerJsonPath)
		return err == nil && !devContainerJsonInfo.IsDir()
	}
	c, err := ioutil.ReadDir(folder)

	if err != nil {
		return []DevcontainerTemplate{}, fmt.Errorf("Error reading devcontainer definitions: %s\n", err)
	}

	templates := []DevcontainerTemplate{}
	for _, entry := range c {
		if isDevcontainerFolder(folder, entry) {
			template := DevcontainerTemplate{
				Name: entry.Name(),
				Path: filepath.Join(folder, entry.Name(), ".devcontainer"),
			}
			templates = append(templates, template)
		}
	}
	return templates, nil
}

func GetDefaultDevcontainerNameForFolder(folderPath string) (string, error) {

	absPath, err := filepath.Abs(folderPath)
	if err != nil {
		return "", err
	}

	_, folderName := filepath.Split(absPath)
	return folderName, nil
}

func SetDevcontainerName(devContainerJsonPath string, name string) error {
	// This doesn't use `json` as devcontainer.json permits comments (and the default templates include them!)

	buf, err := ioutil.ReadFile(devContainerJsonPath)
	if err != nil {
		return fmt.Errorf("error reading file %q: %s", devContainerJsonPath, err)
	}

	r := regexp.MustCompile("(\"name\"\\s*:\\s*\")[^\"]*(\")")
	replacement := []byte("${1}" + name + "${2}")
	buf = r.ReplaceAll(buf, replacement)

	if err = ioutil.WriteFile(devContainerJsonPath, buf, 0777); err != nil {
		return fmt.Errorf("error writing file %q: %s", devContainerJsonPath, err)
	}

	return nil
}

// "remoteUser": "vscode"

func GetDevContainerUserName(devContainerJsonPath string) (string, error) {
	buf, err := ioutil.ReadFile(devContainerJsonPath)
	if err != nil {
		return "", fmt.Errorf("error reading file %q: %s", devContainerJsonPath, err)
	}

	r := regexp.MustCompile("\n[^/]*\"remoteUser\"\\s*:\\s*\"([^\"]*)\"")
	match := r.FindStringSubmatch(string(buf))

	if len(match) <= 0 {
		return "", nil
	}
	return match[1], nil
}
