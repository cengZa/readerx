package source

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/heybox/readerx/internal/domain"
	"github.com/heybox/readerx/internal/parser"
)

type LocalTxtSource struct{}

func (LocalTxtSource) ImportFile(path string) (domain.Book, []domain.Chapter, error) {
	if strings.ToLower(filepath.Ext(path)) != ".txt" {
		return domain.Book{}, nil, fmt.Errorf("unsupported file type %q; first MVP supports .txt only", filepath.Ext(path))
	}
	return parser.ParseTXTFile(path)
}
