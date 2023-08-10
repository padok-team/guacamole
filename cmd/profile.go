/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"guacamole/checks"
	"guacamole/helpers"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// plan represents the run command
var profile = &cobra.Command{
	Use:   "profile",
	Short: "Display informations about resources and datasources contained in the codebase",
	Run: func(cmd *cobra.Command, args []string) {
		layers, err := helpers.ComputeLayers(false)
		if err != nil {
			panic(err)
		}
		verbose := viper.GetBool("verbose")
		checks.Profile(layers, verbose)
		// helpers.RenderTable(checkResults)
	},
}

func init() {
	rootCmd.AddCommand(profile)

	// Add verbose flag
	profile.PersistentFlags().BoolP("verbose", "v", false, "Display verbose output")

	viper.BindPFlag("verbose", profile.PersistentFlags().Lookup("verbose"))

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
