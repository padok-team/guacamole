/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"guacamole/checks"

	"github.com/spf13/cobra"
)

// iterateCmd represents the iterate command
var iterateCmd = &cobra.Command{
	Use:   "iterate",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		checks.Iterate()
		fmt.Println("iterate called")
	},
}

func init() {
	rootCmd.AddCommand(iterateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// iterateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// iterateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
