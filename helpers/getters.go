package helpers

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/padok-team/guacamole/data"

	"github.com/spf13/viper"
)

func GetModules() (map[string]data.TerraformModule, error) {
	codebasePath := viper.GetString("codebase-path")
	modules := make(map[string]data.TerraformModule)
	//Get all subdirectories in root path
	err := filepath.Walk(codebasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to get subdirectories: %w", err)
		}
		// Check if the path is a file and its name matches "*.tf"
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".tf") {
			// exclude the files which are in the .terragrunt-cache or .terraform directory
			if !regexp.MustCompile(`\.terragrunt-cache|\.terraform`).MatchString(path) {
				// Get all whitelisting comments
				whitelistCommentOfFile, _ := GetWhitelistingComments(path)
				// Check if the module is already in the list
				alreadyInList := false
				for _, m := range modules {
					if m.FullPath == filepath.Dir(path) {
						alreadyInList = true
					}
				}
				// If not in list create the module
				if !alreadyInList {
					modules[filepath.Dir(path)], _ = LoadModule(filepath.Dir(path))
				}
				// Create a temporary object with all resources in order
				resourcesInFile := make(map[string]data.TerraformCodeBlock)
				for index, resource := range modules[filepath.Dir(path)].Resources {
					if path == resource.FilePath {
						resourcesInFile[index] = resource
					}
				}
				// Order the list of resources from top to bottom via a key array
				keys := make([]string, 0, len(resourcesInFile))
				for k := range resourcesInFile {
					keys = append(keys, k)
				}

				sort.SliceStable(keys, func(i, j int) bool {
					return resourcesInFile[keys[i]].Pos < resourcesInFile[keys[j]].Pos
				})

				// Associate the whitelisting comments to a resource (Resource, Data, Variable or Output)
				for _, whitelistingComment := range whitelistCommentOfFile {
					previousPos := 0
					for _, i := range keys {
						// Find closest code block within the file
						if previousPos < whitelistingComment.LineNumber && whitelistingComment.LineNumber < resourcesInFile[i].Pos {
							r := modules[filepath.Dir(path)].Resources[i]
							r.WhitelistComments = append(r.WhitelistComments, whitelistingComment)
							modules[filepath.Dir(path)].Resources[i] = r
							break
						}
						previousPos = resourcesInFile[i].Pos
					}
				}
			}
		}
		return nil
	})
	return modules, err
}

func GetLayers() ([]*data.Layer, error) {
	codebasePath := viper.GetString("codebase-path")
	codebaseAbsPath, err := filepath.Abs(codebasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get: %w", err)
	}

	layers := []*data.Layer{}

	err = filepath.Walk(codebasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to get layer subdirectory: %w", err)
		}

		// Check if the current path is a file and its name matches "terragrunt.hcl"
		if !info.IsDir() && info.Name() == "terragrunt.hcl" {
			// exclude the files which are in the .terragrunt-cache or .terraform directory
			if !regexp.MustCompile(`\.terragrunt-cache|\.terraform`).MatchString(path) {
				// TODO: start from the codebase path instead of the relative path
				absPath, err := filepath.Abs(path)
				if err != nil {
					return fmt.Errorf("failed to get absolute path: %w", err)
				}
				fullPath := filepath.Dir(absPath)
				// Get abs path from codebase path
				name := strings.Split(absPath, codebaseAbsPath+"/")[1]
				name = strings.ReplaceAll(name, "/terragrunt.hcl", "")
				layers = append(layers, &data.Layer{Name: name, FullPath: fullPath})
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return layers, nil
}

func LoadModule(path string) (data.TerraformModule, error) {
	// Get tfconfig of module
	moduleConfig, diags := tfconfig.LoadModule(path)

	if diags.HasErrors() {
		return data.TerraformModule{}, diags.Err()
	}
	resources := make(map[string]data.TerraformCodeBlock)
	// Create map of code blocks (resources, data, variables and outputs) of module
	for _, resource := range moduleConfig.ManagedResources {
		resources[resource.Type+resource.Name] = data.TerraformCodeBlock{
			Name:              resource.Type + " " + resource.Name,
			ModulePath:        path,
			Pos:               resource.Pos.Line,
			FilePath:          resource.Pos.Filename,
			WhitelistComments: []data.WhitelistComment{},
		}
	}
	// Load data
	for _, resource := range moduleConfig.DataResources {
		resources[resource.Type+resource.Name] = data.TerraformCodeBlock{
			Name:              resource.Type + " " + resource.Name,
			ModulePath:        path,
			Pos:               resource.Pos.Line,
			WhitelistComments: []data.WhitelistComment{},
			FilePath:          resource.Pos.Filename,
		}
	}
	// Load variables
	for _, variable := range moduleConfig.Variables {
		resources["variable "+variable.Type+variable.Name] = data.TerraformCodeBlock{
			Name:              "variable " + variable.Type + " " + variable.Name,
			ModulePath:        path,
			Pos:               variable.Pos.Line,
			WhitelistComments: []data.WhitelistComment{},
			FilePath:          variable.Pos.Filename,
		}
	}
	// Load outputs
	for _, output := range moduleConfig.Outputs {
		resources["output"+output.Name] = data.TerraformCodeBlock{
			Name:              "output " + output.Name,
			ModulePath:        path,
			Pos:               output.Pos.Line,
			WhitelistComments: []data.WhitelistComment{},
			FilePath:          output.Pos.Filename,
		}
	}
	module := data.TerraformModule{
		FullPath:     path,
		Name:         filepath.Base(path),
		ModuleConfig: *moduleConfig,
		Resources:    resources,
	}

	return module, nil
}

func GetWhitelistingComments(path string) ([]data.WhitelistComment, error) {
	whitelistComments := []data.WhitelistComment{}
	// Parse the file to get whitelist comments
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	// Read the file and find comments containing guacamole-ignore
	scanner := bufio.NewScanner(file)
	i := 1 //Set cursor to the start of the file
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "guacamole-ignore") {
			whitelistComment := data.WhitelistComment{}
			// Regex to match the check ID in the form of TF/TG_XXX_0XX
			regexp := regexp.MustCompile(`(T[F|G]_(\w+)_\d+)`)
			match := regexp.FindStringSubmatch(line)
			if len(match) > 0 {
				whitelistComment.CheckID = match[0]
				whitelistComment.LineNumber = i
				whitelistComment.Path = path

				// Attach comment to an object
			}
			whitelistComments = append(whitelistComments, whitelistComment)
		}
		i++
	}
	return whitelistComments, nil
}

// TODO: add init and plan layers function
