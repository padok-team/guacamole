/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"guacamole/checks"
	"guacamole/helpers"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// state represents the run command
var state = &cobra.Command{
	Use:   "state",
	Short: "[EXPERIMENTAL] Run state-related checks",
	Long: `[EXPERIMENTAL] Run state-related checks
⚠️ WARNING: This command may fail in unexpected way if all the layers you want to check are not initialized properly.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		verbose := viper.GetBool("verbose")
		l := log.New(os.Stderr, "", 0)
		var prompt string
		fmt.Println("⚠️ WARNING: This command may fail in unexpected way if all the layers you want to check are not initialized properly.")
		for prompt != "y" && prompt != "n" {
			fmt.Print("Please confirm that you want to run this command (y/n) : ")
			fmt.Scanln(&prompt)
		}
		if prompt == "y" {
			layers, err := helpers.ComputeLayers(false)
			l.Println("Running state checks...")
			if err != nil {
				panic(err)
			}
			checkResults := checks.StateChecks(layers)
			helpers.RenderChecks(checkResults, verbose)
		} else {
			l.Println("Aborting...")
		}
	},
}

func init() {
	rootCmd.AddCommand(state)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
