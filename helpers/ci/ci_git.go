package ci

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
)

func detectDirs(cfg config) ([]string, []string, error) {
	var files string
	var err error
	if cfg.scanAll {
		files, err = listTrackedFiles(cfg.projectDirAbs)
	} else {
		files, err = listChangedFiles(cfg.projectDirAbs, cfg.baseBranch, cfg.mrSHA)
	}
	if err != nil {
		return nil, nil, err
	}

	return splitScopeDirs(files)
}

// listTrackedFiles returns the Terraform files and Terragrunt layer files
// tracked under the project directory, relative to it, so that only actual
// modules and layers are scanned.
func listTrackedFiles(projectDir string) (string, error) {
	files, err := runGit(projectDir, "ls-files", "--", "*.tf", "*terragrunt.hcl")
	if err != nil {
		return "", logAndReturnErrorf("failed to list tracked files: %w", err)
	}

	fmt.Println("Scanning all tracked files")
	return files, nil
}

// listChangedFiles returns the files changed under the project directory
// between the base branch and the MR SHA, relative to the project directory.
func listChangedFiles(projectDir, baseBranch, mrSHA string) (string, error) {
	targetRef := "origin/" + baseBranch
	log.WithField("target_ref", targetRef).Debug("Validating target ref existence")

	if _, err := runGit(projectDir, "rev-parse", "--verify", targetRef); err != nil {
		log.WithField("base_branch", baseBranch).Debug("Target ref missing locally, fetching from origin")
		if _, fetchErr := runGit(projectDir, "fetch", "origin", baseBranch); fetchErr != nil {
			return "", logAndReturnErrorf("failed to fetch base branch %q: %w", baseBranch, fetchErr)
		}
	}

	changedOutput, err := runGit(projectDir, "diff", "--name-only", "--relative", targetRef+"..."+mrSHA)
	if err != nil {
		if shallow, _ := runGit(projectDir, "rev-parse", "--is-shallow-repository"); shallow == "true" {
			return "", logAndReturnErrorf("failed to compute git diff for %s...%s on a shallow clone, fetch the full history (e.g. fetch-depth: 0 with actions/checkout): %w", targetRef, mrSHA, err)
		}
		return "", logAndReturnErrorf("failed to compute git diff for %s...%s: %w", targetRef, mrSHA, err)
	}
	log.WithField("changed_files_raw", changedOutput).Debug("Computed changed files")

	fmt.Println("Base branch :", baseBranch)
	fmt.Println("Target ref  :", targetRef)
	fmt.Println("MR SHA      :", mrSHA)
	fmt.Println("Changed files:")
	fmt.Println(changedOutput)

	return changedOutput, nil
}

// splitScopeDirs keeps the directories of the files located under layers/
// (Terragrunt layers) or base/, functional/ and modules/ (Terraform modules).
func splitScopeDirs(files string) ([]string, []string, error) {
	layerSet := map[string]struct{}{}
	moduleSet := map[string]struct{}{}

	for _, file := range strings.Split(files, "\n") {
		file = strings.TrimSpace(file)
		if file == "" {
			continue
		}

		dir := filepath.ToSlash(filepath.Dir(file))
		if dir == "." {
			continue
		}

		if strings.HasPrefix(file, "layers/") {
			layerSet[dir] = struct{}{}
		}
		if strings.HasPrefix(file, "base/") || strings.HasPrefix(file, "functional/") || strings.HasPrefix(file, "modules/") {
			moduleSet[dir] = struct{}{}
		}
	}

	layerDirs := mapKeysSorted(layerSet)
	moduleDirs := mapKeysSorted(moduleSet)

	fmt.Println("Changed layers:")
	for _, d := range layerDirs {
		fmt.Println(d)
	}
	fmt.Println("Changed modules:")
	for _, d := range moduleDirs {
		fmt.Println(d)
	}

	return layerDirs, moduleDirs, nil
}

func mapKeysSorted(set map[string]struct{}) []string {
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func runGit(projectDir string, args ...string) (string, error) {
	log.WithFields(log.Fields{"dir": projectDir, "args": strings.Join(args, " ")}).Debug("Running git command")
	// CI runners often check out repos as a different user, which trips git's
	// "dubious ownership" safety check; scope the exception to this invocation only.
	// The check applies to the repository root, which may be a parent of projectDir.
	safeArgs := append([]string{"-c", "safe.directory=*"}, args...)
	cmd := exec.Command("git", safeArgs...)
	cmd.Dir = projectDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", logAndReturnErrorf("git %s failed: %w\n%s", strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return strings.TrimSpace(string(output)), nil
}
