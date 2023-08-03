package checks

import (
	"guacamole/data"
	"guacamole/helpers"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

func NoStuttering() (data.Check, error) {
	name := "Stuttering in the naming of the resources"
	// relatedGuidelines := "https://padok-team.github.io/docs-terraform-guidelines/terraform/terraform_naming.html#resource-andor-data-source-naming"
	relatedGuidelines := "http://bitly.ws/K5Wm"
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
			if strings.Contains(resource.Name, resource.Type) {
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
