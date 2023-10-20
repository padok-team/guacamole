package cmd

import (
	"fmt"
	"guacamole/internal/version"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Get the current guacamole version",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("%v \n", version.BuildVersion())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
