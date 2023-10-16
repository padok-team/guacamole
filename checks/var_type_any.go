package checks

import (
	"guacamole/data"
	"guacamole/helpers"
	"regexp"
	"strconv"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

func VarTypeAny() (data.Check, error) {
	name := "Variables should declare a specific type"
	relatedGuidelines := "https://xxx"
	modules, err := helpers.GetModules()
	if err != nil {
		return data.Check{}, err
	}
	variablesInError := []string{}

	// Regex to match type any in variables even with spaces and newlines
	matcher := regexp.MustCompile(`any`)

	for _, module := range modules {
		moduleConf, diags := tfconfig.LoadModule(module.FullPath)

		if diags.HasErrors() {
			return data.Check{}, diags.Err()
		}

		for _, variable := range moduleConf.Variables {
			if matcher.MatchString(variable.Type) {
				variablesInError = append(variablesInError, variable.Pos.Filename+":"+strconv.Itoa(variable.Pos.Line)+" --> "+variable.Name)
			}
		}
	}

	dataCheck := data.Check{
		Name:              name,
		RelatedGuidelines: relatedGuidelines,
		Status:            "✅",
		Errors:            variablesInError,
	}

	if len(variablesInError) > 0 {
		dataCheck.Status = "❌"
	}

	return dataCheck, nil
}
