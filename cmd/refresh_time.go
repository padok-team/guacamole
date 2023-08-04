/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"guacamole/checks"

	"github.com/spf13/cobra"
)

// refreshTimeCmd represents the profile command
var refreshTimeCmd = &cobra.Command{
	Use:   "refresh-time",
	Short: "Estimate the refresh time of the layers of your codebase",
	Run: func(cmd *cobra.Command, args []string) {
		checks.RefreshTime()
	},
}

func init() {
	rootCmd.AddCommand(refreshTimeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// profileCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// profileCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
