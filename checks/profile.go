package checks

import (
	"fmt"
	"guacamole/data"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/fatih/color"
	"golang.org/x/exp/slices"
)

func Profile(layers []data.Layer, verbose bool) {
	wg := new(sync.WaitGroup)

	wg.Add(len(layers))

	for i := range layers {
		go func(layer *data.Layer) {
			defer wg.Done()
			layer.BuildRootModule()
		}(&layers[i])
	}

	wg.Wait()

	padding := len(strconv.Itoa(layers[0].RootModule.CumulatedSize.Resources + layers[0].RootModule.CumulatedSize.Datasources))
	if padding%2 == 0 {
		padding += 4
	} else {
		padding += 3
	}

	codebaseSizes := data.Size{
		Resources:   0,
		Datasources: 0,
		Modules:     0,
	}

	codebaseStats := data.Stats{
		DistinctResourceTypes:   make(map[string]int),
		DistinctDatasourceTypes: make(map[string]int),
		Depth:                   0,
	}

	var biggestChild data.Module

	// Display a legend
	fmt.Println("Legend:")
	c := color.New(color.FgWhite).Add(color.Bold)
	c.Println("  Module")
	c = color.New(color.FgBlue).Add(color.Bold)
	c.Println("  Datasource")
	c = color.New(color.FgGreen).Add(color.Bold)
	c.Println("  Resource")
	c = color.New(color.FgWhite)
	c.Printf("  Datasource / Resource name\n\n")
	c = color.New(color.FgYellow).Add(color.Bold)
	c.Println("Profiling by layer:")
	for _, layer := range layers {
		fmt.Println(strings.Repeat("-", 50))
		c = color.New(color.FgYellow).Add(color.Bold)
		c.Printf("%s\n", layer.Name)

		if layer.RootModule.CumulatedSize.Resources+layer.RootModule.CumulatedSize.Datasources > 0 {
			printTree(layer.RootModule, padding, verbose)
		} else {
			fmt.Println("  -- Layer has no object in state --")
		}

		codebaseSizes.Resources += layer.RootModule.CumulatedSize.Resources
		codebaseSizes.Datasources += layer.RootModule.CumulatedSize.Datasources
		codebaseSizes.Modules += layer.RootModule.CumulatedSize.Modules

		stats := layer.RootModule.ComputeStats()

		// Update codebase stats
		for k, v := range stats.DistinctDatasourceTypes {
			codebaseStats.DistinctDatasourceTypes[k] += v
		}
		for k, v := range stats.DistinctResourceTypes {
			codebaseStats.DistinctResourceTypes[k] += v
		}

		for _, c := range layer.RootModule.Children {
			if c.CumulatedSize.Resources+c.CumulatedSize.Datasources > biggestChild.CumulatedSize.Resources+biggestChild.CumulatedSize.Datasources {
				biggestChild = c
			}
		}

		codebaseStats.Depth = int(math.Max(float64(codebaseStats.Depth), float64(stats.Depth)))

		dKeys := make([]string, 0, len(stats.DistinctDatasourceTypes))
		for k := range stats.DistinctDatasourceTypes {
			dKeys = append(dKeys, k)
		}

		sort.SliceStable(dKeys, func(i, j int) bool {
			if stats.DistinctDatasourceTypes[dKeys[i]] == stats.DistinctDatasourceTypes[dKeys[j]] {
				return strings.Compare(dKeys[i], dKeys[j]) < 0
			}
			return stats.DistinctDatasourceTypes[dKeys[i]] > stats.DistinctDatasourceTypes[dKeys[j]]
		})

		rKeys := make([]string, 0, len(stats.DistinctResourceTypes))
		for k := range stats.DistinctResourceTypes {
			rKeys = append(rKeys, k)
		}

		sort.SliceStable(rKeys, func(i, j int) bool {
			if stats.DistinctResourceTypes[rKeys[i]] == stats.DistinctResourceTypes[rKeys[j]] {
				return strings.Compare(rKeys[i], rKeys[j]) < 0
			}
			return stats.DistinctResourceTypes[rKeys[i]] > stats.DistinctResourceTypes[rKeys[j]]
		})

		c = color.New(color.FgWhite).Add(color.Bold)
		c.Printf("\nStats:\n")
		c = color.New(color.FgBlue).Add(color.Bold)
		c.Printf("  %s %d\n", "Datasources:", layer.RootModule.CumulatedSize.Datasources)
		c = color.New(color.FgBlue)
		c.Printf("    %-14s %d\n", "Distinct types:", len(stats.DistinctDatasourceTypes))
		if verbose {
			for _, k := range dKeys {
				c.Printf("      [%d] %s\n", stats.DistinctDatasourceTypes[k], k)
			}
		}
		c = color.New(color.FgGreen).Add(color.Bold)
		c.Printf("  %s %d\n", "Resources:", layer.RootModule.CumulatedSize.Resources)
		c = color.New(color.FgGreen)
		c.Printf("    %-14s %d\n", "Distinct types:", len(stats.DistinctResourceTypes))
		if verbose {
			for _, k := range rKeys {
				c.Printf("      [%d] %s\n", stats.DistinctResourceTypes[k], k)
			}
		}
		c = color.New(color.FgWhite).Add(color.Bold)
		c.Printf("  %s\n", "Modules:")
		c = color.New(color.FgWhite)
		c.Printf("    %-12s %d\n", "Module depth:", stats.Depth)
	}

	dgKeys := make([]string, 0, len(codebaseStats.DistinctDatasourceTypes))
	for k := range codebaseStats.DistinctDatasourceTypes {
		dgKeys = append(dgKeys, k)
	}

	sort.SliceStable(dgKeys, func(i, j int) bool {
		if codebaseStats.DistinctDatasourceTypes[dgKeys[i]] == codebaseStats.DistinctDatasourceTypes[dgKeys[j]] {
			return strings.Compare(dgKeys[i], dgKeys[j]) < 0
		}
		return codebaseStats.DistinctDatasourceTypes[dgKeys[i]] > codebaseStats.DistinctDatasourceTypes[dgKeys[j]]
	})

	rgKeys := make([]string, 0, len(codebaseStats.DistinctResourceTypes))
	for k := range codebaseStats.DistinctResourceTypes {
		rgKeys = append(rgKeys, k)
	}

	sort.SliceStable(rgKeys, func(i, j int) bool {
		if codebaseStats.DistinctResourceTypes[rgKeys[i]] == codebaseStats.DistinctResourceTypes[rgKeys[j]] {
			return strings.Compare(rgKeys[i], rgKeys[j]) < 0
		}
		return codebaseStats.DistinctResourceTypes[rgKeys[i]] > codebaseStats.DistinctResourceTypes[rgKeys[j]]
	})

	fmt.Println(strings.Repeat("-", 50))
	c = color.New(color.FgWhite).Add(color.Bold)
	c.Println("Codebase stats:")
	c = color.New(color.FgBlue).Add(color.Bold)
	c.Printf("  %-12s %d\n", "Datasources:", codebaseSizes.Datasources)
	c = color.New(color.FgBlue)
	c.Printf("    %-12s %d\n", "Distinct types:", len(codebaseStats.DistinctDatasourceTypes))
	if verbose {
		for _, k := range dgKeys {
			numberString := "[" + strconv.Itoa(codebaseStats.DistinctDatasourceTypes[k]) + "]"
			c.Printf("%10s %s\n", numberString, k)
		}
	}
	c = color.New(color.FgGreen).Add(color.Bold)
	c.Printf("  %-12s %d\n", "Resources:", codebaseSizes.Resources)
	c = color.New(color.FgGreen)
	c.Printf("    %-12s %d\n", "Distinct types:", len(codebaseStats.DistinctResourceTypes))
	if verbose {
		for _, k := range rgKeys {
			numberString := "[" + strconv.Itoa(codebaseStats.DistinctResourceTypes[k]) + "]"
			c.Printf("%10s %s\n", numberString, k)
		}
	}
	c = color.New(color.FgWhite).Add(color.Bold)
	c.Printf("  %-12s %d\n", "Modules:", codebaseSizes.Modules)
	c = color.New(color.FgWhite)
	c.Printf("    %-12s %d\n", "Max module depth:", codebaseStats.Depth)
	// Show biggest children
	c.Printf("    %s [%d] %s\n", "Biggest child module:", biggestChild.CumulatedSize.Resources+biggestChild.CumulatedSize.Datasources, biggestChild.Address)
}

