package cmd

import (
	"fmt"

	"github.com/heybox/readerx/internal/app"
	"github.com/spf13/cobra"
)

var searchBookID int64
var searchLimit int

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

		limit := searchLimit
		if !cmd.Flags().Changed("limit") {
			if configured, err := app.NewConfigService(store).Get("search.limit"); err == nil {
				if parsed, parseErr := parseIntArg(configured, "search.limit"); parseErr == nil {
					limit = int(parsed)
				}
			}
		}
		results, err := app.NewSearchService(store).Search(args[0], searchBookID, limit)
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
	searchCmd.Flags().IntVar(&searchLimit, "limit", 50, "maximum number of search results")
	rootCmd.AddCommand(searchCmd)
}
