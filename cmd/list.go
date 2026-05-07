package cmd

import (
	"fmt"
	"text/tabwriter"
	"time"

	"github.com/heybox/readerx/internal/app"
	"github.com/spf13/cobra"
)

var listSort string
var listFilter string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List imported books",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := openStore()
		if err != nil {
			return err
		}
		defer store.Close()

		books, err := app.NewLibraryService(store).ListBooksWithOptions(app.ListBooksOptions{Filter: listFilter, Sort: listSort})
		if err != nil {
			return err
		}
		if len(books) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "书架为空，请先使用 readerx import <file> 导入书籍。")
			return nil
		}

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tTitle\tChapters\tProgress\tLast Read")
		for _, book := range books {
			lastRead := "-"
			if book.LastReadAt > 0 {
				lastRead = time.Unix(book.LastReadAt, 0).Format("2006-01-02")
			}
			fmt.Fprintf(w, "%d\t%s\t%d\t%.0f%%\t%s\n", book.ID, book.Title, book.ChapterCount, book.ProgressPercentage, lastRead)
		}
		return w.Flush()
	},
}

func init() {
	listCmd.Flags().StringVar(&listSort, "sort", "recent", "sort books by recent, created, or title")
	listCmd.Flags().StringVar(&listFilter, "filter", "", "filter books by title")
	rootCmd.AddCommand(listCmd)
}
