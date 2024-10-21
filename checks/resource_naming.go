package checks

import (
	"github.com/padok-team/guacamole/data"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

func ResourceNaming(modules map[string]data.TerraformModule) (data.Check, error) {
	dataCheck := data.Check{
		ID:                "TF_NAM_005",
		Name:              "Resources and data sources should not be named \"this\" or \"these\" if there are more than 1 of the same type",
		RelatedGuidelines: "https://padok-team.github.io/docs-terraform-guidelines/terraform/terraform_naming.html#resource-andor-data-source-naming",
		Status:            "✅",
	}

	resourcesInError := []data.Error{}

	for _, module := range modules {
		moduleConf, diags := tfconfig.LoadModule(module.FullPath)
		if diags.HasErrors() {
			return data.Check{}, diags.Err()
		}
		// Check resources
		for _, resource := range moduleConf.ManagedResources {
			// Check if the type of the resource is unique within the module
			numberOfSameType := 0
			for _, res := range moduleConf.ManagedResources {
				if res.Type == resource.Type {
					numberOfSameType++
				}
			}
			// If there is only one instance of this type of resource, check if its named this or these (If they create more than 1 with a for each)
			if numberOfSameType > 1 {
				if resource.Name == "these" || resource.Name == "this" {
					resourcesInError = append(resourcesInError, data.Error{
						Path:        resource.Pos.Filename,
						LineNumber:  resource.Pos.Line,
						Description: resource.Type + " - " + resource.Name,
					})
				}
			}
		}
		// Check data sources
		for _, dataResource := range moduleConf.DataResources {
			// Check if the type of the resource is unique within the module
			numberOfSameType := 0
			for _, res := range moduleConf.DataResources {
				if res.Type == dataResource.Type {
					numberOfSameType++
				}
			}
			// If there is only one instance of this type of resource, check if its named this or these (If they create more than 1 with a for each)
			if numberOfSameType > 1 {
				if dataResource.Name == "these" || dataResource.Name == "this" {
					resourcesInError = append(resourcesInError, data.Error{
						Path:        dataResource.Pos.Filename,
						LineNumber:  dataResource.Pos.Line,
						Description: dataResource.Type + " - " + dataResource.Name,
					})
				}
			}
		}
	}

	dataCheck.Errors = resourcesInError

	if len(resourcesInError) > 0 {
		dataCheck.Status = "❌"
	}

	return dataCheck, nil
}
