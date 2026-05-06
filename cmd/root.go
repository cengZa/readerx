package cmd

import (
	"fmt"
	"os"

	"github.com/heybox/readerx/internal/config"
	"github.com/heybox/readerx/internal/storage"
	"github.com/spf13/cobra"
)

var dbPath string

var rootCmd = &cobra.Command{
	Use:   "readerx",
	Short: "A local-first terminal reader",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "", "SQLite database path")
}

func openStore() (*storage.SQLiteStore, error) {
	path := dbPath
	if path == "" {
		path = config.DefaultDBPath()
	}
	return storage.OpenSQLite(path)
}
