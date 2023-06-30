package helpers

import (
	"fmt"
	"guacamole/data"
	"os"
	"path/filepath"
	"regexp"

	"github.com/spf13/viper"
)

func GetModules() ([]data.Module, error) {
	codebasePath := viper.GetString("codebase-path") + "modules/"
	modules := []data.Module{}
	//Get all subdirectories in root path
	err := filepath.Walk(codebasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println("Error:", err)
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
			return err
		}

		// Check if the current path is a file and its name matches "terragrunt.hcl"
		if !info.IsDir() && info.Name() == "terragrunt.hcl" {
			// exclude the files which are in the .terragrunt-cache directory
			if !regexp.MustCompile(`.terragrunt-cache`).MatchString(path) {
				layers = append(layers, data.Layer{Name: path[len(root) : len(path)-len(info.Name())-1], FullPath: path[:len(path)-len(info.Name())-1]})
			}
		}

		return nil
	})

	if err != nil {
		fmt.Println("Error:", err)
	}
	return layers, nil
}
