package checks

import (
	"fmt"
	"guacamole/data"
	"sort"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

type ResourceType struct {
	Instances map[string]int
	Total     int
}

type DatasourceType struct {
	Instances map[string]int
	Total     int
}

func Profile(layers []data.Layer, verbose bool) {
	c := color.New(color.FgYellow).Add(color.Bold)
	c.Println("Profiling by layer:")
	fmt.Println(strings.Repeat("-", 50))
	totalResources := 0
	totalDatasources := 0
	for _, layer := range layers {
		resourceTypes := make(map[string]ResourceType)
		datasourceTypes := make(map[string]DatasourceType)
		c := color.New(color.FgYellow).Add(color.Bold)

		for _, resource := range layer.Plan.Config.RootModule.Resources {
			if resource.Mode == "managed" {
				if _, ok := resourceTypes[resource.Type]; !ok {
					resourceTypes[resource.Type] = ResourceType{
						Instances: map[string]int{resource.Name: 1},
						Total:     1,
					}
				} else {
					resourceType := resourceTypes[resource.Type]
					resourceType.Instances[resource.Name]++
					resourceType.Total++
					resourceTypes[resource.Type] = resourceType
				}
				totalResources += len(resourceTypes)
			} else {
				if _, ok := datasourceTypes[resource.Type]; !ok {
					datasourceTypes[resource.Type] = DatasourceType{
						Instances: map[string]int{resource.Name: 1},
						Total:     1,
					}
				} else {
					datasourceType := datasourceTypes[resource.Type]
					datasourceType.Instances[resource.Name]++
					datasourceType.Total++
					datasourceTypes[resource.Type] = datasourceType
				}
				totalDatasources += len(datasourceTypes)
			}
		}

		resourceCountStr := "[" + strconv.Itoa(len(layer.Plan.Config.RootModule.Resources)) + "]"

		// Compute padding value to be the closest even number to the length of the resourceCountStr
		padding := 1
		if len(resourceCountStr)%2 != 0 {
			padding += len(resourceCountStr) + 1
		} else {
			padding += len(resourceCountStr)
		}

		c.Println(layer.Name)

		// Sort the resourceTypeKeys
		resourceTypeKeys := make([]string, 0, len(resourceTypes))
		for k := range resourceTypes {
			resourceTypeKeys = append(resourceTypeKeys, k)
		}

		// Sort the datasourceTypeKeys
		datasourceTypeKeys := make([]string, 0, len(datasourceTypes))
		for k := range datasourceTypes {
			datasourceTypeKeys = append(datasourceTypeKeys, k)
		}

		sort.Strings(resourceTypeKeys)

		if len(resourceTypes) > 0 {
			totalLayerResources := 0
			for _, r := range resourceTypes {
				totalLayerResources += r.Total
			}
			resourceTypeCountStr := "[" + strconv.Itoa(totalLayerResources) + "]"
			resourceTypeDistinctCountStr := "[" + strconv.Itoa(len(resourceTypes)) + "]"
			c = color.New(color.FgWhite).Add(color.Bold)
			c.Printf("%*s %s of %s different types\n", padding, resourceTypeCountStr, "Resources", resourceTypeDistinctCountStr)
			for _, k := range resourceTypeKeys {
				c = color.New(color.FgWhite)
				if verbose {
					c.Add(color.Bold)
				}
				totalStr := " "
				if len(resourceTypes[k].Instances) > 1 {
					totalStr = "[" + strconv.Itoa(resourceTypes[k].Total) + "]"
				}
				c.Printf("%*s %s\n", padding+2, totalStr, k)
				if verbose {
					instanceKeys := make([]string, 0, len(resourceTypes[k].Instances))
					for k := range resourceTypes[k].Instances {
						instanceKeys = append(instanceKeys, k)
					}
					sort.Strings(instanceKeys)
					c = color.New(color.FgWhite)
					for _, instance := range instanceKeys {
						instanceStr := " "
						if resourceTypes[k].Instances[instance] > 1 {
							instanceStr = "[" + strconv.Itoa(resourceTypes[k].Instances[instance]) + "]"
						}
						c.Printf("%*s %s\n", padding+4, instanceStr, instance)
					}
				}
			}
		}

		if len(datasourceTypes) > 0 {
			totalLayerDatsources := 0
			for _, r := range resourceTypes {
				totalLayerDatsources += r.Total
			}
			datasourceTypeCountStr := "[" + strconv.Itoa(totalLayerDatsources) + "]"
			datasourceTypeDistinctCountStr := "[" + strconv.Itoa(len(datasourceTypes)) + "]"
			c = color.New(color.FgWhite).Add(color.Bold)
			c.Printf("%*s %s of %s different types\n", padding, datasourceTypeCountStr, "Datasources", datasourceTypeDistinctCountStr)
			for _, k := range datasourceTypeKeys {
				c = color.New(color.FgWhite)
				if verbose {
					c.Add(color.Bold)
				}
				totalStr := " "
				if len(datasourceTypes[k].Instances) > 1 {
					totalStr = "[" + strconv.Itoa(datasourceTypes[k].Total) + "]"
				}
				c.Printf("%*s %s\n", padding+2, totalStr, k)
				if verbose {
					instanceKeys := make([]string, 0, len(datasourceTypes[k].Instances))
					for k := range datasourceTypes[k].Instances {
						instanceKeys = append(instanceKeys, k)
					}
					sort.Strings(instanceKeys)
					c = color.New(color.FgWhite)
					for _, instance := range instanceKeys {
						instanceStr := " "
						if datasourceTypes[k].Instances[instance] > 1 {
							instanceStr = "[" + strconv.Itoa(datasourceTypes[k].Instances[instance]) + "]"
						}
						c.Printf("%*s %s\n", padding+4, instanceStr, instance)
					}
				}
			}
		}
		fmt.Println(strings.Repeat("-", 50))
	}

	c = color.New(color.FgYellow).Add(color.Bold)
	c.Println("Total:")
	c = color.New(color.FgWhite).Add(color.Bold)
	c.Printf("  Resources:   %d\n", totalResources)
	c.Printf("  Datasources: %d\n", totalDatasources)
	c.Printf("  Overall:     %d\n", totalResources+totalDatasources)
}
