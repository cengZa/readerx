package source

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/heybox/readerx/internal/domain"
)

func ImportLocalFile(path string) (domain.Book, []domain.Chapter, error) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".txt":
		return LocalTxtSource{}.ImportFile(path)
	case ".epub":
		return LocalEpubSource{}.ImportFile(path)
	default:
		return domain.Book{}, nil, fmt.Errorf("unsupported file type %q; supported: .txt, .epub", filepath.Ext(path))
	}
}
