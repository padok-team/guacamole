package ci

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func postGitlabComment(results []result, overallScore, overallPass, overallTotal int) error {
	mrIID := strings.TrimSpace(os.Getenv("CI_MERGE_REQUEST_IID"))
	if mrIID == "" {
		return nil
	}

	token := strings.TrimSpace(os.Getenv("GUACAMOLE_GITLAB_TOKEN"))
	if token == "" {
		return logAndReturnErrorf("GUACAMOLE_GITLAB_TOKEN must be set to post MR comment")
	}

	apiV4 := strings.TrimSpace(os.Getenv("CI_API_V4_URL"))
	projectID := strings.TrimSpace(os.Getenv("CI_PROJECT_ID"))
	if apiV4 == "" || projectID == "" {
		return logAndReturnErrorf("CI_API_V4_URL and CI_PROJECT_ID must be set to post MR comment")
	}

	body := buildCommentBody(results, overallScore, overallPass, overallTotal)
	apiURL := fmt.Sprintf("%s/projects/%s/merge_requests/%s/notes", strings.TrimRight(apiV4, "/"), projectID, mrIID)

	values := url.Values{}
	values.Set("body", body)

	req, err := http.NewRequest(http.MethodPost, apiURL, strings.NewReader(values.Encode()))
	if err != nil {
		return logAndReturnErrorf("failed to create GitLab comment request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("PRIVATE-TOKEN", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return logAndReturnErrorf("failed to post GitLab MR comment: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		payload, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return logAndReturnErrorf("gitlab API returned %s while posting MR comment: %s", resp.Status, strings.TrimSpace(string(payload)))
	}

	fmt.Println("Posted Guacamole comment to merge request", mrIID)
	return nil
}

func buildCommentBody(results []result, overallScore, overallPass, overallTotal int) string {
	b := &bytes.Buffer{}
	b.WriteString("### 🥑 Guacamole static checks\n")

	if len(results) == 0 {
		b.WriteString("\nNo modified layers/modules detected under `layers/`, `base/` or `functional/`.\n")
		return b.String()
	}

	scoreEmoji := "🚧"
	if overallScore == 100 {
		scoreEmoji = "🎉"
	}

	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("- Global score: %s %d%% (%d/%d)\n", scoreEmoji, overallScore, overallPass, overallTotal))
	b.WriteString("\n")
	b.WriteString("| Scope | Path | Score | Failed rules |\n")
	b.WriteString("|---|---|---|---|\n")
	for _, r := range results {
		b.WriteString(fmt.Sprintf("| %s | `%s` | %s | %s |\n", r.scope, r.path, r.score, r.failingText))
	}

	return b.String()
}
