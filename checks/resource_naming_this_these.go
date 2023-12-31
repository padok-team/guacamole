package checks

import (
	"strconv"

	"github.com/padok-team/guacamole/data"
	"github.com/padok-team/guacamole/helpers"

	pluralize "github.com/gertd/go-pluralize"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

func ResourceNamingThisThese() (data.Check, error) {
	dataCheck := data.Check{
		ID:                "TF_NAM_001",
		Name:              "Resources and datasources in modules should be named \"this\" or \"these\" if their type is unique",
		RelatedGuidelines: "https://padok-team.github.io/docs-terraform-guidelines/terraform/terraform_naming.html#resource-andor-data-source-naming",
		Status:            "✅",
	}

	resourcesInError := []string{}

	modules, err := helpers.GetModules()
	if err != nil {
		return data.Check{}, err
	}
	pluralize := pluralize.NewClient()

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
			if numberOfSameType == 1 {
				if pluralize.IsPlural(resource.Name) {
					if resource.Name != "these" {
						resourcesInError = append(resourcesInError, resource.Pos.Filename+":"+strconv.Itoa(resource.Pos.Line)+" --> "+resource.Type+" - "+resource.Name)
					}
				} else {
					if resource.Name != "this" {
						resourcesInError = append(resourcesInError, resource.Pos.Filename+":"+strconv.Itoa(resource.Pos.Line)+" --> "+resource.Type+" - "+resource.Name)
					}
				}
			}
		}
		// Check data sources
		for _, data := range moduleConf.DataResources {
			// Check if the type of the resource is unique within the module
			numberOfSameType := 0
			for _, res := range moduleConf.DataResources {
				if res.Type == data.Type {
					numberOfSameType++
				}
			}
			// If there is only one instance of this type of resource, check if its named this or these (If they create more than 1 with a for each)
			if numberOfSameType == 1 {
				if pluralize.IsPlural(data.Name) {
					if data.Name != "these" {
						resourcesInError = append(resourcesInError, data.Pos.Filename+":"+strconv.Itoa(data.Pos.Line)+" --> "+data.Type+" - "+data.Name)
					}
				} else {
					if data.Name != "this" {
						resourcesInError = append(resourcesInError, data.Pos.Filename+":"+strconv.Itoa(data.Pos.Line)+" --> "+data.Type+" - "+data.Name)
					}
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
