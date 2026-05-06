package config

import (
	"os"
	"path/filepath"
)

func DefaultDBPath() string {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "readerx", "reader.db")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "reader.db"
	}
	return filepath.Join(home, ".readerx", "reader.db")
}
