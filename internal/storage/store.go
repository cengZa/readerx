package storage

import "github.com/heybox/readerx/internal/domain"

type Store interface {
	CreateBook(book domain.Book) (int64, error)
	InsertChapters(bookID int64, chapters []domain.Chapter) error
	ReplaceBook(bookID int64, book domain.Book, chapters []domain.Chapter) error
	ListBooks() ([]domain.Book, error)
	GetBook(bookID int64) (domain.Book, error)
	GetBookByContentHash(contentHash string) (domain.Book, error)
	ListChapters(bookID int64) ([]domain.Chapter, error)
	GetChapter(bookID int64, chapterNo int) (domain.Chapter, error)
	CountChapters(bookID int64) (int, error)
	SaveProgress(progress domain.Progress) error
	GetProgress(bookID int64) (domain.Progress, error)
	GetRecentProgress() (domain.Progress, error)
	UpdateLastRead(bookID int64) error
	AddBookmark(bookmark domain.Bookmark) (int64, error)
	ListBookmarks(bookID int64) ([]domain.Bookmark, error)
	RemoveBookmark(bookmarkID int64) error
	AddNote(note domain.Note) (int64, error)
	ListNotes(bookID int64) ([]domain.Note, error)
	RemoveNote(noteID int64) error
	SearchChapters(keyword string, bookID int64, limit int) ([]domain.SearchResult, error)
	SetSetting(key, value string) error
	GetSetting(key string) (string, error)
	ListSettings() ([]domain.Setting, error)
	Close() error
}
