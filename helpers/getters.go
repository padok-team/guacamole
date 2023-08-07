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

func GetModules() ([]data.Module, error) {
	codebasePath := filepath.Join(viper.GetString("codebase-path"), "modules")
	modules := []data.Module{}
	//Get all subdirectories in root path
	err := filepath.Walk(codebasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to get subdirectories: %w", err)
		}
		if info.IsDir() && path != codebasePath {
			modules = append(modules, data.Module{Name: info.Name(), FullPath: path})
		}
		return nil
	})
	return modules, err
}

func GetLayers() ([]data.Layer, error) {
	codebasePath := viper.GetString("codebase-path") // Root directory to start browsing from
	codebaseDirName := filepath.Base(codebasePath)

	layers := []data.Layer{}

	err := filepath.Walk(codebasePath, func(path string, info os.FileInfo, err error) error {
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
				name := strings.Split(absPath, codebaseDirName+"/")[1]
				name = strings.ReplaceAll(name, "/terragrunt.hcl", "")
				layers = append(layers, data.Layer{Name: name, FullPath: fullPath})
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
