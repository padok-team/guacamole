package checks

import (
	"fmt"
	"os"
	"strconv"

	"github.com/jedib0t/go-pretty/v6/table"
)

func Checks() {
	fmt.Println("Guacamole is cooking 🥑")
	totalChecks := 2
	totalChecksOk := 0
	// Checking the module for provider
	check1 := ProviderInModule()
	// Checkoing the naming of the resources
	check2 := Naming()
	//Calculating the score in percentage
	if check1.Status == "✅" {
		totalChecksOk++
	}
	if check2.Status == "✅" {
		totalChecksOk++
	}
	score := strconv.Itoa(totalChecksOk*100/totalChecks) + "%"
	if score == "100%" {
		score = score + " 🎉"
	}
	// Displaying the results
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Check", "Status", "Related guidelines"})
	t.AppendRows([]table.Row{
		{1, check1.Name, check1.Status, check1.RelatedGuidelines},
		{2, check2.Name, check2.Status, check2.RelatedGuidelines},
	})
	t.AppendFooter(table.Row{"", "", "Score", score})
	t.Render()
	fmt.Println("Your guacamole is ready 🥑")
}
