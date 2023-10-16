/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"guacamole/checks"
	"guacamole/helpers"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// plan represents the run command
var static = &cobra.Command{
	Use:   "static",
	Short: "Run static code checks",
	Run: func(cmd *cobra.Command, args []string) {
		l := log.New(os.Stderr, "", 0)
		l.Println("Running static checks...")
		checkResults := checks.StaticChecks()
		// helpers.RenderTable(checkResults)
		verbose := viper.GetBool("verbose")
		helpers.RenderChecks(checkResults, verbose)
	},
}

func init() {
	rootCmd.AddCommand(static)

	static.PersistentFlags().BoolP("verbose", "v", false, "Display verbose output")

	viper.BindPFlag("verbose", static.PersistentFlags().Lookup("verbose"))
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
