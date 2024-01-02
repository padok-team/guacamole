package checks

import (
	"strconv"

	"github.com/padok-team/guacamole/data"
)

func RefreshTime(layers []*data.Layer) (data.Check, error) {
	checkResult := data.Check{
		Name:              "Layers' refresh time",
		Status:            "✅",
		RelatedGuidelines: "https://github.com/padok-team/docs-terraform-guidelines/blob/main/terraform/wysiwg_patterns.md",
		Errors:            []data.Error{},
	}

	for _, layer := range layers {
		refreshTime := 0
		if layer.State.Values == nil {
			continue
		} else {
			// TODO: check if this way of counting resources counts all nested resources
			// refreshTime := len(layer.State.Values.RootModule.Resources)
			for _, resource := range layer.Plan.ResourceChanges {
				if !resource.Change.Actions.Create() {
					refreshTime++
				}
			}

			if refreshTime > 120 {
				checkResult.Errors = append(checkResult.Errors, data.Error{
					Path:        layer.Name,
					LineNumber:  -1,
					Description: "Number of resources in layer" + strconv.Itoa(refreshTime),
				})
			}
		}
	}

	if len(checkResult.Errors) > 0 {
		checkResult.Status = "❌"
	}

	return checkResult, nil
}
