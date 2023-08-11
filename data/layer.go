package data

import (
	"context"
	"fmt"
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
	RootModule Module
}

func (l *Layer) ComputePlan() {
	dirPath := "/tmp/" + strings.ReplaceAll(l.Name, "/", "_")

	tf, err := tfexec.NewTerraform(l.FullPath, "terragrunt")
	if err != nil {
		fmt.Printf("failed to create Terraform instance: %s", err)
	}

	_, err = os.Stat(filepath.Join(l.FullPath, ".terragrunt-cache"))
	if os.IsNotExist(err) {
		err = tf.Init(context.Background())
		if err != nil {
			panic(err)
		}
	}

	// Create Terraform plan
	_, err = tf.Plan(context.Background(), tfexec.Out(dirPath+"_plan.json"))
	if err != nil {
		fmt.Printf("failed to create plan: %s", err)
	}

	// Create JSON plan
	jsonPlan, err := tf.ShowPlanFile(context.Background(), dirPath+"_plan.json")
	if err != nil {
		fmt.Printf("failed to create JSON plan: %s", err)
	}

	l.Plan = jsonPlan
}

func (l *Layer) ComputeState() {
	tf, err := tfexec.NewTerraform(l.FullPath, "terragrunt")
	if err != nil {
		fmt.Printf("failed to create Terraform instance: %s", err)
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
		fmt.Printf("failed to create state: %s", err)
	}

	l.State = state
}

func (layer *Layer) BuildRootModule() {
	if layer.State == nil {
		layer.ComputeState()
	}

	layer.RootModule = Module{
		Address:     "root",
		ObjectTypes: []ObjectType{},
		Children:    []Module{},
	}

	layer.RootModule.buildModule(layer.State.Values.RootModule)
}
