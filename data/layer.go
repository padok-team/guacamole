package data

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	tfjson "github.com/hashicorp/terraform-json"
)

type Layer struct {
	Name       string
	FullPath   string
	Plan       *tfjson.Plan
	State      *tfjson.State
	RootModule *Module
	Warnings   Warnings
}

type Warnings struct {
	DatasourceInModuleWarning []datasourceInModuleWarning
	ModuleDepthWarning        []moduleDepthWarning
}

type datasourceInModuleWarning struct {
	Module     *Module
	Datasource []ObjectType
}

type moduleDepthWarning struct {
	Module Module
}

func (l *Layer) ComputePlan() {
	dirPath := "/tmp/" + strings.ReplaceAll(l.Name, "/", "_")
	planFile := dirPath + "_plan.json"

	_, err := os.Stat(filepath.Join(l.FullPath, ".terragrunt-cache"))
	if os.IsNotExist(err) {
		initCmd := exec.CommandContext(context.Background(), "terragrunt", "init")
		initCmd.Dir = l.FullPath
		if out, initErr := initCmd.CombinedOutput(); initErr != nil {
			log.Error(string(out))
			os.Exit(1)
		}
	}

	// Don't lock the state file while running the plan
	planCmd := exec.CommandContext(context.Background(), "terragrunt", "plan", "-out="+planFile, "-lock=false")
	planCmd.Dir = l.FullPath
	if out, planErr := planCmd.CombinedOutput(); planErr != nil {
		log.Info("failed to create plan: ", string(out))
	}

	// Create JSON plan
	showCmd := exec.CommandContext(context.Background(), "terragrunt", "show", "-json", planFile)
	showCmd.Dir = l.FullPath
	out, err := showCmd.Output()
	if err != nil {
		log.Info("failed to create JSON plan: ", err)
		return
	}

	var plan tfjson.Plan
	if err := json.Unmarshal(out, &plan); err != nil {
		log.Info("failed to parse JSON plan: ", err)
		return
	}

	l.Plan = &plan
}

func (l *Layer) ComputeState() {
	_, err := os.Stat(filepath.Join(l.FullPath, ".terragrunt-cache"))
	if os.IsNotExist(err) {
		initCmd := exec.CommandContext(context.Background(), "terragrunt", "init")
		initCmd.Dir = l.FullPath
		if out, initErr := initCmd.CombinedOutput(); initErr != nil {
			log.Error(string(out))
			os.Exit(1)
		}
	}

	// Create Terraform state file
	showCmd := exec.CommandContext(context.Background(), "terragrunt", "show", "-json")
	showCmd.Dir = l.FullPath
	out, err := showCmd.Output()
	if err != nil {
		log.Infof("failed to create state: %s", err)
		return
	}

	var state tfjson.State
	if err := json.Unmarshal(out, &state); err != nil {
		log.Infof("failed to parse state: %s", err)
		return
	}

	l.State = &state
}

func (layer *Layer) BuildRootModule() {
	if layer.State == nil {
		layer.ComputeState()
	}

	if layer.State != nil && layer.State.Values != nil {
		layer.RootModule = &Module{
			Address:     "root",
			Name:        "root",
			ObjectTypes: []*ObjectType{},
			Children:    []*Module{},
		}

		layer.RootModule.buildModule(layer.State.Values.RootModule)
	}
}

func (layer *Layer) ComputeWarnings() {
	if layer.RootModule == nil {
		layer.BuildRootModule()
	}

	if layer.RootModule != nil {
		layer.Warnings.DatasourceInModuleWarning = computeDatasourceInModuleWarning(layer.RootModule)
	}
}

func computeDatasourceInModuleWarning(module *Module) []datasourceInModuleWarning {
	datasourceInModuleWarnings := []datasourceInModuleWarning{}

	for _, r := range module.ObjectTypes {
		if r.Kind == "datasource" {
			if len(datasourceInModuleWarnings) == 0 {
				datasourceInModuleWarnings = append(datasourceInModuleWarnings, datasourceInModuleWarning{
					Module:     module,
					Datasource: []ObjectType{},
				})
			}
			datasourceInModuleWarnings[0].Datasource = append(datasourceInModuleWarnings[0].Datasource, *r)
		}
	}

	for _, c := range module.Children {
		datasourceInModuleWarnings = append(datasourceInModuleWarnings, computeDatasourceInModuleWarning(c)...)
	}

	return datasourceInModuleWarnings
}
