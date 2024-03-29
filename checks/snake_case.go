package checks

import (
	"regexp"

	"github.com/padok-team/guacamole/data"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

func SnakeCase(modules map[string]data.TerraformModule) (data.Check, error) {
	dataCheck := data.Check{
		ID:                "TF_NAM_002",
		Name:              "snake_case should be used for all resource names",
		RelatedGuidelines: "https://padok-team.github.io/docs-terraform-guidelines/terraform/terraform_naming.html#resource-andor-data-source-naming",
		Status:            "✅",
	}

	namesInError := []data.Error{}

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
				namesInError = append(namesInError, data.Error{
					Path:        resource.Pos.Filename,
					LineNumber:  resource.Pos.Line,
					Description: resource.MapKey(),
				})
			}
		}

		for _, resource := range moduleConf.DataResources {
			// I want to check if the name of the resource contains any word (separated by a dash) of its type
			matched := matcher.MatchString(resource.Name)
			if !matched {
				namesInError = append(namesInError, data.Error{
					Path:        resource.Pos.Filename,
					LineNumber:  resource.Pos.Line,
					Description: resource.MapKey(),
				})
			}
		}

		for _, resource := range moduleConf.ModuleCalls {
			// I want to check if the name of the resource contains any word (separated by a dash) of its type
			matched := matcher.MatchString(resource.Name)
			if !matched {
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
