package checks

import (
	"github.com/padok-team/guacamole/data"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

func VarContainsDescription(modules []data.TerraformModule) (data.Check, error) {
	dataCheck := data.Check{
		ID:                "TF_VAR_001",
		Name:              "Variable should contain a description",
		RelatedGuidelines: "https://padok-team.github.io/docs-terraform-guidelines/terraform/donts.html#variables",
		Status:            "✅",
	}

	variablesInError := []data.Error{}

	// For each module, check if the provider is defined
	for _, module := range modules {
		moduleConf, diags := tfconfig.LoadModule(module.FullPath)

		if diags.HasErrors() {
			return data.Check{}, diags.Err()
		}

		// Check if the name of the resource is not in snake case
		for _, variable := range moduleConf.Variables {
			// I want to check if the name of the resource contains any word (separated by a dash) of its type

			if variable.Description == "" {
				variablesInError = append(variablesInError, data.Error{
					Path:        variable.Pos.Filename,
					LineNumber:  variable.Pos.Line,
					Description: variable.Name,
				})
			}
		}
	}

	dataCheck.Errors = variablesInError

	if len(variablesInError) > 0 {
		dataCheck.Status = "❌"
	}

	return dataCheck, nil
}
