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

func (s *SearchService) Search(keyword string, bookID int64) ([]domain.SearchResult, error) {
	if strings.TrimSpace(keyword) == "" {
		return nil, fmt.Errorf("keyword must not be empty")
	}
	return s.store.SearchChapters(keyword, bookID)
}
