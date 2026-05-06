package app

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"

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

func (s *BookmarkService) ExportMarkdown(bookID int64) (string, error) {
	bookmarks, err := s.store.ListBookmarks(bookID)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	buf.WriteString("# Bookmarks\n\n")
	if len(bookmarks) == 0 {
		buf.WriteString("_No bookmarks._\n")
		return buf.String(), nil
	}
	for _, bookmark := range bookmarks {
		fmt.Fprintf(&buf, "## %s / Chapter %d %s\n\n", bookmark.BookTitle, bookmark.ChapterNo, bookmark.ChapterTitle)
		if bookmark.CreatedAt > 0 {
			fmt.Fprintf(&buf, "- Created: %s\n", time.Unix(bookmark.CreatedAt, 0).Format("2006-01-02 15:04"))
		}
		fmt.Fprintf(&buf, "- Location: line %d, char %d\n\n", bookmark.LineOffset, bookmark.CharOffset)
		if strings.TrimSpace(bookmark.Excerpt) != "" {
			fmt.Fprintf(&buf, "> %s\n\n", strings.ReplaceAll(strings.TrimSpace(bookmark.Excerpt), "\n", "\n> "))
		}
		if strings.TrimSpace(bookmark.Note) != "" {
			fmt.Fprintf(&buf, "Note: %s\n\n", strings.TrimSpace(bookmark.Note))
		}
	}
	return buf.String(), nil
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
