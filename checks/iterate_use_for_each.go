package checks

import (
	"fmt"
	"guacamole/data"
	"regexp"
	"strconv"
	"sync"
)

type Change struct {
	Actions []string `json:"actions"`
}

type ResourceChange struct {
	Address      string `json:"address"`
	Mode         string `json:"mode"`
	Type         string `json:"type"`
	Name         string `json:"name"`
	Index        string `json:"index"`
	ProviderName string `json:"provider_name"`
	Change       Change `json:"change"`
}

type ResourceChanges struct {
	ResourceChanges []ResourceChange `json:"resource_changes"`
}

func IterateUseForEach(layers []*data.Layer) (data.Check, error) {
	dataCheck := data.Check{
		ID:                "TF_MOD_003",
		Name:              "Use for_each to create multiple resources of the same type",
		RelatedGuidelines: "https://padok-team.github.io/docs-terraform-guidelines/terraform/iterate_on_your_resources.html",
		Status:            "✅",
	}

	c := make(chan []string, len(layers))

	wg := new(sync.WaitGroup)
	wg.Add(len(layers))

	for i := range layers {
		go func(layer *data.Layer) {
			defer wg.Done()
			c <- checkLayer(layer)
		}(layers[i])
	}

	wg.Wait()

	// Wait for all goroutines to finish
	for i := 0; i < len(layers); i++ {
		checkErrors := <-c
		if len(checkErrors) > 0 {
			dataCheck.Status = "❌"
			dataCheck.Errors = append(dataCheck.Errors, checkErrors...)
		}
	}

	return dataCheck, nil
}

func checkLayer(layer *data.Layer) []string {
	var index int
	var indexString string
	var checkedResources, checkErrors []string

	// Analyze plan for resources with count > 1
	for _, rc := range layer.Plan.ResourceChanges {
		// Parse the module address to find numbers inside of []
		regexpIndexMatch := regexp.MustCompile(`\[(.*?)\]`).FindStringSubmatch(rc.Address)
		if len(regexpIndexMatch) > 0 {
			indexString = regexpIndexMatch[1]
		}

		// Ignore error in case of a string into the brackets, meaning that the resource was not created with count
		index, _ = strconv.Atoi(indexString)

		if index > 0 {
			// Check if the resource was already checked
			alreadyChecked := false
			for _, checkedResource := range checkedResources {
				if checkedResource == rc.ModuleAddress {
					alreadyChecked = true
					break
				}
			}
			if !alreadyChecked {
				checkedResources = append(checkedResources, rc.ModuleAddress)
				checkErrors = append(checkErrors, fmt.Sprintf("%s --> %s", layer.Name, rc.Address))
			}
		}
	}
	return checkErrors
}
