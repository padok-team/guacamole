package data

import (
	"math"
	"strings"

	tfjson "github.com/hashicorp/terraform-json"
	"golang.org/x/exp/slices"
)

type Module struct {
	Address string
	Name    string
	// Resources and datasources
	ObjectTypes []*ObjectType
	Children    []*Module
	Stats       ModuleStats
}

type ObjectType struct {
	Type      string
	Kind      string // resource or datasource
	Index     any    // Can be a string or an int
	Instances []Object
	Count     int
}

type Object struct {
	Name  string
	Count int
}

type ModuleStats struct {
	DistinctResourceTypes   map[string]int
	DistinctDatasourceTypes map[string]int
	Depth                   int
	Size                    Size
	CumulatedSize           Size
}

type Size struct {
	Resources   int
	Datasources int
	Modules     int
}

func (m *Module) buildModule(stateModule *tfjson.StateModule) {
	for _, c := range stateModule.ChildModules {
		module := Module{
			Address:     c.Address,
			Name:        c.Address,
			ObjectTypes: []*ObjectType{},
			Children:    []*Module{},
		}

		// If the module is not root, we want to keep only the module name
		if module.Name != "root" {
			module.Name = c.Address[strings.LastIndex(module.Name, ".")+1:]
		}

		module.buildModule(c)
		m.Children = append(m.Children, &module)
	}

	m.buildResourcesAndDatasources(stateModule)

	m.Stats.Size.Modules = len(m.Children)
	m.Stats.CumulatedSize.Modules += len(m.Children)

	for _, c := range m.Children {
		m.Stats.CumulatedSize.Resources += c.Stats.CumulatedSize.Resources
		m.Stats.CumulatedSize.Datasources += c.Stats.CumulatedSize.Datasources
		m.Stats.CumulatedSize.Modules += c.Stats.CumulatedSize.Modules
	}
}

func (m *Module) buildResourcesAndDatasources(state *tfjson.StateModule) {
	resourceCount := 0
	datasourceCount := 0

	for _, r := range state.Resources {
		kind := "resource"
		if r.Mode == "data" {
			kind = "datasource"
			datasourceCount++
		} else {
			resourceCount++
		}

		typeIndex := slices.IndexFunc(m.ObjectTypes, func(t *ObjectType) bool {
			return t.Type == r.Type
		})

		if typeIndex == -1 {
			m.ObjectTypes = append(m.ObjectTypes, &ObjectType{
				Type:  r.Type,
				Kind:  kind,
				Index: r.Index,
				Instances: []Object{
					{
						Name:  r.Name,
						Count: 1,
					},
				},
				Count: 1,
			})
		} else {
			instanceIndex := slices.IndexFunc(m.ObjectTypes[typeIndex].Instances, func(i Object) bool {
				return i.Name == r.Name
			})
			if instanceIndex == -1 {
				m.ObjectTypes[typeIndex].Instances = append(m.ObjectTypes[typeIndex].Instances, Object{
					Name:  r.Name,
					Count: 1,
				})
			} else {
				resource := m.ObjectTypes[typeIndex].Instances[instanceIndex]
				resource.Count++
				m.ObjectTypes[typeIndex].Instances[instanceIndex] = resource
			}
			m.ObjectTypes[typeIndex].Count++
		}
	}

	m.Stats.Size.Resources = resourceCount
	m.Stats.Size.Datasources = datasourceCount
	m.Stats.CumulatedSize.Resources = resourceCount
	m.Stats.CumulatedSize.Datasources = datasourceCount

}

func (m *Module) ComputeStats() {
	m.Stats.DistinctResourceTypes = map[string]int{}
	m.Stats.DistinctDatasourceTypes = map[string]int{}

	for _, o := range m.ObjectTypes {
		if o.Kind == "resource" {
			m.Stats.DistinctResourceTypes[o.Type] += o.Count
		} else {
			m.Stats.DistinctDatasourceTypes[o.Type] += o.Count
		}
	}

	for _, c := range m.Children {
		c.ComputeStats()
		m.Stats.Depth = int(math.Max(float64(m.Stats.Depth), float64(c.Stats.Depth+1)))
		for k, v := range c.Stats.DistinctResourceTypes {
			m.Stats.DistinctResourceTypes[k] += v
		}
		for k, v := range c.Stats.DistinctDatasourceTypes {
			m.Stats.DistinctDatasourceTypes[k] += v
		}
		m.Stats.Size.Resources += c.Stats.Size.Resources
		m.Stats.Size.Datasources += c.Stats.Size.Datasources
	}

	m.Stats.DistinctResourceTypes = SortMapByValueAndKey(m.Stats.DistinctResourceTypes)
	m.Stats.DistinctDatasourceTypes = SortMapByValueAndKey(m.Stats.DistinctDatasourceTypes)
}
