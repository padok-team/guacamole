package ci

import (
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

func loadConfig() (config, error) {
	platform := detectPlatform()

	projectDir := firstNonEmptyEnv([]string{"GUACAMOLE_PROJECT_DIR", "CI_PROJECT_DIR", "GITHUB_WORKSPACE"}, ".")
	projectDirAbs, err := filepath.Abs(projectDir)
	if err != nil {
		return config{}, logAndReturnErrorf("failed to resolve project directory %q: %w", projectDir, err)
	}

	scanAll := envBool("GUACAMOLE_CI_SCAN_ALL", false)
	failOnError := envBool("GUACAMOLE_CI_FAIL_ON_ERROR", true)

	baseBranch := firstNonEmptyEnv([]string{"GUACAMOLE_DIFF_BASE_BRANCH", "CI_MERGE_REQUEST_TARGET_BRANCH_NAME", "GITHUB_BASE_REF"}, "")
	if baseBranch == "" && !scanAll {
		return config{}, logAndReturnErrorf("one of GUACAMOLE_DIFF_BASE_BRANCH, CI_MERGE_REQUEST_TARGET_BRANCH_NAME or GITHUB_BASE_REF must be set")
	}

	mrSHA := strings.TrimSpace(getOrDefaultEnv("GUACAMOLE_MR_SHA", "HEAD"))
	postComment := envBool("GUACAMOLE_CI_COMMENT", true)

	log.WithFields(log.Fields{
		"platform":      platform,
		"project_dir":   projectDirAbs,
		"base_branch":   baseBranch,
		"mr_sha":        mrSHA,
		"post_comment":  postComment,
		"scan_all":      scanAll,
		"fail_on_error": failOnError,
	}).Debug("Loaded CI configuration")

	return config{
		platform:      platform,
		projectDirAbs: projectDirAbs,
		baseBranch:    baseBranch,
		mrSHA:         mrSHA,
		postComment:   postComment,
		scanAll:       scanAll,
		failOnError:   failOnError,
	}, nil
}

// detectPlatform defaults to GitLab to keep the historical behaviour of the command.
func detectPlatform() platform {
	if strings.EqualFold(strings.TrimSpace(os.Getenv("GITHUB_ACTIONS")), "true") {
		return platformGithub
	}
	return platformGitlab
}

func firstNonEmptyEnv(keys []string, fallback string) string {
	for _, key := range keys {
		if v := strings.TrimSpace(os.Getenv(key)); v != "" {
			return v
		}
	}
	return fallback
}

func getOrDefaultEnv(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v != "" {
		return v
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if v == "" {
		return fallback
	}
	return v == "1" || v == "true" || v == "yes" || v == "y"
}
