package checks

import (
	"regexp"

	"github.com/padok-team/guacamole/data"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

func RequiredProviderVersionOperatorInModules(modules map[string]data.TerraformModule) (data.Check, error) {
	dataCheck := data.Check{
		ID:                "TF_MOD_003",
		Name:              "Required provider versions in modules should be set with ~> operator",
		RelatedGuidelines: "https://padok-team.github.io/docs-terraform-guidelines/terraform/terraform_versioning.html#required-providers-version-for-modules",
		Status:            "✅",
	}

	requiredProvidersInError := []data.Error{}

	pattern := `~>`
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

		for _, requiredProvider := range moduleConf.RequiredProviders {
			for _, versionConstraint := range requiredProvider.VersionConstraints {
				matched := matcher.MatchString(versionConstraint)
				if !matched {
					requiredProvidersInError = append(requiredProvidersInError, data.Error{
						Path:        module.FullPath + " - " + requiredProvider.Source,
						LineNumber:  -1,
						Description: requiredProvider.Source + " --> " + versionConstraint,
					})
				}
			}
		}
	}

	dataCheck.Errors = requiredProvidersInError

	if len(requiredProvidersInError) > 0 {
		dataCheck.Status = "❌"
	}

	return dataCheck, nil
}
