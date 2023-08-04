package helpers

import (
	"guacamole/data"
	"sync"
)

func ComputeLayers(withPlan bool) ([]data.Layer, error) {
	layers, err := GetLayers()
	if err != nil {
		return nil, err
	}

	wg := new(sync.WaitGroup)
	wg.Add(len(layers))

	for i := range layers {
		go func(layer *data.Layer) {
			defer wg.Done()
			layer.ComputeState()
			if withPlan {
				layer.ComputePlan()
			}
		}(&layers[i])
	}

	wg.Wait()

	return layers, nil
}
