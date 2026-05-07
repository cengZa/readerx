package app

import (
	"errors"
	"fmt"

	"github.com/heybox/readerx/internal/domain"
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
	Existing     bool
	Replaced     bool
	Warnings     []string
}

type ImportOptions struct {
	Replace bool
}

func NewImportService(store storage.Store) *ImportService {
	return &ImportService{store: store}
}

func (s *ImportService) ImportFile(path string) (ImportResult, error) {
	return s.ImportFileWithOptions(path, ImportOptions{})
}

func (s *ImportService) ImportFileWithOptions(path string, options ImportOptions) (ImportResult, error) {
	book, chapters, err := source.ImportLocalFile(path)
	if err != nil {
		return ImportResult{}, err
	}
	if len(chapters) == 0 {
		return ImportResult{}, fmt.Errorf("no chapters parsed from %s", path)
	}
	warnings := ChapterQualityWarnings(chapters)
	existing, err := s.store.GetBookByContentHash(book.ContentHash)
	if err == nil {
		if options.Replace {
			return s.replaceExistingBook(existing.ID, book, chapters, warnings)
		}
		result, err := s.importResultForExistingBook(existing)
		if err != nil {
			return ImportResult{}, err
		}
		result.Warnings = warnings
		return result, nil
	}
	if !errors.Is(err, storage.ErrNotFound) {
		return ImportResult{}, err
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
	return ImportResult{BookID: bookID, Title: book.Title, ChapterCount: len(chapters), WordCount: totalWords, Warnings: warnings}, nil
}

func (s *ImportService) replaceExistingBook(bookID int64, book domain.Book, chapters []domain.Chapter, warnings []string) (ImportResult, error) {
	if err := s.store.ReplaceBook(bookID, book, chapters); err != nil {
		return ImportResult{}, err
	}
	totalWords := 0
	for _, chapter := range chapters {
		totalWords += chapter.WordCount
	}
	return ImportResult{BookID: bookID, Title: book.Title, ChapterCount: len(chapters), WordCount: totalWords, Replaced: true, Warnings: warnings}, nil
}

func (s *ImportService) importResultForExistingBook(book domain.Book) (ImportResult, error) {
	chapters, err := s.store.ListChapters(book.ID)
	if err != nil {
		return ImportResult{}, err
	}
	totalWords := 0
	for _, chapter := range chapters {
		totalWords += chapter.WordCount
	}
	return ImportResult{BookID: book.ID, Title: book.Title, ChapterCount: len(chapters), WordCount: totalWords, Existing: true}, nil
}
