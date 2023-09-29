package data

import (
	"math"
	"sort"
	"strings"
	"sync"
)

type Codebase struct {
	Layers   []*Layer
	Stats    CodebaseStats
	Warnings Warnings
}

type CodebaseStats struct {
	BiggestLayer            *Layer
	BiggestChildModule      *Module
	DistinctResourceTypes   map[string]int
	DistinctDatasourceTypes map[string]int
	Depth                   int
	Size                    Size
}

func (c *Codebase) BuildLayers() {
	wg := new(sync.WaitGroup)

	wg.Add(len(c.Layers))

	for i := range c.Layers {
		go func(layer *Layer) {
			defer wg.Done()
			layer.BuildRootModule()
		}(c.Layers[i])
	}

	wg.Wait()
}

func (c *Codebase) ComputeStats() {
	distinctDatasourceTypes, distinctResourceTypes := map[string]int{}, map[string]int{}

	for _, l := range c.Layers {
		// Compute all module stats
		l.RootModule.ComputeStats()

		// Compute size
		c.Stats.Size.Resources += l.RootModule.Stats.CumulatedSize.Resources
		c.Stats.Size.Datasources += l.RootModule.Stats.CumulatedSize.Datasources
		c.Stats.Size.Modules += l.RootModule.Stats.CumulatedSize.Modules

		// Compute distinct types for each layer and submodules
		// FIXME: This is not correct, we should only count distinct types for the current layer
		for k, v := range l.RootModule.Stats.DistinctDatasourceTypes {
			distinctDatasourceTypes[k] += v
		}
		for k, v := range l.RootModule.Stats.DistinctResourceTypes {
			distinctResourceTypes[k] += v
		}

		for _, c := range l.RootModule.Children {
			for k, v := range c.Stats.DistinctDatasourceTypes {
				distinctDatasourceTypes[k] += v
			}
			for k, v := range c.Stats.DistinctResourceTypes {
				distinctResourceTypes[k] += v
			}
		}

		// Compute depth
		c.Stats.Depth = int(math.Max(float64(c.Stats.Depth), float64(l.RootModule.Stats.Depth)))

		// Compute biggest layer
		if c.Stats.BiggestLayer == nil {
			c.Stats.BiggestLayer = l
		} else if c.Stats.BiggestLayer.RootModule.Stats.CumulatedSize.Resources+c.Stats.BiggestLayer.RootModule.Stats.CumulatedSize.Datasources < l.RootModule.Stats.CumulatedSize.Resources+l.RootModule.Stats.CumulatedSize.Datasources {
			c.Stats.BiggestLayer = l
		}

		// Compute biggest child module
		for _, m := range l.RootModule.Children {
			if c.Stats.BiggestChildModule == nil {
				c.Stats.BiggestChildModule = m
			} else if c.Stats.BiggestChildModule.Stats.CumulatedSize.Resources+c.Stats.BiggestChildModule.Stats.CumulatedSize.Datasources < m.Stats.CumulatedSize.Resources+m.Stats.CumulatedSize.Datasources {
				c.Stats.BiggestChildModule = m
			}
		}
	}

	c.Stats.DistinctResourceTypes = SortMapByValueAndKey(distinctResourceTypes)
	c.Stats.DistinctDatasourceTypes = SortMapByValueAndKey(distinctDatasourceTypes)
}

func SortMapByValueAndKey(m map[string]int) map[string]int {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	// Sort first by value, then by key
	sort.SliceStable(keys, func(i, j int) bool {
		if m[keys[i]] == m[keys[j]] {
			return strings.Compare(keys[i], keys[j]) < 0
		}
		return m[keys[i]] > m[keys[j]]
	})

	sortedMap := map[string]int{}
	for _, k := range keys {
		sortedMap[k] = m[k]
	}

	return sortedMap
}
