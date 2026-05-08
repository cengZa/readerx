package cmd

import (
	"fmt"

	"github.com/heybox/readerx/internal/app"
	"github.com/spf13/cobra"
)

var (
	sourceSearchSource string
	sourceSearchLimit  int
)

var sourceCmd = &cobra.Command{
	Use:   "source",
	Short: "Search legal open ebook sources",
}

var sourceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List supported open ebook sources",
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, source := range app.NewSourceService().ListSources() {
			fmt.Fprintln(cmd.OutOrStdout(), source)
		}
		return nil
	},
}

var sourceSearchCmd = &cobra.Command{
	Use:   "search <keyword>",
	Short: "Search Project Gutenberg OPDS for importable books",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		results, err := app.NewSourceService().Search(cmd.Context(), app.SourceSearchOptions{
			Source: sourceSearchSource,
			Query:  args[0],
			Limit:  sourceSearchLimit,
		})
		if err != nil {
			return err
		}
		if len(results) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "未找到可导入的 TXT 或 EPUB 结果。")
			return nil
		}
		for i, result := range results {
			fmt.Fprintf(cmd.OutOrStdout(), "%d. %s\n", i+1, result.Title)
			if result.Author != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "   作者：%s\n", result.Author)
			}
			if result.ID != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "   Source ID：%s\n", result.ID)
			}
			if result.EPUBURL != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "   EPUB：%s\n", result.EPUBURL)
			}
			if result.TextURL != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "   TXT：%s\n", result.TextURL)
			}
			fmt.Fprintln(cmd.OutOrStdout())
		}
		return nil
	},
}

func init() {
	sourceSearchCmd.Flags().StringVar(&sourceSearchSource, "source", "gutenberg", "open ebook source")
	sourceSearchCmd.Flags().IntVar(&sourceSearchLimit, "limit", 10, "maximum number of results")
	sourceCmd.AddCommand(sourceListCmd)
	sourceCmd.AddCommand(sourceSearchCmd)
	rootCmd.AddCommand(sourceCmd)
}
