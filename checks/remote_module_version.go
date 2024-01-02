package checks

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/padok-team/guacamole/data"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

func RemoteModuleVersion(modules []data.TerraformModule) (data.Check, error) {
	dataCheck := data.Check{
		ID:                "TF_MOD_001",
		Name:              "Remote module call should be pinned to a specific version",
		RelatedGuidelines: "https://padok-team.github.io/docs-terraform-guidelines/terraform/terraform_versioning.html#module-layer-versioning",
		Status:            "✅",
	}

	modulesInError := []data.Error{}

	// Regex versionMatcher that matches a specific version number
	// Example: v1.2.3
	versionMatcher := regexp.MustCompile(`^v?\d+\.\d+\.\d+(-\w+)?$`)
	// Regex matcher for git repository link with a tag
	gitRefMatcher := regexp.MustCompile(`^git.+\?ref=.*$`)

	for _, module := range modules {
		moduleConf, diags := tfconfig.LoadModule(module.FullPath)
		if diags.HasErrors() {
			return data.Check{}, diags.Err()
		}

		for _, moduleCall := range moduleConf.ModuleCalls {
			// Check if the module is a remote module and not a local one
			if moduleCall.Source != "" && moduleCall.Source[0] != '.' {
				// Check if the module comes from the Terraform Registry or from a git repository
				if strings.HasPrefix(moduleCall.Source, "git") {
					// If the module comes from a git repository, check if the version is a tag
					if !gitRefMatcher.MatchString(moduleCall.Source) {
						modulesInError = append(modulesInError, data.Error{
							Path:        moduleCall.Pos.Filename,
							LineNumber:  moduleCall.Pos.Line,
							Description: moduleCall.Name,
						})
					}
				} else {
					if !versionMatcher.MatchString(moduleCall.Version) {
						checkString := moduleCall.Pos.Filename + ":" + strconv.Itoa(moduleCall.Pos.Line) + " --> " + moduleCall.Name
						if moduleCall.Version != "" {
							checkString += " / " + moduleCall.Version
						}
						modulesInError = append(modulesInError, data.Error{
							Path:        moduleCall.Pos.Filename,
							LineNumber:  moduleCall.Pos.Line,
							Description: moduleCall.Name,
						})
					}
				}
			}
		}
	}

	dataCheck.Errors = modulesInError

	if len(modulesInError) > 0 {
		dataCheck.Status = "❌"
	}

	return dataCheck, nil
}
