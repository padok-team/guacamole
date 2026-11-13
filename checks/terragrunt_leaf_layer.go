package checks

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/padok-team/guacamole/data"
	"github.com/spf13/viper"
)

var terragruntCacheRegexp = regexp.MustCompile(`\.terragrunt-cache|\.terraform`)

func TerragruntLeafLayer() (data.Check, error) {
	dataCheck := data.Check{
		ID:                "TG_ARC_001",
		Name:              "terragrunt.hcl should be the last file in the layer",
		RelatedGuidelines: "https://padok-team.github.io/docs-terraform-guidelines/terragrunt/context_pattern.html",
		Status:            "✅",
	}

	codebasePath := viper.GetString("codebase-path")
	layersInError := []data.Error{}

	layers, err := findLayers(codebasePath)
	if err != nil {
		return dataCheck, err
	}

	for _, layer := range layers {
		errors, err := findBelowLayers(filepath.Dir(layer))
		if err != nil {
			return dataCheck, err
		}

		layersInError = append(layersInError, errors...)
	}

	dataCheck.Errors = layersInError
	if len(layersInError) > 0 {
		dataCheck.Status = "❌"
	}

	return dataCheck, nil
}

// Walks a layer directory and returns an error for every inputs.hcl found in a sub-directory, meaning another layer lives below.
func findBelowLayers(layerPath string) ([]data.Error, error) {
	errors := []data.Error{}

	walkErr := filepath.Walk(layerPath, func(subPath string, subInfo os.FileInfo, subErr error) error {
		if subErr != nil {
			return fmt.Errorf("failed to scan layer path %s: %w", layerPath, subErr)
		}

		if subInfo.IsDir() {
			if subPath != layerPath && terragruntCacheRegexp.MatchString(subPath) {
				return filepath.SkipDir
			}
			return nil
		}

		if subInfo.Name() == "inputs.hcl" && filepath.Dir(subPath) != layerPath {
			errors = append(errors, data.Error{
				Path:        subPath,
				LineNumber:  -1,
				Description: fmt.Sprintf("Found below layer to apply: %s", layerPath),
			})
		}

		return nil
	})

	return errors, walkErr
}

// findLayers returns every terragrunt.hcl under the codebase path, from any working directory.
func findLayers(codebasePath string) ([]string, error) {
	layers := []string{}

	err := filepath.Walk(codebasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to scan codebase path: %w", err)
		}

		if terragruntCacheRegexp.MatchString(path) {
			if info.IsDir() {
				return filepath.SkipDir
			}

			return nil
		}

		if info.Name() == "terragrunt.hcl" {
			layers = append(layers, path)
		}

		return nil
	})

	return layers, err
}
