package cmd

import (
	"fmt"

	"github.com/heybox/readerx/internal/app"
	"github.com/spf13/cobra"
)

var importReplace bool

var importCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Import a local TXT or EPUB file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := openStore()
		if err != nil {
			return err
		}
		defer store.Close()

		results, err := app.NewImportService(store).ImportPathWithOptions(args[0], app.ImportOptions{Replace: importReplace})
		if err != nil {
			return err
		}
		if len(results) == 1 {
			printImportResult(cmd, results[0])
			return nil
		}
		printBatchImportResults(cmd, results)
		return nil
	},
}

func printImportResult(cmd *cobra.Command, result app.ImportResult) {
	if result.Replaced {
		fmt.Fprintf(cmd.OutOrStdout(), "重新导入成功：\n书名：%s\n章节数：%d\n总字数：%d\nBook ID：%d\n", result.Title, result.ChapterCount, result.WordCount, result.BookID)
		printImportWarnings(cmd, result.Warnings)
		return
	}
	if result.Existing {
		fmt.Fprintf(cmd.OutOrStdout(), "书籍已存在：\n书名：%s\n章节数：%d\n总字数：%d\nBook ID：%d\n", result.Title, result.ChapterCount, result.WordCount, result.BookID)
		printImportWarnings(cmd, result.Warnings)
		return
	}
	fmt.Fprintf(cmd.OutOrStdout(), "导入成功：\n书名：%s\n章节数：%d\n总字数：%d\nBook ID：%d\n", result.Title, result.ChapterCount, result.WordCount, result.BookID)
	printImportWarnings(cmd, result.Warnings)
}

func printBatchImportResults(cmd *cobra.Command, results []app.ImportResult) {
	imported, existing, replaced := 0, 0, 0
	for _, result := range results {
		switch {
		case result.Replaced:
			replaced++
		case result.Existing:
			existing++
		default:
			imported++
		}
	}
	fmt.Fprintf(cmd.OutOrStdout(), "批量导入完成：新增 %d，本已存在 %d，重新导入 %d，总计 %d\n", imported, existing, replaced, len(results))
	for _, result := range results {
		status := "新增"
		if result.Existing {
			status = "已存在"
		}
		if result.Replaced {
			status = "重新导入"
		}
		fmt.Fprintf(cmd.OutOrStdout(), "- [%s] %s（Book ID：%d，章节数：%d，总字数：%d）\n", status, result.Title, result.BookID, result.ChapterCount, result.WordCount)
		printImportWarnings(cmd, result.Warnings)
	}
}

func printImportWarnings(cmd *cobra.Command, warnings []string) {
	if len(warnings) == 0 {
		return
	}
	fmt.Fprintln(cmd.OutOrStdout(), "\n导入质量提示：")
	for _, warning := range warnings {
		fmt.Fprintf(cmd.OutOrStdout(), "- %s\n", warning)
	}
}

func init() {
	importCmd.Flags().BoolVar(&importReplace, "replace", false, "replace an existing imported book with the same content")
	rootCmd.AddCommand(importCmd)
}
