package cmd

import (
	"fmt"

	"github.com/heybox/readerx/internal/app"
	"github.com/spf13/cobra"
)

var searchBookID int64

var searchCmd = &cobra.Command{
	Use:   "search <keyword>",
	Short: "Search imported books",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := openStore()
		if err != nil {
			return err
		}
		defer store.Close()

		results, err := app.NewSearchService(store).Search(args[0], searchBookID)
		if err != nil {
			return err
		}
		if len(results) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "没有找到匹配结果。")
			return nil
		}
		for _, result := range results {
			fmt.Fprintf(cmd.OutOrStdout(), "[Book %d] %s / 第 %d 章 %s\n%s\n\n",
				result.BookID, result.BookTitle, result.ChapterNo, result.ChapterTitle, result.Snippet)
		}
		return nil
	},
}

func init() {
	searchCmd.Flags().Int64Var(&searchBookID, "book", 0, "limit search to a book id")
	rootCmd.AddCommand(searchCmd)
}
