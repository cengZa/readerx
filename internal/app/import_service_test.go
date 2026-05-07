package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/heybox/readerx/internal/domain"
	"github.com/heybox/readerx/internal/storage"
)

func TestImportServiceImportsTXTIntoStore(t *testing.T) {
	store, err := storage.OpenSQLite(t.TempDir() + "/reader.db")
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	defer store.Close()

	path := filepath.Join(t.TempDir(), "book.txt")
	if err := os.WriteFile(path, []byte("第一章 开始\n内容一\n\n第二章 继续\n内容二"), 0o644); err != nil {
		t.Fatalf("write txt: %v", err)
	}

	result, err := NewImportService(store).ImportFile(path)
	if err != nil {
		t.Fatalf("ImportFile: %v", err)
	}
	if result.BookID == 0 || result.ChapterCount != 2 {
		t.Fatalf("result = %#v, want persisted book with 2 chapters", result)
	}

	books, err := store.ListBooks()
	if err != nil {
		t.Fatalf("ListBooks: %v", err)
	}
	if len(books) != 1 || books[0].Title != "book" {
		t.Fatalf("books = %#v", books)
	}
}

func TestImportServiceReturnsExistingBookWhenFileAlreadyImported(t *testing.T) {
	store, err := storage.OpenSQLite(t.TempDir() + "/reader.db")
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	defer store.Close()

	path := filepath.Join(t.TempDir(), "book.txt")
	if err := os.WriteFile(path, []byte("第一章 开始\n内容一\n\n第二章 继续\n内容二"), 0o644); err != nil {
		t.Fatalf("write txt: %v", err)
	}

	first, err := NewImportService(store).ImportFile(path)
	if err != nil {
		t.Fatalf("first ImportFile: %v", err)
	}
	second, err := NewImportService(store).ImportFile(path)
	if err != nil {
		t.Fatalf("second ImportFile: %v", err)
	}
	if second.BookID != first.BookID {
		t.Fatalf("second book id = %d, want existing id %d", second.BookID, first.BookID)
	}
	if !second.Existing {
		t.Fatalf("second result Existing = false, want true")
	}
	if second.ChapterCount != first.ChapterCount || second.WordCount != first.WordCount {
		t.Fatalf("second result = %#v, want same chapter and word counts as %#v", second, first)
	}

	books, err := store.ListBooks()
	if err != nil {
		t.Fatalf("ListBooks: %v", err)
	}
	if len(books) != 1 {
		t.Fatalf("book count = %d, want 1", len(books))
	}
}

func TestImportServiceCanReplaceExistingBook(t *testing.T) {
	store, err := storage.OpenSQLite(t.TempDir() + "/reader.db")
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	defer store.Close()

	path := filepath.Join(t.TempDir(), "book.txt")
	if err := os.WriteFile(path, []byte("第一章 开始\n内容一\n\n第二章 继续\n内容二"), 0o644); err != nil {
		t.Fatalf("write txt: %v", err)
	}

	first, err := NewImportService(store).ImportFile(path)
	if err != nil {
		t.Fatalf("first ImportFile: %v", err)
	}
	replaced, err := NewImportService(store).ImportFileWithOptions(path, ImportOptions{Replace: true})
	if err != nil {
		t.Fatalf("replace ImportFile: %v", err)
	}
	if replaced.BookID != first.BookID {
		t.Fatalf("replaced book id = %d, want same id %d", replaced.BookID, first.BookID)
	}
	if !replaced.Replaced {
		t.Fatalf("replaced result Replaced = false, want true")
	}

	books, err := store.ListBooks()
	if err != nil {
		t.Fatalf("ListBooks: %v", err)
	}
	if len(books) != 1 {
		t.Fatalf("book count = %d, want 1", len(books))
	}
}

func TestReadServiceContinuePreservesSavedLineOffset(t *testing.T) {
	store, err := storage.OpenSQLite(t.TempDir() + "/reader.db")
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	defer store.Close()
	bookID, err := store.CreateBook(domain.Book{SourceType: "local_txt", Title: "book", ContentHash: "read-progress"})
	if err != nil {
		t.Fatalf("CreateBook: %v", err)
	}
	if err := store.InsertChapters(bookID, []domain.Chapter{{ChapterNo: 1, Title: "第一章", Content: "a\nb\nc", WordCount: 3}}); err != nil {
		t.Fatalf("InsertChapters: %v", err)
	}
	if err := store.SaveProgress(domain.Progress{BookID: bookID, ChapterNo: 1, LineOffset: 2, CharOffset: 4}); err != nil {
		t.Fatalf("SaveProgress: %v", err)
	}
	if err := store.UpdateLastRead(bookID); err != nil {
		t.Fatalf("UpdateLastRead: %v", err)
	}

	view, err := NewReadService(store).Continue()
	if err != nil {
		t.Fatalf("Continue: %v", err)
	}
	if view.Progress.LineOffset != 2 || view.Progress.CharOffset != 4 {
		t.Fatalf("progress = %#v, want saved line and char offsets", view.Progress)
	}
}
