package checks

import (
	"context"
	"fmt"
	"guacamole/data"
	"log"
	"regexp"
	"strconv"

	tfexec "github.com/hashicorp/terraform-exec/tfexec"
	"github.com/spf13/viper"
)

type Change struct {
	Actions []string `json:"actions"`
}

type ResourceChange struct {
	Address      string `json:"address"`
	Mode         string `json:"mode"`
	Type         string `json:"type"`
	Name         string `json:"name"`
	Index        string `json:"index"`
	ProviderName string `json:"provider_name"`
	Change       Change `json:"change"`
}

type ResourceChanges struct {
	ResourceChanges []ResourceChange `json:"resource_changes"`
}

func Iterate() data.Check {
	name := "Don't use count to create multiple resources"
	relatedGuidelines := "https://padok-team.github.io/docs-terraform-guidelines/terraform/iterate_on_your_resources.html#list-iteration-count"
	status := "✅"

	codebasePath := viper.GetString("codebase-path")

	tf, err := tfexec.NewTerraform(codebasePath, "terragrunt")
	if err != nil {
		log.Fatalf("Failed to create Terraform instance: %s", err)
	}

	err = tf.Init(context.TODO(), tfexec.Upgrade(true))
	if err != nil {
		log.Fatalf("Failed to initialize Terraform: %s", err)
	}

	// Create Terraform plan
	_, err = tf.Plan(context.Background(), tfexec.Out("/tmp/plan.json"))
	if err != nil {
		log.Fatalf("Failed to create plan: %s", err)
	}

	// Create JSON plan
	jsonPlan, err := tf.ShowPlanFile(context.Background(), "/tmp/plan.json")
	if err != nil {
		log.Fatalf("Failed to create JSON plan: %s", err)
	}

	var index int
	var indexString string
	var checkedResources []string

	// Analyze plan for resources with count > 1
	for _, rc := range jsonPlan.ResourceChanges {
		// Parse the module address to find numbers inside of []
		regexpIndexMatch := regexp.MustCompile(`\[(.*?)\]`).FindStringSubmatch(rc.Address)
		if len(regexpIndexMatch) > 0 {
			indexString = regexpIndexMatch[1]
		}

		// Ignore error in case of a string into the brackets, meaning that the resource was not created with count
		index, _ = strconv.Atoi(indexString)

		if index > 0 {
			// Check if the resource was already checked
			alreadyChecked := false
			for _, checkedResource := range checkedResources {
				if checkedResource == rc.ModuleAddress {
					alreadyChecked = true
					break
				}
			}
			if !alreadyChecked {
				status = "❌"
				checkedResources = append(checkedResources, rc.ModuleAddress)
				fmt.Printf("WARNING: Resource %s has count more than 1\n", rc.Address)
			}
		}
	}

	return data.Check{
		Name:              name,
		RelatedGuidelines: relatedGuidelines,
		Status:            status,
	}
}
