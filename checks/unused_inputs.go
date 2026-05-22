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
		layerConfig, err := parseTerragruntLayerInputs(layerPath)
		if err != nil {
			return dataCheck, err
		}

		if layerConfig.Terraform == nil || layerConfig.Terraform.Source == nil {
			continue
		}

		if layerConfig.Inputs == nil {
			continue
		}

		source := *layerConfig.Terraform.Source

		if isRemoteTerraformSource(source) {
			continue
		}

		modulePath := resolveTerraformModulePath(layerPath, source)

		module, err := helpers.LoadModule(modulePath)
		if err != nil {
			return dataCheck, fmt.Errorf("failed to load module %q from layer %q: %w", modulePath, layerPath, err)
		}

		declaredVariables := map[string]bool{}

		for name := range module.ModuleConfig.Variables {
			declaredVariables[name] = true
		}

		for inputName := range layerConfig.Inputs.AsValueMap() {
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

func parseTerragruntLayerInputs(layerPath string) (TerragruntLayerInputs, error) {
	var layerConfig TerragruntLayerInputs

	parser := hclparse.NewParser()

	hclFile, err := parser.ParseHCLFile(layerPath)
	if err != nil {
		return layerConfig, err
	}

	diags := gohcl.DecodeBody(hclFile.Body, nil, &layerConfig)
	if diags.HasErrors() {
		return layerConfig, diags
	}

	return layerConfig, nil
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
