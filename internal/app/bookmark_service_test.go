package app

import (
	"testing"

	"github.com/heybox/readerx/internal/domain"
	"github.com/heybox/readerx/internal/storage"
)

func TestBookmarkServiceAddsAtSavedProgress(t *testing.T) {
	store, bookID := seedBook(t)
	defer store.Close()
	if err := store.SaveProgress(domain.Progress{BookID: bookID, ChapterNo: 1, LineOffset: 0, CharOffset: 0}); err != nil {
		t.Fatalf("SaveProgress: %v", err)
	}

	bookmark, err := NewBookmarkService(store).AddCurrent(bookID, "重点")
	if err != nil {
		t.Fatalf("AddCurrent: %v", err)
	}
	if bookmark.ID == 0 || bookmark.Excerpt == "" || bookmark.Note != "重点" {
		t.Fatalf("bookmark = %#v", bookmark)
	}
}

func TestBookmarkServiceListsAndRemoves(t *testing.T) {
	store, bookID := seedBook(t)
	defer store.Close()
	service := NewBookmarkService(store)

	bookmark, err := service.AddAt(domain.Bookmark{BookID: bookID, ChapterNo: 1, Excerpt: "摘录"})
	if err != nil {
		t.Fatalf("AddAt: %v", err)
	}
	bookmarks, err := service.List(bookID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(bookmarks) != 1 {
		t.Fatalf("bookmarks = %#v", bookmarks)
	}
	if err := service.Remove(bookmark.ID); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	bookmarks, err = service.List(bookID)
	if err != nil {
		t.Fatalf("List after remove: %v", err)
	}
	if len(bookmarks) != 0 {
		t.Fatalf("bookmarks after remove = %#v", bookmarks)
	}
}

func seedBook(t *testing.T) (*storage.SQLiteStore, int64) {
	t.Helper()
	store, err := storage.OpenSQLite(t.TempDir() + "/reader.db")
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	bookID, err := store.CreateBook(domain.Book{SourceType: "local_txt", Title: "测试书", ContentHash: "hash"})
	if err != nil {
		t.Fatalf("CreateBook: %v", err)
	}
	if err := store.InsertChapters(bookID, []domain.Chapter{{ChapterNo: 1, Title: "第一章", Content: "这是用于摘录的章节内容。", WordCount: 12}}); err != nil {
		t.Fatalf("InsertChapters: %v", err)
	}
	return store, bookID
}
