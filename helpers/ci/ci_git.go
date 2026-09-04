package ci

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
)

func detectChangedDirs(projectDir, baseBranch, mrSHA string) ([]string, []string, error) {
	targetRef := "origin/" + baseBranch
	log.WithField("target_ref", targetRef).Debug("Validating target ref existence")

	if _, err := runGit(projectDir, "rev-parse", "--verify", targetRef); err != nil {
		log.WithField("base_branch", baseBranch).Debug("Target ref missing locally, fetching from origin")
		if _, fetchErr := runGit(projectDir, "fetch", "origin", baseBranch); fetchErr != nil {
			return nil, nil, logAndReturnErrorf("failed to fetch base branch %q: %w", baseBranch, fetchErr)
		}
	}

	changedOutput, err := runGit(projectDir, "diff", "--name-only", targetRef+"..."+mrSHA)
	if err != nil {
		return nil, nil, logAndReturnErrorf("failed to compute git diff for %s...%s: %w", targetRef, mrSHA, err)
	}
	log.WithField("changed_files_raw", changedOutput).Debug("Computed changed files")

	fmt.Println("Base branch :", baseBranch)
	fmt.Println("Target ref  :", targetRef)
	fmt.Println("MR SHA      :", mrSHA)
	fmt.Println("Changed files:")
	fmt.Println(changedOutput)

	layerSet := map[string]struct{}{}
	moduleSet := map[string]struct{}{}

	for _, file := range strings.Split(changedOutput, "\n") {
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
		if strings.HasPrefix(file, "base/") || strings.HasPrefix(file, "functional/") {
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
	cmd := exec.Command("git", args...)
	cmd.Dir = projectDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", logAndReturnErrorf("git %s failed: %w\n%s", strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return strings.TrimSpace(string(output)), nil
}
