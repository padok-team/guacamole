package checks

import (
	"github.com/padok-team/guacamole/data"
	"github.com/padok-team/guacamole/helpers"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

func ProviderInModule() (data.Check, error) {
	dataCheck := data.Check{
		ID:                "TF_MOD_002",
		Name:              "Provider should be defined by the consumer of the module",
		RelatedGuidelines: "https://padok-team.github.io/docs-terraform-guidelines/terraform/donts.html#using-provider-block-in-modules",
		Status:            "✅",
	}

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

	dataCheck.Errors = modulesInError

	if len(modulesInError) > 0 {
		dataCheck.Status = "❌"
	}

	return dataCheck, nil
}
