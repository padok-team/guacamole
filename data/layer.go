package data

import (
	"bufio"
	"os"
	"os/exec"
)

type Layer struct {
	Name        string
	FullPath    string
	InitStatus  bool
	RefreshTime int
}

// Function to initialize a layer (terragrunt)
func (l *Layer) Init() error {
	if l.InitStatus {
		return nil
	}

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
