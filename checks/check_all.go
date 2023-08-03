package checks

import (
	"fmt"
	"guacamole/data"
	"os"
	"strconv"

	"github.com/jedib0t/go-pretty/v6/table"
)

func CheckAll() {
	fmt.Println("Guacamole is cooking 🥑")
	totalChecksOk := 0
	// List of checks to perform
	listOfChecks := map[string](func() (data.Check, error)){
		"ProviderInModule":  ProviderInModule,
		"NoStuttering":      NoStuttering,
		"IterateNoUseCount": IterateNoUseCount,
	}

	totalChecks := len(listOfChecks)
	// Displaying the results
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Check", "Status", "Related guideline", "Errors"})
	i := 0
	for checkName, checkFunction := range listOfChecks {
		c, err := checkFunction()
		if err != nil {
			fmt.Printf("error while checking %s: %s", checkName, err)
		}
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
	score := strconv.Itoa(totalChecksOk*100/totalChecks) + "%"
	if score == "100%" {
		score = score + " 🎉"
	}
	t.AppendFooter(table.Row{"", "", "Score", score})
	t.Render()
	fmt.Println("Your guacamole is ready 🥑")
}
