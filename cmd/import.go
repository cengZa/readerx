package cmd

import (
	"fmt"

	"github.com/heybox/readerx/internal/app"
	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Import a local text file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := openStore()
		if err != nil {
			return err
		}
		defer store.Close()

		result, err := app.NewImportService(store).ImportFile(args[0])
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "导入成功：\n书名：%s\n章节数：%d\n总字数：%d\nBook ID：%d\n", result.Title, result.ChapterCount, result.WordCount, result.BookID)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(importCmd)
}
