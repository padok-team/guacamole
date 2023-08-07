package data

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
)

type Layer struct {
	Name       string
	FullPath   string
	InitStatus bool
	Plan       *tfjson.Plan
	State      *tfjson.State
}

func (l *Layer) Init() {
	tf, err := tfexec.NewTerraform(l.FullPath, "terragrunt")
	if err != nil {
		fmt.Printf("failed to create Terraform instance: %s", err)
	}

	err = tf.Init(context.Background())
	if err != nil {
		fmt.Printf("failed to initialize Terraform: %s", err)
	}

	l.InitStatus = true
}

func (l *Layer) ComputePlan() {
	dirPath := "/tmp/" + strings.ReplaceAll(l.Name, "/", "_")

	tf, err := tfexec.NewTerraform(l.FullPath, "terragrunt")
	if err != nil {
		fmt.Printf("failed to create Terraform instance: %s", err)
	}

	if !l.InitStatus {
		l.Init()
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

	if !l.InitStatus {
		l.Init()
	}

	// Create Terraform state file
	state, err := tf.Show(context.TODO())
	if err != nil {
		fmt.Printf("failed to create state: %s", err)
	}

	l.State = state
}
