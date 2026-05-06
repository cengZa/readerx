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
