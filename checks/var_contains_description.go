package checks

import (
	"guacamole/data"
	"guacamole/helpers"
	"strconv"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

func VarContainsDescription() (data.Check, error) {
	name := "Variables should contain a description"
	relatedGuidelines := "https://padok-team.github.io/docs-terraform-guidelines/terraform/terraform_naming.html#variables"
	modules, err := helpers.GetModules()
	if err != nil {
		return data.Check{}, err
	}
	variablesInError := []string{}

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
				variablesInError = append(variablesInError, variable.Pos.Filename+":"+strconv.Itoa(variable.Pos.Line)+" --> "+variable.Name)
			}
		}
	}

	dataCheck := data.Check{
		Name:              name,
		RelatedGuidelines: relatedGuidelines,
		Status:            "✅",
		Errors:            variablesInError,
	}

	if len(variablesInError) > 0 {
		dataCheck.Status = "❌"
	}

	return dataCheck, nil
}
