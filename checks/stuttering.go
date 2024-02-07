package checks

import (
	"strings"

	"github.com/padok-team/guacamole/data"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

func Stuttering(modules map[string]data.TerraformModule) (data.Check, error) {
	dataCheck := data.Check{
		ID:                "TF_NAM_003",
		Name:              "Stuttering in the naming of resources",
		RelatedGuidelines: "https://padok-team.github.io/docs-terraform-guidelines/terraform/terraform_naming.html#resource-andor-data-source-naming",
		Status:            "✅",
	}

	namesInError := []data.Error{}
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
				namesInError = append(namesInError, data.Error{
					Path:        resource.Pos.Filename,
					LineNumber:  resource.Pos.Line,
					Description: resource.MapKey(),
				})
			}
		}

		for _, resource := range moduleConf.DataResources {
			// I want to check if the name of the resource contains any word (separated by a dash) of its type
			if containsWord(resource.Name, resource.Type) {
				namesInError = append(namesInError, data.Error{
					Path:        resource.Pos.Filename,
					LineNumber:  resource.Pos.Line,
					Description: resource.MapKey(),
				})
			}
		}

		for _, resource := range moduleConf.ModuleCalls {
			// I want to check if the name of the resource contains any word (separated by a dash) of its type
			if containsWord(resource.Name, "module") {
				namesInError = append(namesInError, data.Error{
					Path:        resource.Pos.Filename,
					LineNumber:  resource.Pos.Line,
					Description: resource.Name,
				})
			}
		}
	}

	dataCheck.Errors = namesInError

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
