package checks

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/viper"
)

type Layer struct {
	Name        string
	FullPath    string
	InitStatus  bool
	RefreshTime int
}

func RunProfileCheck() {
	layers, _ := GetLayers()
	for _, layer := range layers {
		fmt.Println("Initializing layer", layer.FullPath)
		err := layer.Init()
		if err != nil {
			panic(err)
		}
		err = layer.GetRefreshTime()
		if err != nil {
			panic(err)
		}
	}
}

func GetLayers() ([]Layer, error) {
	root := viper.GetString("codebase-path") // Root directory to start browsing from
	layers := []Layer{}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if the current path is a file and its name matches "terragrunt.hcl"
		if !info.IsDir() && info.Name() == "terragrunt.hcl" {
			layers = append(layers, Layer{Name: path[len(root) : len(path)-len(info.Name())-1], FullPath: path[:len(path)-len(info.Name())-1]})
		}

		return nil
	})

	if err != nil {
		fmt.Println("Error:", err)
	}
	return layers, nil
}

// Function to initialize a layer (terragrunt)
func (l *Layer) Init() error {
	// Create the terragrunt command
	terragruntCmd := exec.Command("terragrunt", "init")

	// Set the command's working directory to the Terragrunt configuration directory
	terragruntCmd.Dir = l.FullPath

	// Redirect the command's output to the standard output
	terragruntCmd.Stdout = os.Stdout
	terragruntCmd.Stderr = os.Stderr

	// Run the terragrunt command
	err := terragruntCmd.Run()
	if err != nil {
		return err
	}

	l.InitStatus = true

	return nil
}

// Function to generate a layer plan using terragrunt
func (l *Layer) GetRefreshTime() error {
	// Create the terragrunt command
	terragruntCmd := exec.Command("terragrunt", "state", "list")

	// Set the command's working directory to the Terragrunt configuration directory
	terragruntCmd.Dir = l.FullPath

	// Create a pipe to capture the stdout
	stdoutPipe, err := terragruntCmd.StdoutPipe()
	if err != nil {
		return err
	}

	// Start the terragrunt command
	err = terragruntCmd.Start()
	if err != nil {
		return err
	}

	// Create a scanner to read from the stdout pipe
	scanner := bufio.NewScanner(stdoutPipe)
	lineCount := 0

	// Read each line from the stdout and count the number of lines
	for scanner.Scan() {
		lineCount++
	}

	// Wait for the command to finish
	err = terragruntCmd.Wait()
	if err != nil {
		return err
	}

	// Set the refresh time to the number of lines
	l.RefreshTime = lineCount

	return nil
}
