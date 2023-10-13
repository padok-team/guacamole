package checks

import (
	"guacamole/data"
	"guacamole/helpers"
	"strconv"
	"strings"

	pluralize "github.com/gertd/go-pluralize"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

func VarNumberMatchesType() (data.Check, error) {
	name := "A variable name's number matches its type"
	relatedGuidelines := "https://t.ly/A7P5j"
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

		for _, variable := range moduleConf.Variables {
			// Check if prefix is "list"
			isCollection := strings.HasPrefix(variable.Type, "list") || strings.HasPrefix(variable.Type, "set") || strings.HasPrefix(variable.Type, "map")
			pluralize := pluralize.NewClient()

			// Remove all spaces and new lines from the type
			variable.Type = strings.ReplaceAll(variable.Type, "\n", "")
			variable.Type = strings.ReplaceAll(variable.Type, " ", "")

			if isCollection && !pluralize.IsPlural(variable.Name) || !isCollection && !pluralize.IsSingular(variable.Name) {
				variablesInError = append(variablesInError, variable.Pos.Filename+":"+strconv.Itoa(variable.Pos.Line)+" --> "+variable.Name+" <> "+variable.Type)
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
