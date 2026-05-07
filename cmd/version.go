package cmd

import (
	"fmt"

	"github.com/heybox/readerx/internal/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show ReaderX version information",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(cmd.OutOrStdout(), version.Current().String())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
