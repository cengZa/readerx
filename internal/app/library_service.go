package app

import (
	"github.com/heybox/readerx/internal/domain"
	"github.com/heybox/readerx/internal/storage"
)

type LibraryService struct {
	store storage.Store
}

func NewLibraryService(store storage.Store) *LibraryService {
	return &LibraryService{store: store}
}

func (s *LibraryService) ListBooks() ([]domain.Book, error) {
	return s.store.ListBooks()
}
