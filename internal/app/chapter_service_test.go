package app

import (
	"testing"

	"github.com/heybox/readerx/internal/domain"
	"github.com/heybox/readerx/internal/storage"
)

func TestChapterServiceListsChapterMetadata(t *testing.T) {
	store, err := storage.OpenSQLite(t.TempDir() + "/reader.db")
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	defer store.Close()
	bookID, err := store.CreateBook(domain.Book{SourceType: "local_txt", Title: "book", ContentHash: "chapters"})
	if err != nil {
		t.Fatalf("CreateBook: %v", err)
	}
	if err := store.InsertChapters(bookID, []domain.Chapter{
		{ChapterNo: 1, Title: "第一章", Content: "内容一", WordCount: 3},
		{ChapterNo: 2, Title: "第二章", Content: "内容二", WordCount: 3},
	}); err != nil {
		t.Fatalf("InsertChapters: %v", err)
	}

	chapters, err := NewChapterService(store).List(bookID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(chapters) != 2 || chapters[0].Title != "第一章" || chapters[1].ChapterNo != 2 {
		t.Fatalf("chapters = %#v", chapters)
	}
}
