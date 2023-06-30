package checks

import (
	"fmt"
	"guacamole/data"
	"guacamole/helpers"
	"strings"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

func Naming() data.Check {
	name := "Stuttering in the naming of the resources"
	relatedGuidelines := "https://padok-team.github.io/docs-terraform-guidelines/terraform/terraform_naming.html#resource-andor-data-source-naming"
	fmt.Println("Checking naming convention...")
	modules, err := helpers.GetModules()
	if err != nil {
		fmt.Println("Error:", err)
	}
	namesInError := []string{}
	// For each module, check if the provider is defined
	for _, module := range modules {
		moduleConf, diags := tfconfig.LoadModule(module.FullPath)
		if diags.HasErrors() {
			fmt.Println("Error:", diags)
		}
		//Check if the name of the resource is not a duplicate of its type
		for _, resource := range moduleConf.ManagedResources {
			if strings.Contains(resource.Type, resource.Name) {
				namesInError = append(namesInError, resource.Name)
				fmt.Println("Error: The name of the resource '" + resource.Name + "' is contained in its type '" + resource.Type + "'")
				fmt.Println(resource.Pos.Filename)
			}
		}

	}
	if len(namesInError) > 0 {
		return data.Check{
			Name:              name,
			RelatedGuidelines: relatedGuidelines,
			Status:            "❌",
		}
	}
	return data.Check{
		Name:              name,
		RelatedGuidelines: relatedGuidelines,
		Status:            "✅",
	}
}
