package checks

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/padok-team/guacamole/data"
)

func CodeQualityTf(modules map[string]data.TerraformModule) (data.Check, error) {
	dataCheck := data.Check{
		ID:                "TF_QUA_001",
		Name:              "Terraform code should be properly formatted using `terraform fmt`",
		RelatedGuidelines: "https://padok-team.github.io/docs-terraform-guidelines/tooling/#code-quality",
		Status:            "✅",
	}

	if _, err := exec.LookPath("terraform"); err != nil {
		return data.Check{}, fmt.Errorf("terraform not found in PATH, cannot run TF_QUA_001: %w", err)
	}

	var errors []data.Error
	seen := make(map[string]bool)

	for _, module := range modules {
		cmd := exec.Command("terraform", "fmt", "--recursive", "--check")
		cmd.Dir = module.FullPath

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		runErr := cmd.Run()
		if runErr != nil {
			exitErr, ok := runErr.(*exec.ExitError)
			if !ok || exitErr.ExitCode() != 1 {
				return data.Check{}, fmt.Errorf("terraform fmt failed on %s: %w\n%s", module.FullPath, runErr, stderr.String())
			}
		}

		for _, file := range strings.Split(strings.TrimSpace(stdout.String()), "\n") {
			file = strings.TrimSpace(file)
			if file == "" {
				continue
			}
			fullPath := filepath.Join(module.FullPath, file)
			if seen[fullPath] {
				continue
			}
			seen[fullPath] = true
			errors = append(errors, data.Error{
				Path:        fullPath,
				Description: "File is not properly formatted. Run terraform fmt to fix it.",
			})
		}
	}

	dataCheck.Errors = errors
	if len(errors) > 0 {
		dataCheck.Status = "❌"
	}

	return dataCheck, nil
}
