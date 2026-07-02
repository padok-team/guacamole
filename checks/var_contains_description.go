package checks

import (
	"github.com/padok-team/guacamole/data"
)

func VarContainsDescription(modules map[string]data.TerraformModule) (data.Check, error) {
	dataCheck := data.Check{
		ID:                "TF_VAR_001",
		Name:              "Variable should contain a description",
		RelatedGuidelines: "https://padok-team.github.io/docs-terraform-guidelines/terraform/terraform_naming.html#variables",
		Status:            "✅",
	}

	variablesInError := []data.Error{}

	// For each module, check that all variables have a non-empty description
	for _, module := range modules {
		moduleConf := module.ModuleConfig
		for _, variable := range moduleConf.Variables {
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
