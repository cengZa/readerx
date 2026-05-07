package app

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/heybox/readerx/internal/domain"
	"github.com/heybox/readerx/internal/storage"
)

type LibraryService struct {
	store storage.Store
}

type BookInfo struct {
	Book          domain.Book
	ChapterCount  int
	WordCount     int
	Progress      domain.Progress
	BookmarkCount int
	NoteCount     int
}

type ListBooksOptions struct {
	Filter string
	Sort   string
}

func NewLibraryService(store storage.Store) *LibraryService {
	return &LibraryService{store: store}
}

func (s *LibraryService) ListBooks() ([]domain.Book, error) {
	return s.ListBooksWithOptions(ListBooksOptions{})
}

func (s *LibraryService) ListBooksWithOptions(options ListBooksOptions) ([]domain.Book, error) {
	books, err := s.store.ListBooks()
	if err != nil {
		return nil, err
	}
	return filterAndSortBooks(books, options)
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
	progress, err := s.store.GetProgress(bookID)
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		return BookInfo{}, err
	}
	bookmarks, err := s.store.ListBookmarks(bookID)
	if err != nil {
		return BookInfo{}, err
	}
	notes, err := s.store.ListNotes(bookID)
	if err != nil {
		return BookInfo{}, err
	}
	return BookInfo{
		Book:          book,
		ChapterCount:  len(chapters),
		WordCount:     wordCount,
		Progress:      progress,
		BookmarkCount: len(bookmarks),
		NoteCount:     len(notes),
	}, nil
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

func filterAndSortBooks(books []domain.Book, options ListBooksOptions) ([]domain.Book, error) {
	filter := strings.ToLower(strings.TrimSpace(options.Filter))
	result := make([]domain.Book, 0, len(books))
	for _, book := range books {
		if filter == "" || strings.Contains(strings.ToLower(book.Title), filter) {
			result = append(result, book)
		}
	}

	sortMode := strings.TrimSpace(options.Sort)
	if sortMode == "" {
		sortMode = "recent"
	}
	switch sortMode {
	case "recent":
		sort.SliceStable(result, func(i, j int) bool {
			leftRead := result[i].LastReadAt
			rightRead := result[j].LastReadAt
			if leftRead > 0 && rightRead == 0 {
				return true
			}
			if leftRead == 0 && rightRead > 0 {
				return false
			}
			if leftRead > 0 || rightRead > 0 {
				return leftRead > rightRead
			}
			return result[i].CreatedAt > result[j].CreatedAt
		})
	case "created":
		sort.SliceStable(result, func(i, j int) bool {
			return result[i].CreatedAt > result[j].CreatedAt
		})
	case "title":
		sort.SliceStable(result, func(i, j int) bool {
			return strings.ToLower(result[i].Title) < strings.ToLower(result[j].Title)
		})
	default:
		return nil, fmt.Errorf("unsupported list sort %q; supported: recent, created, title", sortMode)
	}
	return result, nil
}
