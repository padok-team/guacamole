package checks

import (
	"strings"

	"github.com/padok-team/guacamole/data"

	pluralize "github.com/gertd/go-pluralize"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

func VarNumberMatchesType(modules map[string]data.TerraformModule) (data.Check, error) {
	dataCheck := data.Check{
		ID:                "TF_NAM_004",
		Name:              "Variable name's number should match its type",
		RelatedGuidelines: "https://padok-team.github.io/docs-terraform-guidelines/terraform/terraform_naming.html#variables",
		Status:            "✅",
	}

	variablesInError := []data.Error{}

	// For each module, check if the provider is defined
	for _, module := range modules {
		moduleConf, diags := tfconfig.LoadModule(module.FullPath)
		if diags.HasErrors() {
			return data.Check{}, diags.Err()
		}

		for _, variable := range moduleConf.Variables {
			// Check if prefix is "list"
			isCollection := strings.HasPrefix(variable.Type, "list") || strings.HasPrefix(variable.Type, "set") || strings.HasPrefix(variable.Type, "map")
			pluralize := pluralize.NewClient()
			// Add irregular rules
			// If this list gets too big, we should consider moving it to a file or even finding a better way to handle this
			pluralize.AddIrregularRule("uri", "uris")

			// Remove all spaces and new lines from the type
			variable.Type = strings.ReplaceAll(variable.Type, "\n", "")
			variable.Type = strings.ReplaceAll(variable.Type, " ", "")
			if isCollection && !pluralize.IsPlural(variable.Name) {
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
