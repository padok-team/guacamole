/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"guacamole/checks"

	"github.com/spf13/cobra"
)

// iterateNoUseCountCmd represents the iterate command
var iterateNoUseCountCmd = &cobra.Command{
	Use:   "iterate-no-use-count",
	Short: "Check if you are using count to create multiple resources",
	Run: func(cmd *cobra.Command, args []string) {
		checks.IterateNoUseCount()
	},
}

func init() {
	rootCmd.AddCommand(iterateNoUseCountCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// iterateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// iterateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
