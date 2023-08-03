/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"guacamole/checks"

	"github.com/spf13/cobra"
)

// noStutteringCmd represents the naming command
var noStutteringCmd = &cobra.Command{
	Use:   "no-stuttering",
	Short: "Check if you are reusing the name of the resource in the resource name",
	Run: func(cmd *cobra.Command, args []string) {
		checks.NoStuttering()
	},
}

func init() {
	rootCmd.AddCommand(noStutteringCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// namingCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// namingCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
