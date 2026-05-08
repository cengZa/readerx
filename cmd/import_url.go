package cmd

import (
	"github.com/heybox/readerx/internal/app"
	"github.com/spf13/cobra"
)

var (
	importURLReplace  bool
	importURLTitle    string
	importURLMaxBytes int64
)

var importURLCmd = &cobra.Command{
	Use:   "import-url <url>",
	Short: "Import a public TXT or EPUB URL",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := openStore()
		if err != nil {
			return err
		}
		defer store.Close()

		result, err := app.NewImportService(store).ImportURLWithOptions(args[0], app.ImportOptions{
			Replace:          importURLReplace,
			TitleOverride:    importURLTitle,
			MaxDownloadBytes: importURLMaxBytes,
		})
		if err != nil {
			return err
		}
		printImportResult(cmd, result)
		return nil
	},
}

func init() {
	importURLCmd.Flags().BoolVar(&importURLReplace, "replace", false, "replace an existing imported book with the same content")
	importURLCmd.Flags().StringVar(&importURLTitle, "title", "", "override the imported book title")
	importURLCmd.Flags().Int64Var(&importURLMaxBytes, "max-bytes", app.DefaultMaxDownloadBytes, "maximum download size in bytes")
	rootCmd.AddCommand(importURLCmd)
}
