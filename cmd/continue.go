package cmd

import (
	"fmt"
	"os"

	"github.com/heybox/readerx/internal/app"
	"github.com/heybox/readerx/internal/tui"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var continuePlain bool

var continueCmd = &cobra.Command{
	Use:   "continue",
	Short: "Continue the most recently read book",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := openStore()
		if err != nil {
			return err
		}
		defer store.Close()

		view, err := app.NewReadService(store).Continue()
		if err != nil {
			return err
		}
		if !continuePlain && term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd())) {
			return tui.RunReader(store, view, view.Progress.LineOffset, readTUIOptions(store))
		}
		fmt.Fprintf(cmd.OutOrStdout(), "《%s》 %s\n\n%s\n", view.Book.Title, view.Chapter.Title, view.Chapter.Content)
		return nil
	},
}

func init() {
	continueCmd.Flags().BoolVar(&continuePlain, "plain", false, "print chapter without entering TUI")
	rootCmd.AddCommand(continueCmd)
}
