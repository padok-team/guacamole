package checks

import (
	"github.com/padok-team/guacamole/data"
)

func OutputContainsDescription(modules map[string]data.TerraformModule) (data.Check, error) {
	dataCheck := data.Check{
		ID:                "TF_OUT_001",
		Name:              "Output should contain a description",
		RelatedGuidelines: "https://padok-team.github.io/docs-terraform-guidelines/terraform/terraform_naming.html#module-outputs",
		Status:            "✅",
	}

	outputsInError := []data.Error{}

	// For each module, check that all outputs have a non-empty description
	for _, module := range modules {
		moduleConf := module.ModuleConfig
		for _, output := range moduleConf.Outputs {
			if output.Description == "" {
				outputsInError = append(outputsInError, data.Error{
					Path:        output.Pos.Filename,
					LineNumber:  output.Pos.Line,
					Description: output.Name,
				})
			}
		}
	}

	dataCheck.Errors = outputsInError

	if len(outputsInError) > 0 {
		dataCheck.Status = "❌"
	}

	return dataCheck, nil
}
