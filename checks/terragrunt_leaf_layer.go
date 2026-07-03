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

	err := filepath.Walk(codebasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to scan codebase path: %w", err)
		}

		if info.IsDir() {
			if terragruntCacheRegexp.MatchString(path) {
				return filepath.SkipDir
			}
			return nil
		}

		if !isLayerRoot(info, path) {
			return nil
		}

		errors, walkErr := findBelowLayers(filepath.Dir(path))
		if walkErr != nil {
			return walkErr
		}
		layersInError = append(layersInError, errors...)

		return nil
	})
	if err != nil {
		return dataCheck, err
	}

	dataCheck.Errors = layersInError
	if len(layersInError) > 0 {
		dataCheck.Status = "❌"
	}

	return dataCheck, nil
}

// Reports whether the given file is a terragrunt.hcl defining a layer, ignoring the ones nested in terragrunt/terraform cache directories.
func isLayerRoot(info os.FileInfo, path string) bool {
	return info.Name() == "terragrunt.hcl" && !terragruntCacheRegexp.MatchString(path)
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