func printTree(m data.Module, padding int, verbose bool) {
	c := color.New(color.FgWhite).Add(color.Bold)
	totalSizeString := "[" + strconv.Itoa(m.CumulatedSize.Resources+m.CumulatedSize.Datasources) + "]"
	c.Printf("%*s %s\n", padding, totalSizeString, m.Address)
	// Sort objects first by kind, then by name
	slices.SortFunc(m.ObjectTypes, func(i, j data.ObjectType) int {
		if i.Kind == j.Kind {
			return strings.Compare(i.Type, j.Type)
		}
		return strings.Compare(i.Kind, j.Kind)
	})

	for _, r := range m.ObjectTypes {
		c := color.New(color.FgGreen).Add(color.Bold)
		if r.Kind == "datasource" {
			c = color.New(color.FgBlue).Add(color.Bold)
		}
		countStr := "[" + strconv.Itoa(r.Count) + "]"
		c.Printf("%*s %s\n", padding+2, countStr, r.Type)
		if verbose {
			for _, i := range r.Instances {
				c := color.New(color.FgWhite)
				countStr = "[" + strconv.Itoa(i.Count) + "]"
				c.Printf("%*s %s\n", padding+4, countStr, i.Name)
			}
		}
	}
	for _, c := range m.Children {
		printTree(c, padding+2, verbose)
	}
}
