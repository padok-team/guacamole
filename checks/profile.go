package checks

import (
	"fmt"
	"guacamole/data"
	"strconv"
	"strings"
	"sync"

	"github.com/fatih/color"
	tfjson "github.com/hashicorp/terraform-json"
	"golang.org/x/exp/slices"
)

type Module struct {
	Address         string
	ResourceTypes   []ResourceType
	DatasourceTypes []DatasourceType
	Size            Size
	CumulatedSize   Size
	Children        []Module
}

type ResourceType struct {
	Type      string
	Instances []Resource
	Count     int
}

type DatasourceType struct {
	Type      string
	Instances []Datasource
	Count     int
}

type Resource struct {
	Name  string
	Count int
}

type Datasource struct {
	Name  string
	Count int
}

type Size struct {
	Resources   int
	Datasources int
	Modules     int
}

// TODO: Add total for codebase
// TODO: Move the data into the layer object

func Profile(layers []data.Layer, verbose bool) {
	states := []Module{}

	channel := make(chan Module, len(layers))
	defer close(channel)

	wg := new(sync.WaitGroup)

	for _, layer := range layers {
		wg.Add(1)
		go func(layer data.Layer) {
			defer wg.Done()
			state := buildState(layer)
			channel <- state
		}(layer)
	}

	wg.Wait()

	for range layers {
		states = append(states, <-channel)
	}

	padding := len(strconv.Itoa(states[0].CumulatedSize.Resources + states[0].CumulatedSize.Datasources))
	if padding%2 == 0 {
		padding += 4
	} else {
		padding += 3
	}

	c := color.New(color.FgYellow).Add(color.Bold)
	c.Println("Profiling by layer:")
	for _, state := range states {
		fmt.Println(strings.Repeat("-", 50))
		state.PrintTree(padding, verbose)
	}
}

func buildState(layer data.Layer) Module {
	state := Module{
		Address:         "root",
		ResourceTypes:   []ResourceType{},
		DatasourceTypes: []DatasourceType{},
		Children:        []Module{},
	}

	state.buildModule(layer.State.Values.RootModule)

	return state
}

func (m *Module) buildModule(stateModule *tfjson.StateModule) {
	for _, c := range stateModule.ChildModules {
		module := Module{
			Address:         c.Address,
			ResourceTypes:   []ResourceType{},
			DatasourceTypes: []DatasourceType{},
			Children:        []Module{},
		}

		// If the module is not root, we want to keep only the module name
		if module.Address != "root" {
			module.Address = c.Address[strings.LastIndex(module.Address, "module."):]
		}

		module.buildModule(c)
		m.Children = append(m.Children, module)
	}

	m.buildResourcesAndDatasources(stateModule)

	m.Size.Modules = len(m.Children)
	m.CumulatedSize.Modules += len(m.Children)

	for _, c := range m.Children {
		m.CumulatedSize.Resources += c.CumulatedSize.Resources
		m.CumulatedSize.Datasources += c.CumulatedSize.Datasources
		m.CumulatedSize.Modules += c.CumulatedSize.Modules
	}
}

func (m *Module) buildResourcesAndDatasources(state *tfjson.StateModule) {
	resourceCount := 0
	datasourceCount := 0

	for _, r := range state.Resources {
		// spew.Dump(r)
		if r.Mode == "managed" {
			typeIndex := slices.IndexFunc(m.ResourceTypes, func(t ResourceType) bool {
				return t.Type == r.Type
			})

			if typeIndex == -1 {
				m.ResourceTypes = append(m.ResourceTypes, ResourceType{
					Type: r.Type,
					Instances: []Resource{
						{
							Name:  r.Name,
							Count: 1,
						},
					},
					Count: 1,
				})
			} else {
				instanceIndex := slices.IndexFunc(m.ResourceTypes[typeIndex].Instances, func(i Resource) bool {
					return i.Name == r.Name
				})
				if instanceIndex == -1 {
					m.ResourceTypes[typeIndex].Instances = append(m.ResourceTypes[typeIndex].Instances, Resource{
						Name:  r.Name,
						Count: 1,
					})
				} else {
					resource := m.ResourceTypes[typeIndex].Instances[instanceIndex]
					resource.Count++
					m.ResourceTypes[typeIndex].Instances[instanceIndex] = resource
				}
				m.ResourceTypes[typeIndex].Count++
			}
			resourceCount++
		} else {
			typeIndex := slices.IndexFunc(m.DatasourceTypes, func(t DatasourceType) bool {
				return t.Type == r.Type
			})

			if typeIndex == -1 {
				m.DatasourceTypes = append(m.DatasourceTypes, DatasourceType{
					Type: r.Type,
					Instances: []Datasource{
						{
							Name:  r.Name,
							Count: 1,
						},
					},
					Count: 1,
				})
			} else {
				instanceIndex := slices.IndexFunc(m.DatasourceTypes[typeIndex].Instances, func(i Datasource) bool {
					return i.Name == r.Name
				})
				if instanceIndex == -1 {
					m.DatasourceTypes[typeIndex].Instances = append(m.DatasourceTypes[typeIndex].Instances, Datasource{
						Name:  r.Name,
						Count: 1,
					})
				} else {
					resource := m.DatasourceTypes[typeIndex].Instances[instanceIndex]
					resource.Count++
					m.DatasourceTypes[typeIndex].Instances[instanceIndex] = resource
				}
				m.DatasourceTypes[typeIndex].Count++
			}
			datasourceCount++
		}
	}

	m.Size.Resources = resourceCount
	m.Size.Datasources = datasourceCount
	m.CumulatedSize.Resources = resourceCount
	m.CumulatedSize.Datasources = datasourceCount
}

