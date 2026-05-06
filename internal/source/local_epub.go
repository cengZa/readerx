package source

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/heybox/readerx/internal/domain"
	"github.com/heybox/readerx/internal/parser"
)

type LocalEpubSource struct{}

func (LocalEpubSource) ImportFile(path string) (domain.Book, []domain.Chapter, error) {
	if strings.ToLower(filepath.Ext(path)) != ".epub" {
		return domain.Book{}, nil, fmt.Errorf("unsupported file type %q for EPUB source", filepath.Ext(path))
	}
	return parser.ParseEPUBFile(path)
}
