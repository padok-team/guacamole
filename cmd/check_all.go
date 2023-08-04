/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"guacamole/checks"

	"github.com/spf13/cobra"
)

// checkAllCmd represents the run command
var checkAllCmd = &cobra.Command{
	Use:   "check-all",
	Short: "Run all checks",
	Run: func(cmd *cobra.Command, args []string) {
		checks.CheckAll()
	},
}

func init() {
	rootCmd.AddCommand(checkAllCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
