package checks

import (
	"github.com/padok-team/guacamole/data"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

func ProviderInModule(modules []data.TerraformModule) (data.Check, error) {
	dataCheck := data.Check{
		ID:                "TF_MOD_002",
		Name:              "Provider should be defined by the consumer of the module",
		RelatedGuidelines: "https://padok-team.github.io/docs-terraform-guidelines/terraform/donts.html#using-provider-block-in-modules",
		Status:            "✅",
	}

	modulesInError := []data.Error{}
	// For each module, check if the provider is defined
	for _, module := range modules {
		moduleConf, diags := tfconfig.LoadModule(module.FullPath)
		if diags.HasErrors() {
			return data.Check{}, diags.Err()
		}
		//If the module has no provider, display an error
		if len(moduleConf.ProviderConfigs) > 0 {
			modulesInError = append(modulesInError, data.Error{
				Path:        module.FullPath,
				LineNumber:  -1,
				Description: "",
			})
		}
	}

	dataCheck.Errors = modulesInError

	if len(modulesInError) > 0 {
		dataCheck.Status = "❌"
	}

	return dataCheck, nil
}
