package checks

import (
	"regexp"

	"github.com/padok-team/guacamole/data"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

func VarTypeAny(modules map[string]data.TerraformModule) (data.Check, error) {
	dataCheck := data.Check{
		ID:                "TF_VAR_002",
		Name:              "Variable should declare a specific type",
		RelatedGuidelines: "https://padok-team.github.io/docs-terraform-guidelines/terraform/donts.html#using-type-any-in-variables",
		Status:            "✅",
	}

	variablesInError := []data.Error{}

	// Regex to match type any in variables even with spaces and newlines
	matcher := regexp.MustCompile(`any`)

	for _, module := range modules {
		moduleConf, diags := tfconfig.LoadModule(module.FullPath)

		if diags.HasErrors() {
			return data.Check{}, diags.Err()
		}

		for _, variable := range moduleConf.Variables {
			if matcher.MatchString(variable.Type) {
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
