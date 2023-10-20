package helpers

import (
	"fmt"
	"os"
	"strconv"

	"github.com/padok-team/guacamole/data"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/savioxavier/termlink"
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

func RenderChecks(checkResults []data.Check, verbose bool) {
	totalChecksOk, i := 0, 0
	for _, c := range checkResults {
		if c.Status == "✅" {
			totalChecksOk++
		}
		i++
	}
	// Format the score
	score := strconv.Itoa(totalChecksOk*100/len(checkResults)) + "%"
	if score == "100%" {
		score = score + " 🎉"
	}
	// Print the checks
	for _, c := range checkResults {
		fmt.Printf("%s %s - %s\n", c.Status, c.ID, termlink.Link(c.Name, c.RelatedGuidelines))
		if len(c.Errors) > 0 && verbose {
			for _, err := range c.Errors {
				fmt.Println("  -", err)
			}
		}
	}
	// Print the score
	fmt.Printf("Score: %s (%d/%d)\n", score, totalChecksOk, i)
}
