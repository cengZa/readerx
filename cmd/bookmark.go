package cmd

import (
	"fmt"
	"text/tabwriter"
	"time"

	"github.com/heybox/readerx/internal/app"
	"github.com/spf13/cobra"
)

var bookmarkBookID int64
var bookmarkNote string

var bookmarkCmd = &cobra.Command{
	Use:   "bookmark",
	Short: "Manage bookmarks",
}

var bookmarkAddCmd = &cobra.Command{
	Use:   "add <book-id>",
	Short: "Add a bookmark at the saved reading position",
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

		bookmark, err := app.NewBookmarkService(store).AddCurrent(bookID, bookmarkNote)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "书签已添加：ID=%d Book=%d Chapter=%d\n%s\n", bookmark.ID, bookmark.BookID, bookmark.ChapterNo, bookmark.Excerpt)
		return nil
	},
}

var bookmarkListCmd = &cobra.Command{
	Use:   "list",
	Short: "List bookmarks",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := openStore()
		if err != nil {
			return err
		}
		defer store.Close()

		bookmarks, err := app.NewBookmarkService(store).List(bookmarkBookID)
		if err != nil {
			return err
		}
		if len(bookmarks) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "暂无书签。")
			return nil
		}
		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tBook\tChapter\tCreated\tExcerpt")
		for _, bookmark := range bookmarks {
			created := "-"
			if bookmark.CreatedAt > 0 {
				created = time.Unix(bookmark.CreatedAt, 0).Format("2006-01-02")
			}
			fmt.Fprintf(w, "%d\t%s\t%d %s\t%s\t%s\n",
				bookmark.ID, bookmark.BookTitle, bookmark.ChapterNo, bookmark.ChapterTitle, created, bookmark.Excerpt)
		}
		return w.Flush()
	},
}

var bookmarkRemoveCmd = &cobra.Command{
	Use:   "remove <bookmark-id>",
	Short: "Remove a bookmark",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		bookmarkID, err := parseIntArg(args[0], "bookmark-id")
		if err != nil {
			return err
		}
		store, err := openStore()
		if err != nil {
			return err
		}
		defer store.Close()

		if err := app.NewBookmarkService(store).Remove(bookmarkID); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "书签已删除：ID=%d\n", bookmarkID)
		return nil
	},
}

func init() {
	bookmarkAddCmd.Flags().StringVar(&bookmarkNote, "note", "", "bookmark note")
	bookmarkListCmd.Flags().Int64Var(&bookmarkBookID, "book", 0, "limit bookmarks to a book id")
	bookmarkCmd.AddCommand(bookmarkAddCmd, bookmarkListCmd, bookmarkRemoveCmd)
	rootCmd.AddCommand(bookmarkCmd)
}
