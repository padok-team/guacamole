package checks

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/padok-team/guacamole/data"
	"github.com/spf13/viper"
)

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
			if regexp.MustCompile(`\.terragrunt-cache|\.terraform`).MatchString(path) {
				return filepath.SkipDir
			}
			return nil
		}

		if info.Name() != "terragrunt.hcl" {
			return nil
		}

		if regexp.MustCompile(`\.terragrunt-cache|\.terraform`).MatchString(path) {
			return nil
		}

		layerPath := filepath.Dir(path)

		walkErr := filepath.Walk(layerPath, func(subPath string, subInfo os.FileInfo, subErr error) error {
			if subErr != nil {
				return fmt.Errorf("failed to scan layer path %s: %w", layerPath, subErr)
			}

			if subInfo.IsDir() {
				if subPath != layerPath && regexp.MustCompile(`\.terragrunt-cache|\.terraform`).MatchString(subPath) {
					return filepath.SkipDir
				}
				return nil
			}

			if subInfo.Name() == "inputs.hcl" && filepath.Dir(subPath) != layerPath {
				layersInError = append(layersInError, data.Error{
					Path:        subPath,
					LineNumber:  -1,
					Description: fmt.Sprintf("Found below layer to apply: %s", layerPath),
				})
			}

			return nil
		})
		if walkErr != nil {
			return walkErr
		}

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