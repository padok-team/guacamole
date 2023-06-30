package checks

import (
	"fmt"
	"guacamole/data"
	"os"
	"strconv"

	"github.com/jedib0t/go-pretty/v6/table"
)

func Checks() {
	fmt.Println("Guacamole is cooking 🥑")
	totalChecksOk := 0
	// List of checks to perform
	listOfChecks := []func() data.Check{
		ProviderInModule,
		Naming,
		Iterate,
		Profile,
	}
	totalChecks := len(listOfChecks)
	// Displaying the results
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Check", "Status", "Related guidelines"})
	for i, check := range listOfChecks {
		c := check()
		if c.Status == "✅" {
			totalChecksOk++
		}
		t.AppendRows([]table.Row{
			{i + 1, c.Name, c.Status, c.RelatedGuidelines},
		})
	}
	score := strconv.Itoa(totalChecksOk*100/totalChecks) + "%"
	if score == "100%" {
		score = score + " 🎉"
	}
	t.AppendFooter(table.Row{"", "", "Score", score})
	t.Render()
	fmt.Println("Your guacamole is ready 🥑")
}
