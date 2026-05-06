package cmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/heybox/readerx/internal/app"
	"github.com/spf13/cobra"
)

var chaptersCmd = &cobra.Command{
	Use:   "chapters <book-id>",
	Short: "List chapters for a book",
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

		chapters, err := app.NewChapterService(store).List(bookID)
		if err != nil {
			return err
		}
		if len(chapters) == 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "Book ID=%d 暂无章节。\n", bookID)
			return nil
		}

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "No\tTitle\tWords")
		for _, chapter := range chapters {
			fmt.Fprintf(w, "%d\t%s\t%d\n", chapter.ChapterNo, chapter.Title, chapter.WordCount)
		}
		return w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(chaptersCmd)
}
