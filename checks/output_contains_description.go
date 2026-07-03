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

	// For each module, check that all outputs have a non-empty description
	for _, module := range modules {
		moduleConf := module.ModuleConfig
		for _, output := range moduleConf.Outputs {
			if output.Description == "" {
				dataCheck.Errors = append(dataCheck.Errors, data.Error{
					Path:        output.Pos.Filename,
					LineNumber:  output.Pos.Line,
					Description: output.Name,
				})
			}
		}
	}

	if len(dataCheck.Errors) > 0 {
		dataCheck.Status = "❌"
	}

	return dataCheck, nil
}
