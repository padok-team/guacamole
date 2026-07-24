package checks

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gruntwork-io/terragrunt/config"
	"github.com/gruntwork-io/terragrunt/options"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/padok-team/guacamole/data"
	"github.com/padok-team/guacamole/helpers"
	"github.com/spf13/viper"
	"github.com/zclconf/go-cty/cty"
)

type TerragruntLayerInputs struct {
  Terraform *struct {
    Source *string `hcl:"source,attr"`
  } `hcl:"terraform,block"`

  Inputs *cty.Value `hcl:"inputs,attr"`
}

func UnusedInputs() (data.Check, error) {
	dataCheck := data.Check{
		ID:                "TG_INP_001",
		Name:              "Terragrunt inputs should match variables declared in the called module",
		RelatedGuidelines: "https://padok-team.github.io/docs-terraform-guidelines/terraform/remove_unused_variables.html",
		Status:            "✅",
	}

	codebasePath := viper.GetString("codebase-path")

	terragruntOptions := options.NewTerragruntOptions()
	// Get all layers
	layers, err := config.FindConfigFilesInPath(codebasePath, terragruntOptions)
	if err != nil {
		return dataCheck, err
	}

	errors := []data.Error{}

	for _, layerPath := range layers {
		source, inputs, err := resolveLayerSourceAndInputs(layerPath)
		if err != nil {
			return dataCheck, err
		}

		// Skip layers that don't point to a module or don't set any input.
		if source == nil || len(inputs) == 0 {
			continue
		}

		if isRemoteTerraformSource(*source) {
			continue
		}

		modulePath := resolveTerraformModulePath(layerPath, *source)

		module, err := helpers.LoadModule(modulePath)
		if err != nil {
			return dataCheck, fmt.Errorf("failed to load module %q from layer %q: %w", modulePath, layerPath, err)
		}

		declaredVariables := map[string]bool{}

		for name := range module.ModuleConfig.Variables {
			declaredVariables[name] = true
		}

		for inputName := range inputs {
			if declaredVariables[inputName] {
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

// resolveLayerSourceAndInputs resolves a Terragrunt layer the way Terragrunt
// itself composes it: it follows the layer's include blocks and merges every
// file that makes up the layer. This matters because most codebases use the
// "context pattern", where the terraform{} block and the inputs live in
// separate included files (module.hcl, inputs.hcl, ...) rather than directly in
// the layer's terragrunt.hcl. Parsing terragrunt.hcl in isolation would find
// neither, and the check would silently pass for every layer.
//
// It returns the resolved module source (if any) and the union of the input
// names declared across all of the layer's files.
func resolveLayerSourceAndInputs(layerPath string) (*string, map[string]cty.Value, error) {
	// Discover every file that composes the layer: the terragrunt.hcl plus all
	// of its processed includes. findFilesInLayers is shared with the Dry check.
	files, err := findFilesInLayers(layerPath)
	if err != nil {
		return nil, nil, err
	}

	parser := hclparse.NewParser()

	var source *string
	inputs := map[string]cty.Value{}

	for _, file := range files {
		hclFile, diags := parser.ParseHCLFile(file)
		if diags.HasErrors() {
			return nil, nil, diags
		}

		var layerConfig TerragruntLayerInputs
		// Each file only defines a subset of the layer (include/locals/dependency
		// blocks live elsewhere), so we deliberately ignore the strict-decoding
		// diagnostics and merge whatever terraform.source / inputs each file sets.
		_ = gohcl.DecodeBody(hclFile.Body, nil, &layerConfig)

		if layerConfig.Terraform != nil && layerConfig.Terraform.Source != nil {
			source = layerConfig.Terraform.Source
		}

		if layerConfig.Inputs != nil {
			value := *layerConfig.Inputs
			if value.Type().IsObjectType() || value.Type().IsMapType() {
				for name, v := range value.AsValueMap() {
					inputs[name] = v
				}
			}
		}
	}

	return source, inputs, nil
}

func resolveTerraformModulePath(layerPath string, source string) string {
	if filepath.IsAbs(source) {
		return filepath.Clean(source)
	}

	return filepath.Clean(filepath.Join(filepath.Dir(layerPath), source))
}

func isRemoteTerraformSource(source string) bool {
	remotePrefixes := []string{
		"git::",
		"github.com/",
		"git@",
		"https://",
		"http://",
		"tfr://",
	}

	for _, prefix := range remotePrefixes {
		if strings.HasPrefix(source, prefix) {
			return true
		}
	}

	return false
}
