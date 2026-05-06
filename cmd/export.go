package cmd

import (
	"fmt"
	"os"

	"github.com/heybox/readerx/internal/app"
	"github.com/spf13/cobra"
)

var exportBookID int64
var exportOutput string

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export reading data",
}

var exportBookmarksCmd = &cobra.Command{
	Use:   "bookmarks",
	Short: "Export bookmarks as Markdown",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := openStore()
		if err != nil {
			return err
		}
		defer store.Close()

		markdown, err := app.NewBookmarkService(store).ExportMarkdown(exportBookID)
		if err != nil {
			return err
		}
		return writeExport(cmd, markdown)
	},
}

var exportNotesCmd = &cobra.Command{
	Use:   "notes",
	Short: "Export notes as Markdown",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := openStore()
		if err != nil {
			return err
		}
		defer store.Close()

		markdown, err := app.NewNoteService(store).ExportMarkdown(exportBookID)
		if err != nil {
			return err
		}
		return writeExport(cmd, markdown)
	},
}

func writeExport(cmd *cobra.Command, content string) error {
	if exportOutput == "" {
		fmt.Fprint(cmd.OutOrStdout(), content)
		return nil
	}
	if err := os.WriteFile(exportOutput, []byte(content), 0o644); err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "导出成功：%s\n", exportOutput)
	return nil
}

func init() {
	exportBookmarksCmd.Flags().Int64Var(&exportBookID, "book", 0, "limit export to a book id")
	exportBookmarksCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "write Markdown to file")
	exportNotesCmd.Flags().Int64Var(&exportBookID, "book", 0, "limit export to a book id")
	exportNotesCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "write Markdown to file")
	exportCmd.AddCommand(exportBookmarksCmd, exportNotesCmd)
	rootCmd.AddCommand(exportCmd)
}
