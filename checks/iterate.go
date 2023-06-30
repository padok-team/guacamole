package checks

import (
	"context"
	"fmt"
	"guacamole/data"
	"guacamole/helpers"
	"log"
	"regexp"
	"strconv"
	"strings"

	tfexec "github.com/hashicorp/terraform-exec/tfexec"
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
	// relatedGuidelines := "https://padok-team.github.io/docs-terraform-guidelines/terraform/iterate_on_your_resources.html#list-iteration-count"
	relatedGuidelines := "http://bitly.ws/K5WA"
	status := "✅"

	layers, err := helpers.GetLayers()
	if err != nil {
		log.Fatalf("Failed to get layers: %s", err)
	}

	// Create a channel to receive the results
	results := make(chan string)
	errs := make(chan error)

	for _, layer := range layers {
		go checkLayer(layer, results, errs)
	}

	// Wait for all goroutines to finish
	for i := 0; i < len(layers); i++ {
		select {
		case result := <-results:
			if result == "❌" {
				status = "❌"
			}
		case err := <-errs:
			log.Fatalf("Failed to check layer: %s", err)
		}
	}

	close(results)
	close(errs)

	return data.Check{
		Name:              name,
		RelatedGuidelines: relatedGuidelines,
		Status:            status,
	}
}

func checkLayer(layer data.Layer, results chan string, errs chan error) {
	status := "✅"

	dirPath := "/tmp/" + strings.ReplaceAll(layer.Name, "/", "_")

	tf, err := tfexec.NewTerraform(layer.FullPath, "terragrunt")
	if err != nil {
		errs <- fmt.Errorf("failed to create Terraform instance: %w", err)
	}

	err = tf.Init(context.TODO(), tfexec.Upgrade(true))
	if err != nil {
		errs <- fmt.Errorf("failed to initialize Terraform: %w", err)
	}

	// Create Terraform plan
	_, err = tf.Plan(context.Background(), tfexec.Out(dirPath+"_plan.json"))
	if err != nil {
		errs <- fmt.Errorf("failed to create plan: %w", err)
	}

	// Create JSON plan
	jsonPlan, err := tf.ShowPlanFile(context.Background(), dirPath+"_plan.json")
	if err != nil {
		errs <- fmt.Errorf("failed to create JSON plan: %w", err)
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
				// fmt.Printf("WARNING: Resource %s has count more than 1\n", rc.Address)
			}
		}
	}
	results <- status
}
