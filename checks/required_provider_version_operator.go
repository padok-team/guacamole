package checks

import (
	"guacamole/data"
	"guacamole/helpers"
	"regexp"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

func RequiredProviderVersionOperatorInModules() (data.Check, error) {
	name := "Required provider versions in modules should be set with ~> operator"
	relatedGuidelines := "https://padok-team.github.io/docs-terraform-guidelines/terraform/terraform_versioning.html"
	modules, err := helpers.GetModules()
	if err != nil {
		return data.Check{}, err
	}
	requiredProvidersInError := []string{}

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
					requiredProvidersInError = append(requiredProvidersInError, requiredProvider.Source+" --> "+versionConstraint)
				}
			}
		}
	}

	dataCheck := data.Check{
		Name:              name,
		RelatedGuidelines: relatedGuidelines,
		Status:            "✅",
		Errors:            requiredProvidersInError,
	}

	if len(requiredProvidersInError) > 0 {
		dataCheck.Status = "❌"
	}

	return dataCheck, nil
}
