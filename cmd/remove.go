package cmd

import (
	"fmt"

	"github.com/heybox/readerx/internal/app"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <book-id>",
	Short: "Remove a book from the library",
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

		book, err := app.NewLibraryService(store).RemoveBook(bookID)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "已删除书籍：%s（Book ID：%d）\n", book.Title, book.ID)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
