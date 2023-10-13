package helpers

import (
	"guacamole/data"
	"os"
	"strconv"

	"github.com/jedib0t/go-pretty/v6/table"
	"golang.org/x/term"
)

func RenderTable(checkResults []data.Check) {
	totalChecksOk, i, t := 0, 0, table.NewWriter()

	// Get the terminal width
	width, _, err := term.GetSize(0)
	if err != nil {
		panic(err)
	}
	// Determine column widths depending on the terminal size
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, WidthMax: 1 * width / 10},
		{Number: 2, WidthMax: 2 * width / 10},
		{Number: 3, WidthMax: 1 * width / 10},
		{Number: 4, WidthMax: 1 * width / 10},
		{Number: 5, WidthMax: 5 * width / 10},
	})
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
