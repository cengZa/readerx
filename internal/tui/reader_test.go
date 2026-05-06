package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/heybox/readerx/internal/app"
	"github.com/heybox/readerx/internal/domain"
)

func TestReaderModelMovesToNextAndPreviousChapter(t *testing.T) {
	store := &fakeReaderStore{
		book: domain.Book{ID: 1, Title: "测试书"},
		chapters: map[int]domain.Chapter{
			1: {BookID: 1, ChapterNo: 1, Title: "第一章", Content: "内容一"},
			2: {BookID: 1, ChapterNo: 2, Title: "第二章", Content: "内容二"},
		},
		chapterCount: 2,
	}
	model := NewReaderModel(store, app.ChapterView{Book: store.book, Chapter: store.chapters[1]}, 0)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	next := updated.(ReaderModel)
	if next.view.Chapter.ChapterNo != 2 {
		t.Fatalf("chapter after n = %d, want 2", next.view.Chapter.ChapterNo)
	}
	if store.saved.ChapterNo != 2 {
		t.Fatalf("saved chapter after n = %d, want 2", store.saved.ChapterNo)
	}

	updated, _ = next.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	prev := updated.(ReaderModel)
	if prev.view.Chapter.ChapterNo != 1 {
		t.Fatalf("chapter after p = %d, want 1", prev.view.Chapter.ChapterNo)
	}
}

func TestOverallPercentageIncludesChapterAndPage(t *testing.T) {
	got := overallPercentage(2, 4, 2, 2)
	if got != 37.5 {
		t.Fatalf("percentage = %v, want 37.5", got)
	}
}

type fakeReaderStore struct {
	book         domain.Book
	chapters     map[int]domain.Chapter
	chapterCount int
	saved        domain.Progress
}

func (f *fakeReaderStore) SaveProgress(progress domain.Progress) error {
	f.saved = progress
	return nil
}

func (f *fakeReaderStore) UpdateLastRead(bookID int64) error {
	return nil
}

func (f *fakeReaderStore) AddBookmark(bookmark domain.Bookmark) (int64, error) {
	return 1, nil
}

func (f *fakeReaderStore) GetChapter(bookID int64, chapterNo int) (domain.Chapter, error) {
	return f.chapters[chapterNo], nil
}

func (f *fakeReaderStore) CountChapters(bookID int64) (int, error) {
	return f.chapterCount, nil
}
