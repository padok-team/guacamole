package ci

import (
	"bytes"
	"fmt"
)

// commentMarker is a hidden HTML comment used to find the comment previously
// posted by Guacamole, so it can be updated instead of posting a new one.
const commentMarker = "<!-- guacamole-ci -->"

func buildCommentBody(results []result, overallScore, overallPass, overallTotal int) string {
	b := &bytes.Buffer{}
	b.WriteString(commentMarker + "\n")
	b.WriteString("### 🥑 Guacamole static checks\n")

	if len(results) == 0 {
		b.WriteString("\nNo modified layers/modules detected under `layers/`, `base/`, `functional/` or `modules/`.\n")
		return b.String()
	}

	scoreEmoji := "🚧"
	if overallScore == 100 {
		scoreEmoji = "🎉"
	}

	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Global score: %s %d%% (%d/%d)\n", scoreEmoji, overallScore, overallPass, overallTotal))
	b.WriteString("\n")
	b.WriteString("| Scope | Path | Score | Failed rules |\n")
	b.WriteString("|---|---|---|---|\n")
	for _, r := range results {
		b.WriteString(fmt.Sprintf("| %s | `%s` | %s | %s |\n", r.scope, r.path, r.score, r.failingText))
	}

	return b.String()
}
