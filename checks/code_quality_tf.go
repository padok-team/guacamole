package checks

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/padok-team/guacamole/data"
	log "github.com/sirupsen/logrus"
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

	if _, err := exec.LookPath("terraform"); err != nil {
		return data.Check{}, fmt.Errorf("terraform not found in PATH, cannot run TF_QUA_001: %w", err)
	}

	log.Debugf("[TF_QUA_001] Starting check on %d module(s)", len(modules))

	for _, module := range modules {
		log.Debugf("[TF_QUA_001] Checking module: %s", module.FullPath)

		cmd := exec.Command("terraform", "fmt", "--recursive", "--check")
		cmd.Dir = module.FullPath
		log.Debugf("[TF_QUA_001] Running command: %v in dir: %s", cmd.Args, cmd.Dir)

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		runErr := cmd.Run()
		log.Debugf("[TF_QUA_001] Command error: %v", runErr)
		log.Debugf("[TF_QUA_001] stdout: %q", stdout.String())
		log.Debugf("[TF_QUA_001] stderr: %q", stderr.String())

		output := strings.TrimSpace(stdout.String())
		if output == "" {
			log.Debugf("[TF_QUA_001] No formatting issues found in module: %s", module.FullPath)
			continue
		}

		files := strings.Split(output, "\n")
		log.Debugf("[TF_QUA_001] Files needing formatting in %s: %v", module.FullPath, files)

		for _, file := range files {
			file = strings.TrimSpace(file)
			if file == "" {
				continue
			}
			fullPath := filepath.Join(module.FullPath, file)
			log.Debugf("[TF_QUA_001] Unformatted file: %s (already seen: %v)", fullPath, seen[fullPath])
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

	log.Debugf("[TF_QUA_001] Total errors: %d", len(errors))
	dataCheck.Errors = errors

	if len(errors) > 0 {
		dataCheck.Status = "❌"
	}

	return dataCheck, nil
}
