package data

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform-exec/tfexec"
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
	log := log.New(os.Stderr, "", 0)
	dirPath := "/tmp/" + strings.ReplaceAll(l.Name, "/", "_")

	tf, err := tfexec.NewTerraform(l.FullPath, "terragrunt")
	if err != nil {
		log.Printf("failed to create Terraform instance: %s", err)
	}

	_, err = os.Stat(filepath.Join(l.FullPath, ".terragrunt-cache"))
	if os.IsNotExist(err) {
		err = tf.Init(context.Background())
		if err != nil {
			panic(err)
		}
	}

	// Don't lock the state file while running the plan
	_, err = tf.Plan(context.Background(), tfexec.Out(dirPath+"_plan.json"), tfexec.Lock(false))
	if err != nil {
		log.Printf("failed to create plan: %s", err)
	}

	// Create JSON plan
	jsonPlan, err := tf.ShowPlanFile(context.Background(), dirPath+"_plan.json")
	if err != nil {
		log.Printf("failed to create JSON plan: %s", err)
	}

	l.Plan = jsonPlan
}

func (l *Layer) ComputeState() {
	tf, err := tfexec.NewTerraform(l.FullPath, "terragrunt")
	if err != nil {
		log.Printf("failed to create Terraform instance: %s", err)
	}

	_, err = os.Stat(filepath.Join(l.FullPath, ".terragrunt-cache"))
	if os.IsNotExist(err) {
		err = tf.Init(context.Background())
		if err != nil {
			panic(err)
		}
	}

	// Create Terraform state file
	state, err := tf.Show(context.TODO())
	if err != nil {
		log.Printf("failed to create state: %s", err)
	}

	l.State = state
}

func (layer *Layer) BuildRootModule() {
	if layer.State == nil {
		layer.ComputeState()
	}

	if layer.State.Values != nil {
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
	if layer.RootModule.Name == "" {
		layer.BuildRootModule()
	}

	// Check for datasources in modules
	layer.Warnings.DatasourceInModuleWarning = computeDatasourceInModuleWarning(layer.RootModule)
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
