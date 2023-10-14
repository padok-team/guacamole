package checks

import (
	"guacamole/data"
	"guacamole/helpers"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

func ProviderInModule() (data.Check, error) {
	name := "No provider in module"
	relatedGuidelines := "https://padok-team.github.io/docs-terraform-guidelines/terraform/donts.html#using-provider-block-in-modules"
	// Find recusively all the modules in the current directory
	modules, err := helpers.GetModules()
	if err != nil {
		return data.Check{}, err
	}
	modulesInError := []string{}
	// For each module, check if the provider is defined
	for _, module := range modules {
		moduleConf, diags := tfconfig.LoadModule(module.FullPath)
		if diags.HasErrors() {
			return data.Check{}, diags.Err()
		}
		//If the module has no provider, display an error
		if len(moduleConf.ProviderConfigs) > 0 {
			modulesInError = append(modulesInError, module.FullPath)
		}
	}

	dataCheck := data.Check{
		Name:              name,
		RelatedGuidelines: relatedGuidelines,
		Status:            "✅",
		Errors:            modulesInError,
	}

	if len(modulesInError) > 0 {
		dataCheck.Status = "❌"
	}

	return dataCheck, nil
}
