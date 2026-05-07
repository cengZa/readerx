package cmd

import (
	"fmt"
	"time"

	"github.com/heybox/readerx/internal/app"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info <book-id>",
	Short: "Show book details",
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

		info, err := app.NewLibraryService(store).BookInfo(bookID)
		if err != nil {
			return err
		}
		book := info.Book
		lastRead := "-"
		if book.LastReadAt > 0 {
			lastRead = time.Unix(book.LastReadAt, 0).Format("2006-01-02 15:04")
		}
		progress := "-"
		if info.Progress.BookID > 0 {
			progress = fmt.Sprintf("第 %d 章，%.0f%%", info.Progress.ChapterNo, info.Progress.Percentage)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Book ID：%d\n书名：%s\n作者：%s\n来源：%s\n文件：%s\n章节数：%d\n总字数：%d\n阅读进度：%s\n书签数：%d\n笔记数：%d\n最后阅读：%s\n",
			book.ID, book.Title, emptyAsDash(book.Author), book.SourceType, emptyAsDash(book.FilePath), info.ChapterCount, info.WordCount,
			progress, info.BookmarkCount, info.NoteCount, lastRead)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
