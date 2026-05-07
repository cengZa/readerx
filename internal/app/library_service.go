package app

import (
	"github.com/heybox/readerx/internal/domain"
	"github.com/heybox/readerx/internal/storage"
)

type LibraryService struct {
	store storage.Store
}

type BookInfo struct {
	Book         domain.Book
	ChapterCount int
	WordCount    int
}

func NewLibraryService(store storage.Store) *LibraryService {
	return &LibraryService{store: store}
}

func (s *LibraryService) ListBooks() ([]domain.Book, error) {
	return s.store.ListBooks()
}

func (s *LibraryService) BookInfo(bookID int64) (BookInfo, error) {
	book, err := s.store.GetBook(bookID)
	if err != nil {
		return BookInfo{}, err
	}
	chapters, err := s.store.ListChapters(bookID)
	if err != nil {
		return BookInfo{}, err
	}
	wordCount := 0
	for _, chapter := range chapters {
		wordCount += chapter.WordCount
	}
	return BookInfo{Book: book, ChapterCount: len(chapters), WordCount: wordCount}, nil
}

func (s *LibraryService) RemoveBook(bookID int64) (domain.Book, error) {
	book, err := s.store.GetBook(bookID)
	if err != nil {
		return domain.Book{}, err
	}
	if err := s.store.DeleteBook(bookID); err != nil {
		return domain.Book{}, err
	}
	return book, nil
}
