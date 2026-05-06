package app

import (
	"github.com/heybox/readerx/internal/domain"
	"github.com/heybox/readerx/internal/storage"
)

type ChapterService struct {
	store storage.Store
}

func NewChapterService(store storage.Store) *ChapterService {
	return &ChapterService{store: store}
}

func (s *ChapterService) List(bookID int64) ([]domain.Chapter, error) {
	return s.store.ListChapters(bookID)
}
