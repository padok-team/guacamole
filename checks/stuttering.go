package checks

import (
	"guacamole/data"
	"guacamole/helpers"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

func Stuttering() (data.Check, error) {
	name := "Stuttering in the naming of resources"
	relatedGuidelines := "https://padok-team.github.io/docs-terraform-guidelines/terraform/terraform_naming.html#resource-andor-data-source-naming"
	modules, err := helpers.GetModules()
	if err != nil {
		return data.Check{}, err
	}
	namesInError := []string{}
	// For each module, check if the provider is defined
	for _, module := range modules {
		moduleConf, diags := tfconfig.LoadModule(module.FullPath)
		if diags.HasErrors() {
			return data.Check{}, diags.Err()
		}
		//Check if the name of the resource is not a duplicate of its type
		for _, resource := range moduleConf.ManagedResources {
			// I want to check if the name of the resource contains any word (separated by a dash) of its type
			if containsWord(resource.Name, resource.Type) {
				namesInError = append(namesInError, resource.Pos.Filename+":"+strconv.Itoa(resource.Pos.Line)+" --> "+resource.MapKey())
			}
		}

		for _, resource := range moduleConf.DataResources {
			// I want to check if the name of the resource contains any word (separated by a dash) of its type
			if containsWord(resource.Name, resource.Type) {
				namesInError = append(namesInError, resource.Pos.Filename+":"+strconv.Itoa(resource.Pos.Line)+" --> "+resource.MapKey())
			}
		}

		for _, resource := range moduleConf.ModuleCalls {
			// I want to check if the name of the resource contains any word (separated by a dash) of its type
			if containsWord(resource.Name, "module") {
				namesInError = append(namesInError, resource.Pos.Filename+":"+strconv.Itoa(resource.Pos.Line)+" --> "+resource.Name)
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

func containsWord(s1, s2 string) bool {
	// Split the strings into words taking into account multiple possible separators
	// A name must start with a letter or underscore and may contain only letters, digits, underscores, and dashes.
	words1 := strings.FieldsFunc(s1, func(r rune) bool {
		return r == '-' || r == '_'
	})
	words2 := strings.FieldsFunc(s2, func(r rune) bool {
		return r == '-' || r == '_'
	})

	for _, word2 := range words2 {
		for _, word1 := range words1 {
			if word1 == word2 {
				return true // word from s2 found in s1
			}
		}
	}
	return false
}
