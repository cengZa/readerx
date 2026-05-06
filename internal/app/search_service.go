package app

import (
	"fmt"
	"strings"

	"github.com/heybox/readerx/internal/domain"
	"github.com/heybox/readerx/internal/storage"
)

type SearchService struct {
	store storage.Store
}

func NewSearchService(store storage.Store) *SearchService {
	return &SearchService{store: store}
}

func (s *SearchService) Search(keyword string, bookID int64, limit int) ([]domain.SearchResult, error) {
	if strings.TrimSpace(keyword) == "" {
		return nil, fmt.Errorf("keyword must not be empty")
	}
	if limit <= 0 {
		limit = 50
	}
	return s.store.SearchChapters(keyword, bookID, limit)
}
