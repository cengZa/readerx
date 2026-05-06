package cmd

import (
	"strconv"

	"github.com/heybox/readerx/internal/app"
	"github.com/heybox/readerx/internal/storage"
	"github.com/heybox/readerx/internal/tui"
)

func readTUIOptions(store storage.Store) tui.ReaderOptions {
	config := app.NewConfigService(store)
	options := tui.ReaderOptions{}
	if value, err := config.Get("reader.width"); err == nil {
		if parsed, parseErr := strconv.Atoi(value); parseErr == nil {
			options.MaxWidth = parsed
		}
	}
	if value, err := config.Get("reader.theme"); err == nil {
		options.Theme = value
	}
	return options
}
