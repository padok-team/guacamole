package checks

import (
	"fmt"
	"guacamole/data"
	"strconv"
	"strings"
	"sync"

	"github.com/fatih/color"
	"golang.org/x/exp/slices"
)

// TODO: Add total for codebase
// TODO: Move the data into the layer object

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

	codebaseTotal := data.Size{
		Resources:   0,
		Datasources: 0,
		Modules:     0,
	}

	// Display a legend
	fmt.Println("Legend:")
	c := color.New(color.FgWhite).Add(color.Bold)
	c.Println("  Module")
	c = color.New(color.FgBlue).Add(color.Bold)
	c.Println("  Datasource")
	c = color.New(color.FgGreen).Add(color.Bold)
	c.Println("  Resource")
	c = color.New(color.FgWhite)
	c.Printf("  Instance\n\n")
	c = color.New(color.FgYellow).Add(color.Bold)
	c.Println("Profiling by layer:")
	for _, layer := range layers {
		fmt.Println(strings.Repeat("-", 50))
		c = color.New(color.FgYellow).Add(color.Bold)
		c.Printf("%s\n", layer.Name)
		printTree(layer.RootModule, padding, verbose)

		codebaseTotal.Resources += layer.RootModule.CumulatedSize.Resources
		codebaseTotal.Datasources += layer.RootModule.CumulatedSize.Datasources
		codebaseTotal.Modules += layer.RootModule.CumulatedSize.Modules

		c = color.New(color.FgWhite).Add(color.Bold)
		c.Printf("\nTotal:\n")
		c = color.New(color.FgBlue).Add(color.Bold)
		c.Printf("  %-12s %d\n", "Datasources:", layer.RootModule.CumulatedSize.Datasources)
		c = color.New(color.FgGreen).Add(color.Bold)
		c.Printf("  %-12s %d\n", "Resources:", layer.RootModule.CumulatedSize.Resources)
		c := color.New(color.FgWhite).Add(color.Bold)
		c.Printf("  %-12s %d\n", "Modules:", layer.RootModule.CumulatedSize.Modules)
	}
	fmt.Println(strings.Repeat("-", 50))
	c = color.New(color.FgWhite).Add(color.Bold)
	c.Println("Codebase total:")
	c = color.New(color.FgBlue).Add(color.Bold)
	c.Printf("  %-12s %d\n", "Datasources:", codebaseTotal.Datasources)
	c = color.New(color.FgGreen).Add(color.Bold)
	c.Printf("  %-12s %d\n", "Resources:", codebaseTotal.Resources)
	c = color.New(color.FgWhite).Add(color.Bold)
	c.Printf("  %-12s %d\n", "Modules:", codebaseTotal.Modules)
}

func printTree(m data.Module, padding int, verbose bool) {
	c := color.New(color.FgWhite).Add(color.Bold)
	totalSizeString := ""
	if m.CumulatedSize.Resources+m.CumulatedSize.Datasources > 1 {
		totalSizeString = "[" + strconv.Itoa(m.CumulatedSize.Resources+m.CumulatedSize.Datasources) + "]"
	}
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
		countStr := ""
		if r.Count > 1 {
			countStr = "[" + strconv.Itoa(r.Count) + "]"
		}
		c.Printf("%*s %s\n", padding+2, countStr, r.Type)
		if verbose {
			for _, i := range r.Instances {
				c := color.New(color.FgWhite)
				countStr = ""
				if i.Count > 1 {
					countStr = "[" + strconv.Itoa(i.Count) + "]"
				}
				c.Printf("%*s %s\n", padding+4, countStr, i.Name)
			}
		}
	}
	for _, c := range m.Children {
		printTree(c, padding+2, verbose)
	}
}
