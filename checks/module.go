package checks

import (
	"fmt"
	"guacamole/data"
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

func ProviderInModule() data.Check {
	name := "No provider in module"
	relatedGuidelines := "https://github.com/padok-team/docs-terraform-guidelines/blob/main/terraform/donts.md#using-provider-block-in-modules"
	fmt.Println("Checking none prescence provider in module...")
	// Find recusively all the modules in the current directory
	modules, err := getModules()
	if err != nil {
		fmt.Println("Error:", err)
	}
	modulesInError := []Module{}
	// For each module, check if the provider is defined
	for _, module := range modules {
		moduleConf, diags := tfconfig.LoadModule(module.FullPath)
		if diags.HasErrors() {
			fmt.Println("Error:", diags)
		}
		//If the module has no provider, display an error
		if len(moduleConf.ProviderConfigs) > 0 {
			fmt.Println("Error: provider found")
			fmt.Println("Module:", module.FullPath)
			modulesInError = append(modulesInError, module)
		}

	}
	if len(modulesInError) > 0 {
		return data.Check{
			Name:              name,
			RelatedGuidelines: relatedGuidelines,
			Status:            "❌",
		}
	}
	return data.Check{
		Name:              name,
		RelatedGuidelines: relatedGuidelines,
		Status:            "✅",
	}
}

type Module struct {
	Name     string
	FullPath string
}

func getModules() ([]Module, error) {
	root := "/Users/benjaminsanvoisin/dev/wizzair/aws-network/modules"
	modules := []Module{}
	//Get all subdirectories in root path
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println("Error:", err)
		}
		if info.IsDir() && path != root {
			modules = append(modules, Module{Name: info.Name(), FullPath: path})
		}
		return nil
	})
	return modules, err
}
