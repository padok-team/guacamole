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
	root := viper.GetString("codebase-path") // Root directory to start browsing from
	layers := []data.Layer{}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to get layer subdirectory: %w", err)
		}

		// Check if the current path is a file and its name matches "terragrunt.hcl"
		if !info.IsDir() && info.Name() == "terragrunt.hcl" {
			// exclude the files which are in the .terragrunt-cache directory
			if !regexp.MustCompile(`.terragrunt-cache`).MatchString(path) {
				// TODO: for name, remove first and last element, for fullpath remove only last element
				splitPath := strings.Split(path, "/")
				fullPath := strings.Join(splitPath[:len(splitPath)-1], "/")
				name := strings.Join(splitPath[1:len(splitPath)-1], "/")
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
