package checks

import (
	"bytes"
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

	var errors []data.Error
	seen := make(map[string]bool)

	for _, module := range modules {
		cmd := exec.Command("terraform", "fmt", "--recursive", "--check")
		cmd.Dir = module.FullPath

		var stdout bytes.Buffer
		cmd.Stdout = &stdout

		_ = cmd.Run()

		output := strings.TrimSpace(stdout.String())
		if output == "" {
			continue
		}

		for _, file := range strings.Split(output, "\n") {
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
