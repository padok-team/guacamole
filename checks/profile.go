package checks

import (
	"guacamole/data"
	"guacamole/helpers"
)

func Profile() data.Check {
	checkResult := data.Check{
		Name:   "Layers' refresh time",
		Status: "✅",
		// RelatedGuidelines: "https://github.com/padok-team/docs-terraform-guidelines/blob/main/terraform/wysiwg_patterns.md",
		RelatedGuidelines: "http://bitly.ws/K5VV",
		Errors:            []string{},
	}

	layers, _ := helpers.GetLayers()
	for _, layer := range layers {
		err := layer.Init()
		if err != nil {
			panic(err)
		}
		err = layer.GetRefreshTime()
		if err != nil {
			panic(err)
		}
		if layer.RefreshTime > 120 {
			checkResult.Errors = append(checkResult.Errors, layer.Name)
		}
	}

	if len(checkResult.Errors) > 0 {
		checkResult.Status = "❌"
	}

	return checkResult
}
