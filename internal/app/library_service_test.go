package app

import (
	"testing"

	"github.com/heybox/readerx/internal/domain"
	"github.com/heybox/readerx/internal/storage"
)

func TestLibraryServiceBookInfo(t *testing.T) {
	store, bookID := seedLibraryBook(t)
	defer store.Close()

	info, err := NewLibraryService(store).BookInfo(bookID)
	if err != nil {
		t.Fatalf("BookInfo: %v", err)
	}
	if info.Book.ID != bookID || info.Book.Title != "剑来" {
		t.Fatalf("book info = %#v", info)
	}
	if info.ChapterCount != 2 || info.WordCount != 10 {
		t.Fatalf("counts = chapters %d words %d, want 2 and 10", info.ChapterCount, info.WordCount)
	}
}

func TestLibraryServiceRemoveBook(t *testing.T) {
	store, bookID := seedLibraryBook(t)
	defer store.Close()

	removed, err := NewLibraryService(store).RemoveBook(bookID)
	if err != nil {
		t.Fatalf("RemoveBook: %v", err)
	}
	if removed.ID != bookID || removed.Title != "剑来" {
		t.Fatalf("removed = %#v", removed)
	}
	if _, err := store.GetBook(bookID); err != storage.ErrNotFound {
		t.Fatalf("GetBook after remove err = %v, want ErrNotFound", err)
	}
}

func seedLibraryBook(t *testing.T) (*storage.SQLiteStore, int64) {
	t.Helper()
	store, err := storage.OpenSQLite(t.TempDir() + "/reader.db")
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	bookID, err := store.CreateBook(domain.Book{
		SourceType:  "local_txt",
		Title:       "剑来",
		FilePath:    "/tmp/book.txt",
		ContentHash: "library-seed",
	})
	if err != nil {
		t.Fatalf("CreateBook: %v", err)
	}
	err = store.InsertChapters(bookID, []domain.Chapter{
		{ChapterNo: 1, Title: "第一章", Content: "第一段内容", WordCount: 5},
		{ChapterNo: 2, Title: "第二章", Content: "第二段内容", WordCount: 5},
	})
	if err != nil {
		t.Fatalf("InsertChapters: %v", err)
	}
	return store, bookID
}
