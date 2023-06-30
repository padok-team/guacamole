package checks

import (
	"fmt"
	"guacamole/data"
	"guacamole/helpers"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

func ProviderInModule() data.Check {
	name := "No provider in module"
	relatedGuidelines := "https://padok-team.github.io/docs-terraform-guidelines/terraform/donts.html#using-provider-block-in-modules"
	fmt.Println("Checking none prescence provider in module...")
	// Find recusively all the modules in the current directory
	modules, err := helpers.GetModules()
	if err != nil {
		fmt.Println("Error:", err)
	}
	modulesInError := []data.Module{}
	// For each module, check if the provider is defined
	for _, module := range modules {
		moduleConf, diags := tfconfig.LoadModule(module.FullPath)
		if diags.HasErrors() {
			fmt.Println("Error:", diags)
		}
		//If the module has no provider, display an error
		if len(moduleConf.ProviderConfigs) > 0 {
			fmt.Println("Error: provider found")
			fmt.Println("Module:", module.FullPath)
			modulesInError = append(modulesInError, module)
		}

	}
	if len(modulesInError) > 0 {
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
