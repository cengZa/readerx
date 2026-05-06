package app

import (
	"testing"

	"github.com/heybox/readerx/internal/domain"
	"github.com/heybox/readerx/internal/storage"
)

func TestNoteServiceAddsAtSavedProgress(t *testing.T) {
	store, bookID := seedNoteBook(t)
	defer store.Close()
	if err := store.SaveProgress(domain.Progress{BookID: bookID, ChapterNo: 1, LineOffset: 1, CharOffset: 2}); err != nil {
		t.Fatalf("SaveProgress: %v", err)
	}

	note, err := NewNoteService(store).AddCurrent(bookID, "这里需要回看")
	if err != nil {
		t.Fatalf("AddCurrent: %v", err)
	}
	if note.ID == 0 || note.Content != "这里需要回看" || note.ChapterNo != 1 || note.CharOffset != 2 {
		t.Fatalf("note = %#v", note)
	}
}

func TestNoteServiceListsAndRemoves(t *testing.T) {
	store, bookID := seedNoteBook(t)
	defer store.Close()
	service := NewNoteService(store)

	note, err := service.AddAt(domain.Note{BookID: bookID, ChapterNo: 1, Content: "笔记"})
	if err != nil {
		t.Fatalf("AddAt: %v", err)
	}
	notes, err := service.List(bookID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(notes) != 1 || notes[0].BookTitle != "笔记书" {
		t.Fatalf("notes = %#v", notes)
	}
	if err := service.Remove(note.ID); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	notes, err = service.List(bookID)
	if err != nil {
		t.Fatalf("List after remove: %v", err)
	}
	if len(notes) != 0 {
		t.Fatalf("notes after remove = %#v", notes)
	}
}

func seedNoteBook(t *testing.T) (*storage.SQLiteStore, int64) {
	t.Helper()
	store, err := storage.OpenSQLite(t.TempDir() + "/reader.db")
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	bookID, err := store.CreateBook(domain.Book{SourceType: "local_txt", Title: "笔记书", ContentHash: "note-book"})
	if err != nil {
		t.Fatalf("CreateBook: %v", err)
	}
	if err := store.InsertChapters(bookID, []domain.Chapter{{ChapterNo: 1, Title: "第一章", Content: "章节内容", WordCount: 4}}); err != nil {
		t.Fatalf("InsertChapters: %v", err)
	}
	return store, bookID
}
