/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"
	"os"

	"github.com/padok-team/guacamole/checks"
	"github.com/padok-team/guacamole/helpers"

	"github.com/spf13/cobra"
)

// allCmd represents the run command
var allCmd = &cobra.Command{
	Use:   "all",
	Short: "Run all checks",
	Run: func(cmd *cobra.Command, args []string) {
		l := log.New(os.Stderr, "", 0)
		l.Println("Running all checks...")
		layers, err := helpers.ComputeLayers(true)
		if err != nil {
			panic(err)
		}
		checkResults := checks.All(layers)
		helpers.RenderTable(checkResults)
	},
}

func init() {
	rootCmd.AddCommand(allCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
