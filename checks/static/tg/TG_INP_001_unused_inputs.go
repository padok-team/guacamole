package tg

import (
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"strings"

	"github.com/gruntwork-io/terragrunt/pkg/config"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/padok-team/guacamole/data"
	"github.com/padok-team/guacamole/helpers"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

func UnusedInputs() (data.Check, error) {
	dataCheck := data.Check{
		ID:                "TG_INP_001",
		Name:              "Terragrunt inputs should match variables declared in the called module",
		RelatedGuidelines: "https://padok-team.github.io/docs-terraform-guidelines/terraform/remove_unused_variables.html",
		Status:            "✅",
	}

	codebasePath := viper.GetString("codebase-path")

	layers, err := findLayers(codebasePath)
	if err != nil {
		return dataCheck, err
	}

	errors := []data.Error{}

	for _, layerPath := range layers {
		inputNames, err := resolveLayerInputNames(layerPath)
		if err != nil {
			log.Warnf("[TG_INP_001] skipping layer %s: %s", layerPath, err)
			continue
		}

		// Terragrunt evaluates the source (includes, locals, get_repo_root(), ...).
		source, err := resolveLayerSource(layerPath)
		if err != nil {
			log.Warnf("[TG_INP_001] skipping layer %s: failed to resolve terraform.source: %s", layerPath, err)
			continue
		}

		// Only local sources can be read: skip registry, git::, tfr://, ... and layers without inputs.
		if len(inputNames) == 0 || !(strings.HasPrefix(source, "./") || strings.HasPrefix(source, "../") || filepath.IsAbs(source)) {
			continue
		}

		modulePath := source
		if !filepath.IsAbs(modulePath) {
			modulePath = filepath.Join(filepath.Dir(layerPath), source)
		}

		// An unreadable module must not abort the run (and the other checks with it).
		module, err := helpers.LoadModule(modulePath)
		if err != nil {
			log.Warnf("[TG_INP_001] skipping layer %s: failed to load module %q: %s", layerPath, modulePath, err)
			continue
		}

		for _, inputName := range inputNames {
			if _, declared := module.ModuleConfig.Variables[inputName]; declared {
				continue
			}

			errors = append(errors, data.Error{
				Path:       layerPath,
				LineNumber: -1,
				Description: fmt.Sprintf(
					"input %q is not declared as a variable in module %q",
					inputName,
					modulePath,
				),
			})
		}
	}

	dataCheck.Errors = errors

	if len(errors) > 0 {
		dataCheck.Status = "❌"
	}

	return dataCheck, nil
}

func resolveLayerSource(layerPath string) (string, error) {
	layerConfig, err := partialParseLayer(layerPath, config.TerraformSource)
	if err != nil || layerConfig.Terraform == nil || layerConfig.Terraform.Source == nil {
		return "", err
	}

	return *layerConfig.Terraform.Source, nil
}

// resolveLayerInputNames returns the inputs set by the layer and its includes (context pattern).
func resolveLayerInputNames(layerPath string) ([]string, error) {
	files, err := findFilesInLayers(layerPath)
	if err != nil {
		return nil, err
	}

	// merge() is the only function supported, the other ones make the inputs unreadable.
	evalContext := &hcl.EvalContext{Functions: map[string]function.Function{"merge": stdlib.MergeFunc}}
	parser := hclparse.NewParser()
	names := map[string]bool{}

	for _, file := range files {
		hclFile, diags := parser.ParseHCLFile(file)
		if diags.HasErrors() {
			return nil, diags
		}

		// Decoding errors are expected (locals, dependencies, ...): the keys are still decoded.
		var input Input
		_ = gohcl.DecodeBody(hclFile.Body, evalContext, &input)

		if input.Inputs != nil && input.Inputs.IsKnown() && (input.Inputs.Type().IsObjectType() || input.Inputs.Type().IsMapType()) {
			for name := range input.Inputs.AsValueMap() {
				names[name] = true
			}
		}
	}

	return slices.Sorted(maps.Keys(names)), nil
}