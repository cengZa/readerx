package cmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/heybox/readerx/internal/app"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage ReaderX config",
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a config value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := openStore()
		if err != nil {
			return err
		}
		defer store.Close()

		value, err := app.NewConfigService(store).Get(args[0])
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n", value)
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a config value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := openStore()
		if err != nil {
			return err
		}
		defer store.Close()

		if err := app.NewConfigService(store).Set(args[0], args[1]); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s=%s\n", args[0], args[1])
		return nil
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List config values",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := openStore()
		if err != nil {
			return err
		}
		defer store.Close()

		settings, err := app.NewConfigService(store).List()
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "Key\tValue")
		for _, setting := range settings {
			fmt.Fprintf(w, "%s\t%s\n", setting.Key, setting.Value)
		}
		return w.Flush()
	},
}

func init() {
	configCmd.AddCommand(configGetCmd, configSetCmd, configListCmd)
	rootCmd.AddCommand(configCmd)
}
