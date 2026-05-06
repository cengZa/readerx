package app

import (
	"testing"

	"github.com/heybox/readerx/internal/domain"
	"github.com/heybox/readerx/internal/storage"
)

func TestSearchServiceFindsKeyword(t *testing.T) {
	store, err := storage.OpenSQLite(t.TempDir() + "/reader.db")
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	defer store.Close()
	bookID, err := store.CreateBook(domain.Book{SourceType: "local_txt", Title: "测试书", ContentHash: "hash-search"})
	if err != nil {
		t.Fatalf("CreateBook: %v", err)
	}
	if err := store.InsertChapters(bookID, []domain.Chapter{{ChapterNo: 1, Title: "第一章", Content: "一道剑气自山巅而起", WordCount: 10}}); err != nil {
		t.Fatalf("InsertChapters: %v", err)
	}

	results, err := NewSearchService(store).Search("剑气", 0, 50)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 || results[0].BookTitle != "测试书" || results[0].Snippet == "" {
		t.Fatalf("results = %#v", results)
	}
}
