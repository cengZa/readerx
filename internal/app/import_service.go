package app

import (
	"fmt"

	"github.com/heybox/readerx/internal/source"
	"github.com/heybox/readerx/internal/storage"
)

type ImportService struct {
	store storage.Store
}

type ImportResult struct {
	BookID       int64
	Title        string
	ChapterCount int
	WordCount    int
}

func NewImportService(store storage.Store) *ImportService {
	return &ImportService{store: store}
}

func (s *ImportService) ImportFile(path string) (ImportResult, error) {
	book, chapters, err := source.ImportLocalFile(path)
	if err != nil {
		return ImportResult{}, err
	}
	if len(chapters) == 0 {
		return ImportResult{}, fmt.Errorf("no chapters parsed from %s", path)
	}
	bookID, err := s.store.CreateBook(book)
	if err != nil {
		return ImportResult{}, err
	}
	if err := s.store.InsertChapters(bookID, chapters); err != nil {
		return ImportResult{}, err
	}
	totalWords := 0
	for _, chapter := range chapters {
		totalWords += chapter.WordCount
	}
	return ImportResult{BookID: bookID, Title: book.Title, ChapterCount: len(chapters), WordCount: totalWords}, nil
}