func (m *Module) PrintTree(padding int, verbose bool) {
	c := color.New(color.FgWhite).Add(color.Bold)
	totalSizeString := ""
	if m.CumulatedSize.Resources+m.CumulatedSize.Datasources > 1 {
		totalSizeString = "[" + strconv.Itoa(m.CumulatedSize.Resources+m.CumulatedSize.Datasources) + "]"
	}
	c.Printf("%*s %s\n", padding, totalSizeString, m.Address)
	if verbose {
		// Sort resources and datasources by name
		slices.SortFunc(m.ResourceTypes, func(i, j ResourceType) int {
			return strings.Compare(i.Type, j.Type)
		})

		for _, r := range m.ResourceTypes {
			c := color.New(color.FgGreen).Add(color.Bold)
			countStr := ""
			if r.Count > 1 {
				countStr = "[" + strconv.Itoa(r.Count) + "]"
			}
			c.Printf("%*s %s\n", padding+2, countStr, r.Type)
			for _, i := range r.Instances {
				c := color.New(color.FgWhite)
				countStr = ""
				if i.Count > 1 {
					countStr = "[" + strconv.Itoa(i.Count) + "]"
				}
				c.Printf("%*s %s\n", padding+4, countStr, i.Name)
			}
		}

		slices.SortFunc(m.DatasourceTypes, func(i, j DatasourceType) int {
			return strings.Compare(i.Type, j.Type)
		})

		for _, d := range m.DatasourceTypes {
			c := color.New(color.FgBlue).Add(color.Bold)
			countStr := ""

			if d.Count > 1 {
				countStr = "[" + strconv.Itoa(d.Count) + "]"
			}
			c.Printf("%*s %s\n", padding+2, countStr, d.Type)
			for _, i := range d.Instances {
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
		c.PrintTree(padding+2, verbose)
	}
}

// func ProfileLayer(layers []data.Layer, verbose bool) {
// 	c := color.New(color.FgYellow).Add(color.Bold)
// 	c.Println("Profiling by layer:")
// 	fmt.Println(strings.Repeat("-", 50))
// 	totalResources := 0
// 	totalDatasources := 0
// 	for _, layer := range layers {
// 		resourceTypes := make(map[string]ResourceType)
// 		datasourceTypes := make(map[string]DatasourceType)
// 		c := color.New(color.FgYellow).Add(color.Bold)

// 		spew.Dump(layer.State)

// 		for _, resource := range layer.Plan.Config.RootModule.Resources {
// 			if resource.Mode == "managed" {
// 				if _, ok := resourceTypes[resource.Type]; !ok {
// 					resourceTypes[resource.Type] = ResourceType{
// 						Instances: map[string]int{resource.Name: 1},
// 						Total:     1,
// 					}
// 				}
// 				resourceType := resourceTypes[resource.Type]
// 				resourceType.Instances[resource.Name]++
// 				resourceType.Total++
// 				resourceTypes[resource.Type] = resourceType
// 				totalResources += len(resourceTypes)
// 			} else {
// 				if _, ok := datasourceTypes[resource.Type]; !ok {
// 					datasourceTypes[resource.Type] = DatasourceType{
// 						Instances: map[string]int{resource.Name: 1},
// 						Total:     1,
// 					}
// 				}
// 				datasourceType := datasourceTypes[resource.Type]
// 				datasourceType.Instances[resource.Name]++
// 				datasourceType.Total++
// 				datasourceTypes[resource.Type] = datasourceType
// 				totalDatasources += len(datasourceTypes)
// 			}
// 		}

// 		resourceCountStr := "[" + strconv.Itoa(len(layer.Plan.Config.RootModule.Resources)) + "]"

// 		// Compute padding value to be the closest even number to the length of the resourceCountStr
// 		padding := 1
// 		if len(resourceCountStr)%2 != 0 {
// 			padding += len(resourceCountStr) + 1
// 		} else {
// 			padding += len(resourceCountStr)
// 		}

// 		c.Println(layer.Name)

// 		// Sort the resourceTypeKeys
// 		resourceTypeKeys := make([]string, 0, len(resourceTypes))
// 		for k := range resourceTypes {
// 			resourceTypeKeys = append(resourceTypeKeys, k)
// 		}

// 		// Sort the datasourceTypeKeys
// 		datasourceTypeKeys := make([]string, 0, len(datasourceTypes))
// 		for k := range datasourceTypes {
// 			datasourceTypeKeys = append(datasourceTypeKeys, k)
// 		}

// 		sort.Strings(resourceTypeKeys)

// 		if len(resourceTypes) > 0 {
// 			totalLayerResources := 0
// 			for _, r := range resourceTypes {
// 				totalLayerResources += r.Total
// 			}
// 			resourceTypeCountStr := "[" + strconv.Itoa(totalLayerResources) + "]"
// 			resourceTypeDistinctCountStr := "[" + strconv.Itoa(len(resourceTypes)) + "]"
// 			c = color.New(color.FgWhite).Add(color.Bold)
// 			c.Printf("%*s %s of %s different types\n", padding, resourceTypeCountStr, "Resources", resourceTypeDistinctCountStr)
// 			for _, k := range resourceTypeKeys {
// 				c = color.New(color.FgWhite)
// 				if verbose {
// 					c.Add(color.Bold)
// 				}
// 				totalStr := " "
// 				if len(resourceTypes[k].Instances) > 1 {
// 					totalStr = "[" + strconv.Itoa(resourceTypes[k].Total) + "]"
// 				}
// 				c.Printf("%*s %s\n", padding+2, totalStr, k)
// 				if verbose {
// 					instanceKeys := make([]string, 0, len(resourceTypes[k].Instances))
// 					for k := range resourceTypes[k].Instances {
// 						instanceKeys = append(instanceKeys, k)
// 					}
// 					sort.Strings(instanceKeys)
// 					c = color.New(color.FgWhite)
// 					for _, instance := range instanceKeys {
// 						instanceStr := " "
// 						if resourceTypes[k].Instances[instance] > 1 {
// 							instanceStr = "[" + strconv.Itoa(resourceTypes[k].Instances[instance]) + "]"
// 						}
// 						c.Printf("%*s %s\n", padding+4, instanceStr, instance)
// 					}
// 				}
// 			}
// 		}

// 		if len(datasourceTypes) > 0 {
// 			totalLayerDatsources := 0
// 			for _, r := range resourceTypes {
// 				totalLayerDatsources += r.Total
// 			}
// 			datasourceTypeCountStr := "[" + strconv.Itoa(totalLayerDatsources) + "]"
// 			datasourceTypeDistinctCountStr := "[" + strconv.Itoa(len(datasourceTypes)) + "]"
// 			c = color.New(color.FgWhite).Add(color.Bold)
// 			c.Printf("%*s %s of %s different types\n", padding, datasourceTypeCountStr, "Datasources", datasourceTypeDistinctCountStr)
// 			for _, k := range datasourceTypeKeys {
// 				c = color.New(color.FgWhite)
// 				if verbose {
// 					c.Add(color.Bold)
// 				}
// 				totalStr := " "
// 				if len(datasourceTypes[k].Instances) > 1 {
// 					totalStr = "[" + strconv.Itoa(datasourceTypes[k].Total) + "]"
// 				}
// 				c.Printf("%*s %s\n", padding+2, totalStr, k)
// 				if verbose {
// 					instanceKeys := make([]string, 0, len(datasourceTypes[k].Instances))
// 					for k := range datasourceTypes[k].Instances {
// 						instanceKeys = append(instanceKeys, k)
// 					}
// 					sort.Strings(instanceKeys)
// 					c = color.New(color.FgWhite)
// 					for _, instance := range instanceKeys {
// 						instanceStr := " "
// 						if datasourceTypes[k].Instances[instance] > 1 {
// 							instanceStr = "[" + strconv.Itoa(datasourceTypes[k].Instances[instance]) + "]"
// 						}
// 						c.Printf("%*s %s\n", padding+4, instanceStr, instance)
// 					}
// 				}
// 			}
// 		}
// 		fmt.Println(strings.Repeat("-", 50))
// 	}

// 	c = color.New(color.FgYellow).Add(color.Bold)
// 	c.Println("Total:")
// 	c = color.New(color.FgWhite).Add(color.Bold)
// 	c.Printf("  Resources:   %d\n", totalResources)
// 	c.Printf("  Datasources: %d\n", totalDatasources)
// 	c.Printf("  Overall:     %d\n", totalResources+totalDatasources)
// }
