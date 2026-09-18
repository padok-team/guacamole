package ci

import (
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

func loadConfig() (config, error) {
	projectDir := strings.TrimSpace(getOrDefaultEnv("CI_PROJECT_DIR", "."))
	projectDirAbs, err := filepath.Abs(projectDir)
	if err != nil {
		return config{}, logAndReturnErrorf("failed to resolve project directory %q: %w", projectDir, err)
	}

	baseBranch := strings.TrimSpace(getOrDefaultEnv("GUACAMOLE_DIFF_BASE_BRANCH", os.Getenv("CI_MERGE_REQUEST_TARGET_BRANCH_NAME")))
	if baseBranch == "" {
		return config{}, logAndReturnErrorf("either GUACAMOLE_DIFF_BASE_BRANCH or CI_MERGE_REQUEST_TARGET_BRANCH_NAME must be set")
	}

	mrSHA := strings.TrimSpace(getOrDefaultEnv("GUACAMOLE_MR_SHA", "HEAD"))
	postComment := shouldPostComment()

	log.WithFields(log.Fields{
		"project_dir":  projectDirAbs,
		"base_branch":  baseBranch,
		"mr_sha":       mrSHA,
		"post_comment": postComment,
	}).Debug("Loaded CI configuration")

	return config{
		projectDirAbs: projectDirAbs,
		baseBranch:    baseBranch,
		mrSHA:         mrSHA,
		postComment:   postComment,
	}, nil
}

func getOrDefaultEnv(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v != "" {
		return v
	}
	return fallback
}

func shouldPostComment() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("GUACAMOLE_CI_COMMENT")))
	if v == "" {
		return true
	}
	return v == "1" || v == "true" || v == "yes" || v == "y"
}
