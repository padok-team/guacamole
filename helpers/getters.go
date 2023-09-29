package helpers

import (
	"fmt"
	"guacamole/data"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/viper"
)

func GetModules() ([]data.TerraformModule, error) {
	codebasePath := viper.GetString("codebase-path")
	modules := []data.TerraformModule{}
	//Get all subdirectories in root path
	err := filepath.Walk(codebasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to get subdirectories: %w", err)
		}
		// Check if the path is a file and its name matches "*.tf"
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".tf") {
			// exclude the files which are in the .terragrunt-cache directory
			if !regexp.MustCompile(`.terragrunt-cache`).MatchString(path) {
				module := data.TerraformModule{FullPath: filepath.Dir(path), Name: filepath.Base(filepath.Dir(path))}
				// Check if the module is already in the list
				alreadyInList := false
				for _, m := range modules {
					if m.FullPath == module.FullPath {
						alreadyInList = true
					}
				}
				if !alreadyInList {
					modules = append(modules, module)
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
			// exclude the files which are in the .terragrunt-cache directory
			if !regexp.MustCompile(`.terragrunt-cache`).MatchString(path) {
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

// TODO: add init and plan layers function
