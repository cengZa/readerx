package storage

import (
	"testing"

	"github.com/heybox/readerx/internal/domain"
)

func TestSQLiteStoreCreatesAndReadsBookWithChapters(t *testing.T) {
	store, err := OpenSQLite(t.TempDir() + "/reader.db")
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	defer store.Close()

	bookID, err := store.CreateBook(domain.Book{
		SourceType:  "local_txt",
		Title:       "剑来",
		FilePath:    "/tmp/book.txt",
		ContentHash: "hash-1",
	})
	if err != nil {
		t.Fatalf("CreateBook: %v", err)
	}
	err = store.InsertChapters(bookID, []domain.Chapter{
		{ChapterNo: 1, Title: "第一章", Content: "第一段", WordCount: 3},
		{ChapterNo: 2, Title: "第二章", Content: "第二段", WordCount: 3},
	})
	if err != nil {
		t.Fatalf("InsertChapters: %v", err)
	}

	books, err := store.ListBooks()
	if err != nil {
		t.Fatalf("ListBooks: %v", err)
	}
	if len(books) != 1 || books[0].ChapterCount != 2 {
		t.Fatalf("books = %#v, want one book with 2 chapters", books)
	}

	chapter, err := store.GetChapter(bookID, 2)
	if err != nil {
		t.Fatalf("GetChapter: %v", err)
	}
	if chapter.Title != "第二章" || chapter.Content != "第二段" {
		t.Fatalf("chapter = %#v", chapter)
	}

	chapters, err := store.ListChapters(bookID)
	if err != nil {
		t.Fatalf("ListChapters: %v", err)
	}
	if len(chapters) != 2 {
		t.Fatalf("chapter count = %d, want 2", len(chapters))
	}
	if chapters[0].Content != "" {
		t.Fatalf("ListChapters should not load full content, got %q", chapters[0].Content)
	}
	if chapters[1].Title != "第二章" || chapters[1].WordCount != 3 {
		t.Fatalf("chapter metadata = %#v", chapters[1])
	}
}

func TestSQLiteStoreRejectsDuplicateContentHash(t *testing.T) {
	store, err := OpenSQLite(t.TempDir() + "/reader.db")
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	defer store.Close()

	_, err = store.CreateBook(domain.Book{SourceType: "local_txt", Title: "A", ContentHash: "same"})
	if err != nil {
		t.Fatalf("CreateBook first: %v", err)
	}
	_, err = store.CreateBook(domain.Book{SourceType: "local_txt", Title: "B", ContentHash: "same"})
	if err == nil {
		t.Fatalf("CreateBook duplicate hash returned nil error")
	}
}

func TestSQLiteStoreManagesBookmarks(t *testing.T) {
	store, bookID := seedBookWithChapters(t)
	defer store.Close()

	bookmarkID, err := store.AddBookmark(domain.Bookmark{
		BookID:     bookID,
		ChapterNo:  1,
		LineOffset: 2,
		CharOffset: 6,
		Excerpt:    "摘录",
		Note:       "备注",
	})
	if err != nil {
		t.Fatalf("AddBookmark: %v", err)
	}
	bookmarks, err := store.ListBookmarks(bookID)
	if err != nil {
		t.Fatalf("ListBookmarks: %v", err)
	}
	if len(bookmarks) != 1 || bookmarks[0].ID != bookmarkID || bookmarks[0].BookTitle != "剑来" {
		t.Fatalf("bookmarks = %#v", bookmarks)
	}
	if err := store.RemoveBookmark(bookmarkID); err != nil {
		t.Fatalf("RemoveBookmark: %v", err)
	}
	bookmarks, err = store.ListBookmarks(bookID)
	if err != nil {
		t.Fatalf("ListBookmarks after remove: %v", err)
	}
	if len(bookmarks) != 0 {
		t.Fatalf("bookmarks after remove = %#v, want empty", bookmarks)
	}
}

func TestSQLiteStoreSearchesChapters(t *testing.T) {
	store, bookID := seedBookWithChapters(t)
	defer store.Close()

	results, err := store.SearchChapters("第二段", 0)
	if err != nil {
		t.Fatalf("SearchChapters: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("result count = %d, want 1: %#v", len(results), results)
	}
	if results[0].BookID != bookID || results[0].ChapterNo != 2 || results[0].Snippet == "" {
		t.Fatalf("result = %#v", results[0])
	}

	results, err = store.SearchChapters("第一段", bookID+100)
	if err != nil {
		t.Fatalf("SearchChapters scoped: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("scoped results = %#v, want empty", results)
	}
}

func TestSQLiteStoreManagesNotes(t *testing.T) {
	store, bookID := seedBookWithChapters(t)
	defer store.Close()

	noteID, err := store.AddNote(domain.Note{BookID: bookID, ChapterNo: 1, LineOffset: 1, CharOffset: 2, Content: "笔记"})
	if err != nil {
		t.Fatalf("AddNote: %v", err)
	}
	notes, err := store.ListNotes(bookID)
	if err != nil {
		t.Fatalf("ListNotes: %v", err)
	}
	if len(notes) != 1 || notes[0].ID != noteID || notes[0].BookTitle != "剑来" || notes[0].ChapterTitle != "第一章" {
		t.Fatalf("notes = %#v", notes)
	}
	if err := store.RemoveNote(noteID); err != nil {
		t.Fatalf("RemoveNote: %v", err)
	}
	notes, err = store.ListNotes(bookID)
	if err != nil {
		t.Fatalf("ListNotes after remove: %v", err)
	}
	if len(notes) != 0 {
		t.Fatalf("notes after remove = %#v, want empty", notes)
	}
}

func seedBookWithChapters(t *testing.T) (*SQLiteStore, int64) {
	t.Helper()
	store, err := OpenSQLite(t.TempDir() + "/reader.db")
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	bookID, err := store.CreateBook(domain.Book{
		SourceType:  "local_txt",
		Title:       "剑来",
		FilePath:    "/tmp/book.txt",
		ContentHash: "hash-seed",
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
