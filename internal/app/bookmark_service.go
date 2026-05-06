package app

import (
	"errors"
	"fmt"
	"strings"

	"github.com/heybox/readerx/internal/domain"
	"github.com/heybox/readerx/internal/storage"
)

type BookmarkService struct {
	store storage.Store
}

func NewBookmarkService(store storage.Store) *BookmarkService {
	return &BookmarkService{store: store}
}

func (s *BookmarkService) AddCurrent(bookID int64, note string) (domain.Bookmark, error) {
	progress, err := s.store.GetProgress(bookID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			progress = domain.Progress{BookID: bookID, ChapterNo: 1}
		} else {
			return domain.Bookmark{}, err
		}
	}
	chapter, err := s.store.GetChapter(bookID, progress.ChapterNo)
	if err != nil {
		return domain.Bookmark{}, err
	}
	bookmark := domain.Bookmark{
		BookID:     bookID,
		ChapterNo:  progress.ChapterNo,
		LineOffset: progress.LineOffset,
		CharOffset: progress.CharOffset,
		Excerpt:    excerptAt(chapter.Content, progress.CharOffset, 36),
		Note:       note,
	}
	return s.AddAt(bookmark)
}

func (s *BookmarkService) AddAt(bookmark domain.Bookmark) (domain.Bookmark, error) {
	if bookmark.BookID <= 0 {
		return domain.Bookmark{}, fmt.Errorf("book id must be positive")
	}
	if bookmark.ChapterNo <= 0 {
		bookmark.ChapterNo = 1
	}
	if strings.TrimSpace(bookmark.Excerpt) == "" {
		chapter, err := s.store.GetChapter(bookmark.BookID, bookmark.ChapterNo)
		if err != nil {
			return domain.Bookmark{}, err
		}
		bookmark.Excerpt = excerptAt(chapter.Content, bookmark.CharOffset, 36)
	}
	id, err := s.store.AddBookmark(bookmark)
	if err != nil {
		return domain.Bookmark{}, err
	}
	bookmark.ID = id
	return bookmark, nil
}

func (s *BookmarkService) List(bookID int64) ([]domain.Bookmark, error) {
	return s.store.ListBookmarks(bookID)
}

func (s *BookmarkService) Remove(bookmarkID int64) error {
	return s.store.RemoveBookmark(bookmarkID)
}

func excerptAt(content string, charOffset, length int) string {
	runes := []rune(content)
	if len(runes) == 0 {
		return ""
	}
	if charOffset < 0 {
		charOffset = 0
	}
	if charOffset > len(runes) {
		charOffset = len(runes)
	}
	end := charOffset + length
	if end > len(runes) {
		end = len(runes)
	}
	return strings.TrimSpace(string(runes[charOffset:end]))
}
