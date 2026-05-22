package checks

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/padok-team/guacamole/data"
	"github.com/spf13/viper"
)

func CodeQualityTg() (data.Check, error) {
	dataCheck := data.Check{
		ID:                "TG_QUA_001",
		Name:              "Terragrunt code should be properly formatted using `terragrunt hcl format`",
		RelatedGuidelines: "https://padok-team.github.io/docs-terraform-guidelines/tooling/#code-quality",
		Status:            "✅",
	}

	if _, err := exec.LookPath("terragrunt"); err != nil {
		return data.Check{}, fmt.Errorf("terragrunt not found in PATH, cannot run TG_QUA_001: %w", err)
	}

	codebasePath := viper.GetString("codebase-path")

	cmd := exec.Command("terragrunt", "hcl", "format", "--check")
	cmd.Dir = codebasePath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()
	if runErr != nil {
		exitErr, ok := runErr.(*exec.ExitError)
		if !ok || exitErr.ExitCode() != 1 {
			return data.Check{}, fmt.Errorf("terragrunt hcl format failed: %w\n%s", runErr, stderr.String())
		}
	}

	// terragrunt hcl format --check writes to stderr with ANSI color codes
	// Lines with unformatted files look like: "... ERROR  File './path/to/file.hcl' needs formatting"
	filePattern := regexp.MustCompile(`File '([^']+)' needs formatting`)
	ansiPattern := regexp.MustCompile(`\x1b\[[0-9;]*m`)

	var errors []data.Error
	for _, line := range strings.Split(stderr.String(), "\n") {
		clean := ansiPattern.ReplaceAllString(line, "")
		match := filePattern.FindStringSubmatch(clean)
		if match == nil {
			continue
		}
		errors = append(errors, data.Error{
			Path:        match[1],
			Description: "File is not properly formatted. Run terragrunt hcl format to fix it.",
		})
	}

	dataCheck.Errors = errors
	if len(errors) > 0 {
		dataCheck.Status = "❌"
	}

	return dataCheck, nil
}
