package checks

import (
	"guacamole/data"
	"guacamole/helpers"
	"regexp"
	"strconv"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

func SnakeCase() (data.Check, error) {
	name := "snake_case in the naming of the resources"
	// relatedGuidelines := "https://padok-team.github.io/docs-terraform-guidelines/terraform/terraform_naming.html#resource-andor-data-source-naming"
	relatedGuidelines := "https://t.ly/HQM86"
	modules, err := helpers.GetModules()
	if err != nil {
		return data.Check{}, err
	}
	namesInError := []string{}

	pattern := `^[a-z0-9_]+$`
	matcher, err := regexp.Compile(pattern)
	if err != nil {
		return data.Check{}, err
	}

	// For each module, check if the provider is defined
	for _, module := range modules {
		moduleConf, diags := tfconfig.LoadModule(module.FullPath)
		if diags.HasErrors() {
			return data.Check{}, diags.Err()
		}

		// Check if the name of the resource is not in snake case
		for _, resource := range moduleConf.ManagedResources {
			// I want to check if the name of the resource contains any word (separated by a dash) of its type
			matched := matcher.MatchString(resource.Name)
			if !matched {
				namesInError = append(namesInError, resource.Pos.Filename+":"+strconv.Itoa(resource.Pos.Line)+" --> "+resource.MapKey())
			}
		}
	}

	dataCheck := data.Check{
		Name:              name,
		RelatedGuidelines: relatedGuidelines,
		Status:            "✅",
		Errors:            namesInError,
	}

	if len(namesInError) > 0 {
		dataCheck.Status = "❌"
	}

	return dataCheck, nil
}
