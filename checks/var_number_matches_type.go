package checks

import (
	"slices"
	"strings"

	"github.com/padok-team/guacamole/data"

	pluralize "github.com/gertd/go-pluralize"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

var pluralEquivalentWords = []string{
	"whitelist",
	"allowlist",
	"denylist",
	"blacklist",
}

func VarNumberMatchesType(modules map[string]data.TerraformModule) (data.Check, error) {
	dataCheck := data.Check{
		ID:                "TF_NAM_004",
		Name:              "Variable name's number should match its type",
		RelatedGuidelines: "https://padok-team.github.io/docs-terraform-guidelines/terraform/terraform_naming.html#variables",
		Status:            "✅",
	}

	variablesInError := []data.Error{}
	pluralizeClient := pluralize.NewClient()
	pluralizeClient.AddIrregularRule("uri", "uris")

	// For each module, check if the provider is defined
	for _, module := range modules {
		moduleConf, diags := tfconfig.LoadModule(module.FullPath)
		if diags.HasErrors() {
			return data.Check{}, diags.Err()
		}

		for _, variable := range moduleConf.Variables {
			isCollection := strings.HasPrefix(variable.Type, "list") || strings.HasPrefix(variable.Type, "set") || strings.HasPrefix(variable.Type, "map")

			// Remove all spaces and new lines from the type
			variable.Type = strings.ReplaceAll(variable.Type, "\n", "")
			variable.Type = strings.ReplaceAll(variable.Type, " ", "")
			nameParts := strings.Split(variable.Name, "_")
			lastWord := nameParts[len(nameParts)-1]
			hasPluralEquivalentWord := slices.Contains(pluralEquivalentWords, lastWord)
			if isCollection && !pluralizeClient.IsPlural(variable.Name) && !hasPluralEquivalentWord {
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
