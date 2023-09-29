package checks

import (
	"fmt"
	"guacamole/data"
	"log"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"golang.org/x/exp/slices"
)

func Profile(codebase data.Codebase, verbose bool) {
	codebase.BuildLayers()
	codebase.ComputeStats()
	// Take the length of the longest number of objects in a layer for padding
	if len(codebase.Layers) == 0 {
		log.Println("No layers found")
		return
	}
	padding := len(strconv.Itoa(codebase.Layers[0].RootModule.Stats.CumulatedSize.Resources + codebase.Layers[0].RootModule.Stats.CumulatedSize.Datasources))
	if padding%2 == 0 {
		padding += 4
	} else {
		padding += 3
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
	c.Printf("  Datasource / Resource name\n\n")
	c = color.New(color.FgYellow).Add(color.Bold)
	c.Println("Profiling by layer:")
	for _, layer := range codebase.Layers {
		fmt.Println(strings.Repeat("-", 50))
		c = color.New(color.FgYellow).Add(color.Bold)
		c.Printf("%s\n", layer.Name)

		// If the layer has no object in state, we don't want to display it
		if layer.RootModule.Stats.CumulatedSize.Resources+layer.RootModule.Stats.CumulatedSize.Datasources > 0 {
			printTree(layer.RootModule, padding, verbose)
		} else {
			fmt.Println("  -- Layer has no object in state --")
		}

		c = color.New(color.FgWhite).Add(color.Bold)
		c.Printf("\nStats:\n")
		c = color.New(color.FgBlue).Add(color.Bold)
		c.Printf("  %s %d\n", "Datasources:", layer.RootModule.Stats.CumulatedSize.Datasources)
		c = color.New(color.FgBlue)
		c.Printf("    %-14s %d\n", "Distinct types:", len(layer.RootModule.Stats.DistinctDatasourceTypes))
		if verbose {
			for k, v := range layer.RootModule.Stats.DistinctDatasourceTypes {
				c.Printf("      [%d] %s\n", v, k)
			}
		}
		c = color.New(color.FgGreen).Add(color.Bold)
		c.Printf("  %s %d\n", "Resources:", layer.RootModule.Stats.CumulatedSize.Resources)
		c = color.New(color.FgGreen)
		c.Printf("    %-14s %d\n", "Distinct types:", len(layer.RootModule.Stats.DistinctResourceTypes))
		if verbose {
			for k, v := range layer.RootModule.Stats.DistinctResourceTypes {
				c.Printf("      [%d] %s\n", v, k)
			}
		}
		c = color.New(color.FgWhite).Add(color.Bold)
		c.Printf("  %s\n", "Modules:")
		c = color.New(color.FgWhite)
		c.Printf("    %-12s %d\n", "Module depth:", layer.RootModule.Stats.Depth)
	}

	fmt.Println(strings.Repeat("-", 50))
	c = color.New(color.FgWhite).Add(color.Bold)
	c.Println("Codebase stats:")
	c = color.New(color.FgBlue).Add(color.Bold)
	c.Printf("  %-12s %d\n", "Datasources:", codebase.Stats.Size.Datasources)
	c = color.New(color.FgBlue)
	c.Printf("    %-12s %d\n", "Distinct types:", len(codebase.Stats.DistinctDatasourceTypes))

	if verbose {
		for k, v := range codebase.Stats.DistinctDatasourceTypes {
			numberString := "[" + strconv.Itoa(v) + "]"
			c.Printf("%10s %s\n", numberString, k)
		}
	}
	c = color.New(color.FgGreen).Add(color.Bold)
	c.Printf("  %-12s %d\n", "Resources:", codebase.Stats.Size.Resources)
	c = color.New(color.FgGreen)
	c.Printf("    %-12s %d\n", "Distinct types:", len(codebase.Stats.DistinctResourceTypes))
	if verbose {
		for k, v := range codebase.Stats.DistinctResourceTypes {
			numberString := "[" + strconv.Itoa(v) + "]"
			c.Printf("%10s %s\n", numberString, k)
		}
	}
	c = color.New(color.FgWhite).Add(color.Bold)
	c.Printf("  %-12s %d\n", "Modules:", codebase.Stats.Size.Modules)
	c = color.New(color.FgWhite)
	c.Printf("    %-12s %d\n", "Max module depth:", codebase.Stats.Depth)
	// Show biggest layer
	c.Printf("    %s [%d] %s\n", "Biggest layer:", codebase.Stats.BiggestLayer.RootModule.Stats.CumulatedSize.Resources+codebase.Stats.BiggestLayer.RootModule.Stats.CumulatedSize.Datasources, codebase.Stats.BiggestLayer.Name)
	// Show biggest children
	c.Printf("    %s [%d] %s\n", "Biggest child module:", codebase.Stats.BiggestChildModule.Stats.CumulatedSize.Resources+codebase.Stats.BiggestChildModule.Stats.CumulatedSize.Datasources, codebase.Stats.BiggestChildModule.Address)
	// Show warnings
	c = color.New(color.FgYellow).Add(color.Bold)
	c.Printf("  %s\n", "Warnings:")
	for _, l := range codebase.Layers {
		l.ComputeWarnings()
		if len(l.Warnings.DatasourceInModuleWarning) > 0 {
			c = color.New(color.FgYellow)
			c.Printf("    Datasource(s) in module(s) in layer : %s\n", l.Name)
			for _, w := range l.Warnings.DatasourceInModuleWarning {
				c = color.New(color.FgWhite)
				c.Printf("      %s\n", w.Module.Address)
				for _, d := range w.Datasource {
					c = color.New(color.FgBlue)
					c.Printf("        %s\n", d.Type)
				}
			}
		}
	}
}

func printTree(m *data.Module, padding int, verbose bool) {
	c := color.New(color.FgWhite).Add(color.Bold)
	totalSizeString := "[" + strconv.Itoa(m.Stats.CumulatedSize.Resources+m.Stats.CumulatedSize.Datasources) + "]"
	c.Printf("%*s %s\n", padding, totalSizeString, m.Name)
	// Sort objects first by kind, then by name
	slices.SortFunc(m.ObjectTypes, func(i, j *data.ObjectType) int {
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
