/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"
	"os"

	"github.com/padok-team/guacamole/checks"
	"github.com/padok-team/guacamole/helpers"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// layerCmd represents the layer command
var layerCmd = &cobra.Command{
	Use:   "layer",
	Short: "Run static code checks on layers",
	Run: func(cmd *cobra.Command, args []string) {
		l := log.New(os.Stderr, "", 0)
		l.Println("Running static checks on layers...")
		checkResults := checks.LayerStaticChecks()
		// helpers.RenderTable(checkResults)
		verbose := viper.GetBool("verbose")
		helpers.RenderChecks(checkResults, verbose)
		// If there is at least one error, exit with code 1
		if helpers.HasError(checkResults) {
			os.Exit(1)
		}
	},
}

func init() {
	static.AddCommand(layerCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// layerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// layerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
