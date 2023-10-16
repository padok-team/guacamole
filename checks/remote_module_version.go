package checks

import (
	"guacamole/data"
	"guacamole/helpers"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

func RemoteModuleVersion() (data.Check, error) {
	name := "Remote module call should be pinned to a specific version"
	relatedGuidelines := "https://padok-team.github.io/docs-terraform-guidelines/terraform/terraform_versioning.html"
	modules, err := helpers.GetModules()
	if err != nil {
		return data.Check{}, err
	}
	modulesInError := []string{}

	// Regex versionMatcher that matches a specific version number
	// Example: v1.2.3
	versionMatcher := regexp.MustCompile(`^v?\d+\.\d+\.\d+(-\w+)?$`)
	// Regex matcher for git repository link with a tag
	gitRefMatcher := regexp.MustCompile(`^git::https:\/\/github\.com\/.*\.git\?ref=.*$`)

	for _, module := range modules {
		moduleConf, diags := tfconfig.LoadModule(module.FullPath)
		if diags.HasErrors() {
			return data.Check{}, diags.Err()
		}

		for _, moduleCall := range moduleConf.ModuleCalls {
			// Check if the module is a remote module and not a local one
			if moduleCall.Source != "" && moduleCall.Source[0] != '.' {
				// Check if the module comes from the Terraform Registry or from a git repository
				if strings.HasPrefix(moduleCall.Source, "git::") {
					// If the module comes from a git repository, check if the version is a tag
					if !gitRefMatcher.MatchString(moduleCall.Source) {
						modulesInError = append(modulesInError, moduleCall.Pos.Filename+":"+strconv.Itoa(moduleCall.Pos.Line)+" --> "+moduleCall.Name)
					}
				} else {
					if !versionMatcher.MatchString(moduleCall.Version) {
						checkString := moduleCall.Pos.Filename + ":" + strconv.Itoa(moduleCall.Pos.Line) + " --> " + moduleCall.Name
						if moduleCall.Version != "" {
							checkString += " / " + moduleCall.Version
						}
						modulesInError = append(modulesInError, checkString)
					}
				}
			}
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
