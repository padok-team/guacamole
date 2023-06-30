package checks

import (
	"fmt"
	"guacamole/data"
	"strings"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

func Naming() data.Check {
	name := "Stuttering in the naming of the resources"
	relatedGuidelines := "https://github.com/padok-team/docs-terraform-guidelines/blob/main/terraform/terraform_naming.md"
	fmt.Println("Checking naming convention...")
	modules, err := getModules()
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
