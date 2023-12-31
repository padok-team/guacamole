package helpers

import (
	"runtime"
	"sync"

	"github.com/padok-team/guacamole/data"
)

func ComputeLayers(withPlan bool) ([]*data.Layer, error) {
	layers, err := GetLayers()
	if err != nil {
		return nil, err
	}

	maxProcs := runtime.NumCPU()

	// fmt.Printf("Number of CPUs: %d, number of layers: %d, should be processed in %d batches\n", maxProcs, len(layers), int(math.Ceil(float64(len(layers))/float64(maxProcs))))

	// Channel to limit the number of goroutines running at the same time
	guard := make(chan struct{}, maxProcs)

	wg := new(sync.WaitGroup)

	for i := range layers {
		if layers[i].RootModule != nil {
			continue
		}
		wg.Add(1)
		// Add a struct to the channel to start a goroutine
		// If the channel is full, the goroutine will wait until another one finishes and removes the struct from the channel
		guard <- struct{}{}
		// fmt.Println("Processing layer: ", layers[i].Name)
		go func(layer *data.Layer) {
			defer wg.Done()
			layer.BuildRootModule()
			if withPlan {
				layer.ComputePlan()
			}
			// Remove the struct from the channel to allow another goroutine to start
			<-guard
			// fmt.Println("Finished processing layer: ", layer.Name)
		}(layers[i])
	}

	wg.Wait()

	return layers, nil
}
