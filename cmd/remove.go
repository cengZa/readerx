package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/heybox/readerx/internal/app"
	"github.com/heybox/readerx/internal/domain"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var removeYes bool

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

		service := app.NewLibraryService(store)
		info, err := service.BookInfo(bookID)
		if err != nil {
			return err
		}
		if err := confirmRemove(cmd, info.Book, removeYes); err != nil {
			return err
		}
		book, err := service.RemoveBook(bookID)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "已删除书籍：%s（Book ID：%d）\n", book.Title, book.ID)
		return nil
	},
}

func confirmRemove(cmd *cobra.Command, book domain.Book, yes bool) error {
	if yes {
		return nil
	}
	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		return fmt.Errorf("remove requires --yes when not running interactively")
	}
	fmt.Fprintf(cmd.OutOrStdout(), "确认删除《%s》（Book ID：%d）。请输入 Book ID 继续：", book.Title, book.ID)
	scanner := bufio.NewScanner(cmd.InOrStdin())
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return err
		}
		return fmt.Errorf("remove cancelled")
	}
	if !removeConfirmationMatches(scanner.Text(), book.ID) {
		return fmt.Errorf("remove cancelled")
	}
	return nil
}

func removeConfirmationMatches(input string, bookID int64) bool {
	confirmed, err := parseIntArg(input, "confirmation")
	return err == nil && confirmed == bookID
}

func init() {
	removeCmd.Flags().BoolVar(&removeYes, "yes", false, "confirm removal without prompting")
	rootCmd.AddCommand(removeCmd)
}
