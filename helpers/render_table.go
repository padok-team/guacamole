package helpers

import (
	"guacamole/data"
	"os"
	"strconv"

	"github.com/jedib0t/go-pretty/v6/table"
)

func RenderTable(checkResults []data.Check) {
	totalChecksOk, i, t := 0, 0, table.NewWriter()

	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Check", "Status", "Related guideline", "Errors"})

	for _, c := range checkResults {
		if c.Status == "✅" {
			totalChecksOk++
		}
		t.AppendRows([]table.Row{
			{i + 1, c.Name, c.Status, c.RelatedGuidelines},
		})
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				t.AppendRows([]table.Row{
					{"", "", "", "", err},
				})
			}
		}
		t.AppendSeparator()
		i++
	}

	score := strconv.Itoa(totalChecksOk*100/len(checkResults)) + "%"
	if score == "100%" {
		score = score + " 🎉"
	}
	t.AppendFooter(table.Row{"", "", "Score", score})
	t.Render()
}
