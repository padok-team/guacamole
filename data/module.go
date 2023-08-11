package data

import (
	"strings"

	tfjson "github.com/hashicorp/terraform-json"
	"golang.org/x/exp/slices"
)

type Module struct {
	Address string
	// Resources and datasources
	ObjectTypes   []ObjectType
	Size          Size
	CumulatedSize Size
	Children      []Module
}

type ObjectType struct {
	Type      string
	Kind      string // resource or datasource
	Instances []Object
	Count     int
}

type Object struct {
	Name  string
	Count int
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
			ObjectTypes: []ObjectType{},
			Children:    []Module{},
		}

		// If the module is not root, we want to keep only the module name
		if module.Address != "root" {
			module.Address = c.Address[strings.LastIndex(module.Address, ".")+1:]
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
		kind := "resource"
		if r.Mode == "data" {
			kind = "datasource"
			datasourceCount++
		} else {
			resourceCount++
		}

		typeIndex := slices.IndexFunc(m.ObjectTypes, func(t ObjectType) bool {
			return t.Type == r.Type
		})

		if typeIndex == -1 {
			m.ObjectTypes = append(m.ObjectTypes, ObjectType{
				Type: r.Type,
				Kind: kind,
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

	m.Size.Resources = resourceCount
	m.Size.Datasources = datasourceCount
	m.CumulatedSize.Resources = resourceCount
	m.CumulatedSize.Datasources = datasourceCount
}
