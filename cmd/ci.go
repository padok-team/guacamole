package cmd

import (
	"os"

	"github.com/padok-team/guacamole/helpers/ci"
	"github.com/spf13/cobra"
)

var ciCmd = &cobra.Command{
	Use:          "ci",
	Short:        "Run CI-oriented static checks and post an optional GitLab MR comment",
	SilenceUsage: true,
	PreRun:       toggleDebug,
	RunE: func(cmd *cobra.Command, args []string) error {
		exitCode, err := ci.Run()
		if err != nil {
			return err
		}

		if exitCode != 0 {
			os.Exit(exitCode)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(ciCmd)
}
