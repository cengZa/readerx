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

func TestLibraryServiceBookInfoIncludesProgressBookmarksAndNotes(t *testing.T) {
	store, bookID := seedLibraryBook(t)
	defer store.Close()
	if err := store.SaveProgress(domain.Progress{BookID: bookID, ChapterNo: 2, LineOffset: 3, Percentage: 50}); err != nil {
		t.Fatalf("SaveProgress: %v", err)
	}
	if _, err := store.AddBookmark(domain.Bookmark{BookID: bookID, ChapterNo: 1, Excerpt: "摘录"}); err != nil {
		t.Fatalf("AddBookmark: %v", err)
	}
	if _, err := store.AddNote(domain.Note{BookID: bookID, ChapterNo: 2, Content: "笔记"}); err != nil {
		t.Fatalf("AddNote: %v", err)
	}

	info, err := NewLibraryService(store).BookInfo(bookID)
	if err != nil {
		t.Fatalf("BookInfo: %v", err)
	}
	if info.Progress.ChapterNo != 2 || info.Progress.Percentage != 50 {
		t.Fatalf("progress = %#v, want saved progress", info.Progress)
	}
	if info.BookmarkCount != 1 || info.NoteCount != 1 {
		t.Fatalf("bookmark/note count = %d/%d, want 1/1", info.BookmarkCount, info.NoteCount)
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

func TestListBooksWithOptionsFiltersAndSortsByTitle(t *testing.T) {
	books := []domain.Book{
		{ID: 1, Title: "Beta", CreatedAt: 20},
		{ID: 2, Title: "Alpha", CreatedAt: 10},
		{ID: 3, Title: "Other", CreatedAt: 30},
	}

	got, err := filterAndSortBooks(books, ListBooksOptions{Filter: "a", Sort: "title"})
	if err != nil {
		t.Fatalf("filterAndSortBooks: %v", err)
	}
	if len(got) != 2 || got[0].Title != "Alpha" || got[1].Title != "Beta" {
		t.Fatalf("books = %#v, want Alpha then Beta", got)
	}
}

func TestListBooksWithOptionsSortsByCreatedAndRecent(t *testing.T) {
	books := []domain.Book{
		{ID: 1, Title: "A", CreatedAt: 20, LastReadAt: 5},
		{ID: 2, Title: "B", CreatedAt: 30, LastReadAt: 0},
		{ID: 3, Title: "C", CreatedAt: 10, LastReadAt: 50},
	}

	created, err := filterAndSortBooks(books, ListBooksOptions{Sort: "created"})
	if err != nil {
		t.Fatalf("created sort: %v", err)
	}
	if created[0].ID != 2 || created[1].ID != 1 || created[2].ID != 3 {
		t.Fatalf("created order = %#v", created)
	}

	recent, err := filterAndSortBooks(books, ListBooksOptions{Sort: "recent"})
	if err != nil {
		t.Fatalf("recent sort: %v", err)
	}
	if recent[0].ID != 3 || recent[1].ID != 1 || recent[2].ID != 2 {
		t.Fatalf("recent order = %#v", recent)
	}
}

func TestListBooksWithOptionsRejectsUnknownSort(t *testing.T) {
	_, err := filterAndSortBooks(nil, ListBooksOptions{Sort: "unknown"})
	if err == nil {
		t.Fatalf("expected error for unknown sort")
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
