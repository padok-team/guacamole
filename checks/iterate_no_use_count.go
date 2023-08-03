package checks

import (
	"context"
	"fmt"
	"guacamole/data"
	"guacamole/helpers"
	"regexp"
	"strconv"
	"strings"
	"sync"

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

func IterateNoUseCount() (data.Check, error) {
	name := "Don't use count to create multiple resources"
	// relatedGuidelines := "https://padok-team.github.io/docs-terraform-guidelines/terraform/iterate_on_your_resources.html#list-iteration-count"
	relatedGuidelines := "http://bitly.ws/K5WA"
	status := "✅"
	errors := []string{}

	layers, err := helpers.GetLayers()
	if err != nil {
		return data.Check{}, err
	}

	// Create a channel to receive the results
	results, checkErrors, errs := make(chan string), make(chan string), make(chan error)
	defer close(results)
	defer close(checkErrors)
	defer close(errs)

	wg := new(sync.WaitGroup)
	wg.Add(len(layers))

	for _, layer := range layers {
		go checkLayer(layer, results, checkErrors, errs, wg)
	}

	wg.Wait() // Here we wait for all the goroutines to finish

	// Wait for all goroutines to finish
	for i := 0; i < len(layers); i++ {
		select {
		case result := <-results:
			if result == "❌" {
				status = "❌"
			}
		case checkError := <-checkErrors:
			errors = append(errors, checkError)

		case err := <-errs:
			return data.Check{}, fmt.Errorf("error while checking layer: %w", err)
		}
	}

	data := data.Check{
		Name:              name,
		RelatedGuidelines: relatedGuidelines,
		Status:            status,
		Errors:            errors,
	}

	return data, nil
}

func checkLayer(layer data.Layer, result, checkError chan string, errs chan error, wg *sync.WaitGroup) {
	status := "✅"

	dirPath := "/tmp/" + strings.ReplaceAll(layer.Name, "/", "_")

	fmt.Println("Checking layer", layer.Name)
	tf, err := tfexec.NewTerraform(layer.FullPath, "terragrunt")
	if err != nil {
		errs <- fmt.Errorf("failed to create Terraform instance: %w", err)
	}

	fmt.Println("Initializing Terraform for layer", layer.Name)
	err = tf.Init(context.TODO(), tfexec.Upgrade(true))
	if err != nil {
		errs <- fmt.Errorf("failed to initialize Terraform: %w", err)
	}

	// Create Terraform plan
	fmt.Println("Creating plan for layer", layer.Name)
	_, err = tf.Plan(context.Background(), tfexec.Out(dirPath+"_plan.json"))
	if err != nil {
		errs <- fmt.Errorf("failed to create plan: %w", err)
	}

	// Create JSON plan
	fmt.Println("Showing JSON plan for layer", layer.Name)
	jsonPlan, err := tf.ShowPlanFile(context.Background(), dirPath+"_plan.json")
	if err != nil {
		errs <- fmt.Errorf("failed to create JSON plan: %w", err)
	}

	var index int
	var indexString string
	var checkedResources []string

	// Analyze plan for resources with count > 1
	fmt.Println("Analyzing plan for layer", layer.Name)
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
				fmt.Println("Resource", rc.Address, "in layer", layer.Name, "has count more than 1")
				checkError <- fmt.Sprintf("Resource %s in layer %s has count more than 1\n", rc.Address, layer.Name)
			}
		}
	}
	fmt.Println("Done checking layer", layer.Name)
	result <- status
	wg.Done()
}
