package ci

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

const githubCommentsPerPage = 100

type githubComment struct {
	ID   int64  `json:"id"`
	Body string `json:"body"`
}

// postGithubComment creates the Guacamole comment on the pull request, or
// updates it in place if a previous run already posted one.
func postGithubComment(body string) error {
	prNumber, err := githubPullRequestNumber()
	if err != nil {
		return err
	}
	if prNumber == 0 {
		fmt.Println("No pull request detected, skipping GitHub comment")
		return nil
	}

	token := strings.TrimSpace(getOrDefaultEnv("GUACAMOLE_GITHUB_TOKEN", os.Getenv("GITHUB_TOKEN")))
	if token == "" {
		return logAndReturnErrorf("GUACAMOLE_GITHUB_TOKEN or GITHUB_TOKEN must be set to post PR comment")
	}

	repository := strings.TrimSpace(os.Getenv("GITHUB_REPOSITORY"))
	if repository == "" {
		return logAndReturnErrorf("GITHUB_REPOSITORY must be set to post PR comment")
	}

	client := githubClient{
		apiURL:     strings.TrimRight(getOrDefaultEnv("GITHUB_API_URL", "https://api.github.com"), "/"),
		repository: repository,
		token:      token,
	}

	existing, err := client.findComment(prNumber)
	if err != nil {
		return err
	}

	if existing != 0 {
		if err := client.do(http.MethodPatch, fmt.Sprintf("/issues/comments/%d", existing), map[string]string{"body": body}, nil); err != nil {
			return logAndReturnErrorf("failed to update GitHub PR comment: %w", err)
		}
		fmt.Println("Updated Guacamole comment on pull request", prNumber)
		return nil
	}

	if err := client.do(http.MethodPost, fmt.Sprintf("/issues/%d/comments", prNumber), map[string]string{"body": body}, nil); err != nil {
		return logAndReturnErrorf("failed to post GitHub PR comment: %w", err)
	}
	fmt.Println("Posted Guacamole comment to pull request", prNumber)
	return nil
}

// githubPullRequestNumber returns the PR number from GUACAMOLE_GITHUB_PR_NUMBER
// or from the event payload, and 0 when the workflow is not running on a PR.
func githubPullRequestNumber() (int, error) {
	if v := strings.TrimSpace(os.Getenv("GUACAMOLE_GITHUB_PR_NUMBER")); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return 0, logAndReturnErrorf("invalid GUACAMOLE_GITHUB_PR_NUMBER %q: %w", v, err)
		}
		return n, nil
	}

	eventPath := strings.TrimSpace(os.Getenv("GITHUB_EVENT_PATH"))
	if eventPath == "" {
		return 0, nil
	}

	content, err := os.ReadFile(eventPath)
	if err != nil {
		return 0, logAndReturnErrorf("failed to read GitHub event payload %q: %w", eventPath, err)
	}

	var event struct {
		PullRequest struct {
			Number int `json:"number"`
		} `json:"pull_request"`
	}
	if err := json.Unmarshal(content, &event); err != nil {
		return 0, logAndReturnErrorf("failed to parse GitHub event payload %q: %w", eventPath, err)
	}

	return event.PullRequest.Number, nil
}

type githubClient struct {
	apiURL     string
	repository string
	token      string
}

func (c githubClient) findComment(prNumber int) (int64, error) {
	for page := 1; ; page++ {
		var comments []githubComment
		path := fmt.Sprintf("/issues/%d/comments?per_page=%d&page=%d", prNumber, githubCommentsPerPage, page)
		if err := c.do(http.MethodGet, path, nil, &comments); err != nil {
			return 0, logAndReturnErrorf("failed to list GitHub PR comments: %w", err)
		}

		for _, comment := range comments {
			if strings.Contains(comment.Body, commentMarker) {
				log.WithField("comment_id", comment.ID).Debug("Found previous Guacamole comment")
				return comment.ID, nil
			}
		}

		if len(comments) < githubCommentsPerPage {
			return 0, nil
		}
	}
}

func (c githubClient) do(method, path string, payload any, out any) error {
	var reqBody io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		reqBody = bytes.NewReader(encoded)
	}

	req, err := http.NewRequest(method, fmt.Sprintf("%s/repos/%s%s", c.apiURL, c.repository, path), reqBody)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("github API returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

// writeGithubStepOutputs publishes the report to the job summary and exposes
// the score as step outputs, when running inside GitHub Actions. Failures are
// only logged: in a Docker action, guacamole runs as a non-root user that
// cannot write the runner files.
func writeGithubStepOutputs(body string, overallScore, overallPass, overallTotal int) {
	if summaryPath := strings.TrimSpace(os.Getenv("GITHUB_STEP_SUMMARY")); summaryPath != "" {
		if err := appendToFile(summaryPath, body+"\n"); err != nil {
			log.Warnf("Could not write GitHub step summary: %s", err)
		}
	}

	if outputPath := strings.TrimSpace(os.Getenv("GITHUB_OUTPUT")); outputPath != "" {
		outputs := fmt.Sprintf("score=%d\npassed=%d\ntotal=%d\n", overallScore, overallPass, overallTotal)
		if err := appendToFile(outputPath, outputs); err != nil {
			log.Warnf("Could not write GitHub step outputs: %s", err)
		}
	}
}

func appendToFile(path, content string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(content)
	return err
}
