/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"guacamole/checks"
	"guacamole/helpers"

	"github.com/spf13/cobra"
)

// plan represents the run command
var static = &cobra.Command{
	Use:   "static",
	Short: "Run static code checks",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running static checks...")
		checkResults := checks.StaticChecks()
		helpers.RenderTable(checkResults)
	},
}

func init() {
	rootCmd.AddCommand(static)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}