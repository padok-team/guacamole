/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/padok-team/guacamole/checks"
	"github.com/padok-team/guacamole/helpers"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// plan represents the run command
var static = &cobra.Command{
	Use:    "static",
	Short:  "Run static code checks",
	PreRun: toggleDebug,
	Run: func(cmd *cobra.Command, args []string) {
		log.Warn("You can specify what you want to check : layer or module")
		log.Info("Defaulting to checking both")
		log.Info("Running static checks on module...")
		checkResults := checks.ModuleStaticChecks()
		log.Info("Running static checks on layer...")
		checkResults = append(checkResults, checks.LayerStaticChecks()...)
		verbose := viper.GetBool("verbose")
		helpers.RenderChecks(checkResults, verbose)
		// If there is at least one error, exit with code 1
		if helpers.HasError(checkResults) {
			os.Exit(1)
		}
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
