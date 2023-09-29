package checks

import (
	"fmt"
	"guacamole/data"
	"guacamole/helpers"
	"strconv"
	"strings"

	pluralize "github.com/gertd/go-pluralize"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

func CollectionVarNamePlural() (data.Check, error) {
	name := "Collection variable name must be plural"
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

		// Check if the name of the resource is not in snake case
		for _, variable := range moduleConf.Variables {
			// I want to check if the name of the resource contains any word (separated by a dash) of its type

			// Check if prefix is "list"

			if strings.HasPrefix(variable.Type, "list") || strings.HasPrefix(variable.Type, "set") || strings.HasPrefix(variable.Type, "map") {
				// Check if the variable name is plural
				pluralize := pluralize.NewClient()
				fmt.Println(variable.Name, pluralize.IsPlural(variable.Name))
				if !pluralize.IsPlural(variable.Name) {
					variablesInError = append(variablesInError, variable.Pos.Filename+":"+strconv.Itoa(variable.Pos.Line)+" --> "+variable.Name)
				}
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
