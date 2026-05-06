package cmd

import (
	"fmt"
	"os"

	"github.com/heybox/readerx/internal/app"
	"github.com/heybox/readerx/internal/tui"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var readChapter int
var readPlain bool

var readCmd = &cobra.Command{
	Use:   "read <book-id>",
	Short: "Read a chapter",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		bookID, err := parseIntArg(args[0], "book-id")
		if err != nil {
			return err
		}
		store, err := openStore()
		if err != nil {
			return err
		}
		defer store.Close()

		view, err := app.NewReadService(store).ReadChapter(bookID, readChapter)
		if err != nil {
			return err
		}
		if !readPlain && term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd())) {
			return tui.RunReader(store, view, view.Progress.LineOffset, readTUIOptions(store))
		}
		fmt.Fprintf(cmd.OutOrStdout(), "《%s》 %s\n\n%s\n", view.Book.Title, view.Chapter.Title, view.Chapter.Content)
		return nil
	},
}

func init() {
	readCmd.Flags().IntVar(&readChapter, "chapter", 0, "chapter number")
	readCmd.Flags().BoolVar(&readPlain, "plain", false, "print chapter without entering TUI")
	rootCmd.AddCommand(readCmd)
}
