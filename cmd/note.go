package cmd

import (
	"fmt"
	"text/tabwriter"
	"time"

	"github.com/heybox/readerx/internal/app"
	"github.com/spf13/cobra"
)

var noteBookID int64
var noteContent string

var noteCmd = &cobra.Command{
	Use:   "note",
	Short: "Manage notes",
}

var noteAddCmd = &cobra.Command{
	Use:   "add <book-id>",
	Short: "Add a note at the saved reading position",
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

		note, err := app.NewNoteService(store).AddCurrent(bookID, noteContent)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "笔记已添加：ID=%d Book=%d Chapter=%d\n%s\n", note.ID, note.BookID, note.ChapterNo, note.Content)
		return nil
	},
}

var noteListCmd = &cobra.Command{
	Use:   "list",
	Short: "List notes",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := openStore()
		if err != nil {
			return err
		}
		defer store.Close()

		notes, err := app.NewNoteService(store).List(noteBookID)
		if err != nil {
			return err
		}
		if len(notes) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "暂无笔记。")
			return nil
		}
		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tBook\tChapter\tUpdated\tContent")
		for _, note := range notes {
			updated := "-"
			if note.UpdatedAt > 0 {
				updated = time.Unix(note.UpdatedAt, 0).Format("2006-01-02")
			}
			fmt.Fprintf(w, "%d\t%s\t%d %s\t%s\t%s\n",
				note.ID, note.BookTitle, note.ChapterNo, note.ChapterTitle, updated, note.Content)
		}
		return w.Flush()
	},
}

var noteRemoveCmd = &cobra.Command{
	Use:   "remove <note-id>",
	Short: "Remove a note",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		noteID, err := parseIntArg(args[0], "note-id")
		if err != nil {
			return err
		}
		store, err := openStore()
		if err != nil {
			return err
		}
		defer store.Close()

		if err := app.NewNoteService(store).Remove(noteID); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "笔记已删除：ID=%d\n", noteID)
		return nil
	},
}

func init() {
	noteAddCmd.Flags().StringVar(&noteContent, "content", "", "note content")
	noteAddCmd.MarkFlagRequired("content")
	noteListCmd.Flags().Int64Var(&noteBookID, "book", 0, "limit notes to a book id")
	noteCmd.AddCommand(noteAddCmd, noteListCmd, noteRemoveCmd)
	rootCmd.AddCommand(noteCmd)
}
