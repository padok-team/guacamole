/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/padok-team/guacamole/checks"
	"github.com/padok-team/guacamole/data"
	"github.com/padok-team/guacamole/helpers"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// plan represents the run command
var profile = &cobra.Command{
	Use:   "profile",
	Short: "[EXPERIMENTAL] Display informations about resources and datasources contained in the codebase",
	Long: `[EXPERIMENTAL] Display informations about resources and datasources contained in the codebase
⚠️ WARNING: This command may fail in unexpected way if all the layers you want to check are not initialized properly.
`,
	Run: func(cmd *cobra.Command, args []string) {
		l := log.New(os.Stderr, "", 0)
		var prompt string
		fmt.Println("⚠️ WARNING: This command may fail in unexpected way if all the layers you want to check are not initialized properly.")
		for prompt != "y" && prompt != "n" {
			fmt.Print("Please confirm that you want to run this command (y/n) : ")
			fmt.Scanln(&prompt)
		}
		if prompt == "y" {
			l.Println("Profiling layers...")
			layers, err := helpers.ComputeLayers(false)
			codebase := data.Codebase{
				Layers: layers,
			}
			if err != nil {
				panic(err)
			}
			verbose := viper.GetBool("verbose")
			checks.Profile(codebase, verbose)
			// helpers.RenderTable(checkResults)
		}
	},
}

func init() {
	rootCmd.AddCommand(profile)

	// Add verbose flag

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
