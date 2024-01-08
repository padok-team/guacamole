package checks

import (
	"encoding/json"
	"strconv"
	"sync"

	"github.com/padok-team/guacamole/data"
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
		ID:                "TF_MOD_004",
		Name:              "Use for_each to create multiple resources of the same type",
		RelatedGuidelines: "https://padok-team.github.io/docs-terraform-guidelines/terraform/iterate_on_your_resources.html",
		Status:            "✅",
	}

	var checkableLayers []*data.Layer

	var allCheckErrors []data.Error

	for _, layer := range layers {
		if layer.RootModule != nil {
			checkableLayers = append(checkableLayers, layer)
		}
	}

	c := make(chan []data.Error, len(checkableLayers))

	wg := new(sync.WaitGroup)
	wg.Add(len(checkableLayers))

	for i := range checkableLayers {
		go func(layer *data.Layer) {
			defer wg.Done()
			c <- checkModules(layer.Name, layer.RootModule)
		}(checkableLayers[i])
	}

	wg.Wait()

	// Wait for all goroutines to finish
	for i := 0; i < len(checkableLayers); i++ {
		checkErrors := <-c
		if len(checkErrors) > 0 {
			dataCheck.Status = "❌"
			allCheckErrors = append(allCheckErrors, checkErrors...)
		}
	}

	dataCheck.Errors = allCheckErrors

	return dataCheck, nil
}

func checkModules(layerAddress string, m *data.Module) []data.Error {
	var checkErrors []data.Error

	if len(m.ObjectTypes) > 0 {
		for _, o := range m.ObjectTypes {
			// Check if type of o.Index is int or string
			switch o.Index.(type) {
			// If it's an int, we can assume that the resource was created with count
			case json.Number:
				if o.Count > 1 {
					checkErrors = append(checkErrors, data.Error{
						Path:        layerAddress,
						LineNumber:  -1,
						Description: "[module] " + m.Address + " --> [resource] " + o.Type + " (" + strconv.Itoa(o.Count) + ")",
					})

				}
			}
		}
	}

	for _, c := range m.Children {
		checkErrors = append(checkErrors, checkModules(layerAddress, c)...)
	}

	return checkErrors
}
