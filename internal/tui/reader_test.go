package tui

import (
	"regexp"
	"strings"
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

func TestReaderModelJumpsToTypedChapter(t *testing.T) {
	store := &fakeReaderStore{
		book: domain.Book{ID: 1, Title: "测试书"},
		chapters: map[int]domain.Chapter{
			1: {BookID: 1, ChapterNo: 1, Title: "第一章", Content: "内容一"},
			2: {BookID: 1, ChapterNo: 2, Title: "第二章", Content: "内容二"},
			3: {BookID: 1, ChapterNo: 3, Title: "第三章", Content: "内容三"},
		},
		chapterCount: 3,
	}
	model := NewReaderModel(store, app.ChapterView{Book: store.book, Chapter: store.chapters[1]}, 0)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	jump := updated.(ReaderModel)
	if !jump.jump.active {
		t.Fatalf("jump mode should be active")
	}

	updated, _ = jump.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	jump = updated.(ReaderModel)
	if jump.jump.input != "3" {
		t.Fatalf("jump input = %q, want 3", jump.jump.input)
	}

	updated, _ = jump.Update(tea.KeyMsg{Type: tea.KeyEnter})
	jumped := updated.(ReaderModel)
	if jumped.jump.active {
		t.Fatalf("jump mode should be inactive after enter")
	}
	if jumped.view.Chapter.ChapterNo != 3 {
		t.Fatalf("chapter after jump = %d, want 3", jumped.view.Chapter.ChapterNo)
	}
	if store.saved.ChapterNo != 3 {
		t.Fatalf("saved chapter after jump = %d, want 3", store.saved.ChapterNo)
	}
}

func TestReaderModelCancelsJumpMode(t *testing.T) {
	store := &fakeReaderStore{
		book: domain.Book{ID: 1, Title: "测试书"},
		chapters: map[int]domain.Chapter{
			1: {BookID: 1, ChapterNo: 1, Title: "第一章", Content: "内容一"},
		},
		chapterCount: 1,
	}
	model := NewReaderModel(store, app.ChapterView{Book: store.book, Chapter: store.chapters[1]}, 0)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	jump := updated.(ReaderModel)
	updated, _ = jump.Update(tea.KeyMsg{Type: tea.KeyEsc})
	cancelled := updated.(ReaderModel)
	if cancelled.jump.active {
		t.Fatalf("jump mode should be inactive after esc")
	}
}

func TestReaderModelAddsTypedNote(t *testing.T) {
	store := &fakeReaderStore{
		book: domain.Book{ID: 1, Title: "测试书"},
		chapters: map[int]domain.Chapter{
			1: {BookID: 1, ChapterNo: 1, Title: "第一章", Content: "内容一"},
		},
		chapterCount: 1,
	}
	model := NewReaderModel(store, app.ChapterView{Book: store.book, Chapter: store.chapters[1]}, 0)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	note := updated.(ReaderModel)
	if !note.note.active {
		t.Fatalf("note mode should be active")
	}

	updated, _ = note.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("hello")})
	note = updated.(ReaderModel)
	if note.note.input != "hello" {
		t.Fatalf("note input = %q, want hello", note.note.input)
	}

	updated, _ = note.Update(tea.KeyMsg{Type: tea.KeyEnter})
	added := updated.(ReaderModel)
	if added.note.active {
		t.Fatalf("note mode should be inactive after enter")
	}
	if store.note.Content != "hello" {
		t.Fatalf("stored note = %#v", store.note)
	}
}

func TestReaderModelSearchesAndJumpsToResult(t *testing.T) {
	store := &fakeReaderStore{
		book: domain.Book{ID: 1, Title: "测试书"},
		chapters: map[int]domain.Chapter{
			1: {BookID: 1, ChapterNo: 1, Title: "第一章", Content: "内容一"},
			2: {BookID: 1, ChapterNo: 2, Title: "第二章", Content: "一道剑气"},
		},
		chapterCount: 2,
		searchResults: []domain.SearchResult{
			{BookID: 1, BookTitle: "测试书", ChapterNo: 2, ChapterTitle: "第二章", Snippet: "一道剑气"},
		},
	}
	model := NewReaderModel(store, app.ChapterView{Book: store.book, Chapter: store.chapters[1]}, 0)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	search := updated.(ReaderModel)
	if !search.search.active || search.search.showResults {
		t.Fatalf("search input mode should be active: %#v", search.search)
	}

	updated, _ = search.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("剑气")})
	search = updated.(ReaderModel)
	updated, _ = search.Update(tea.KeyMsg{Type: tea.KeyEnter})
	results := updated.(ReaderModel)
	if !results.search.showResults || len(results.search.results) != 1 {
		t.Fatalf("search results mode = %#v", results.search)
	}

	updated, _ = results.Update(tea.KeyMsg{Type: tea.KeyEnter})
	jumped := updated.(ReaderModel)
	if jumped.view.Chapter.ChapterNo != 2 {
		t.Fatalf("chapter after search jump = %d, want 2", jumped.view.Chapter.ChapterNo)
	}
}

func TestSearchStatusShowsChapterTitleAndCleanSnippet(t *testing.T) {
	model := ReaderModel{
		search: searchState{
			input:       "剑气",
			showResults: true,
			results: []domain.SearchResult{
				{BookTitle: "测试书", ChapterNo: 2, ChapterTitle: "第二章 风起", Snippet: "一道\n剑气   从山巅而起"},
			},
		},
	}

	status := model.searchStatus()
	if !strings.Contains(status, "第二章 风起") {
		t.Fatalf("status = %q, want chapter title", status)
	}
	if strings.Contains(status, "\n") || strings.Contains(status, "   ") {
		t.Fatalf("status = %q, want clean one-line snippet", status)
	}
	if !strings.Contains(status, "一道 剑气 从山巅而起") {
		t.Fatalf("status = %q, want cleaned snippet", status)
	}
}

func TestReaderModelUsesConfiguredMaxWidth(t *testing.T) {
	store := &fakeReaderStore{
		book: domain.Book{ID: 1, Title: "测试书"},
		chapters: map[int]domain.Chapter{
			1: {BookID: 1, ChapterNo: 1, Title: "第一章", Content: "内容一"},
		},
		chapterCount: 1,
	}
	model := NewReaderModelWithOptions(store, app.ChapterView{Book: store.book, Chapter: store.chapters[1]}, 0, ReaderOptions{MaxWidth: 60})
	model.width = 120
	if got := model.contentWidth(); got != 60 {
		t.Fatalf("content width = %d, want 60", got)
	}
}

func TestReaderViewKeepsTitleBodyAndStatusInsideTerminalHeight(t *testing.T) {
	store := &fakeReaderStore{
		book: domain.Book{ID: 1, Title: "测试书"},
		chapters: map[int]domain.Chapter{
			1: {BookID: 1, ChapterNo: 1, Title: "第一章", Content: strings.Join([]string{
				"第一行",
				"第二行",
				"第三行",
				"第四行",
				"第五行",
				"第六行",
				"第七行",
			}, "\n")},
		},
		chapterCount: 1,
	}
	model := NewReaderModel(store, app.ChapterView{Book: store.book, Chapter: store.chapters[1]}, 0)
	model.width = 50
	model.height = 8
	model.repaginate()

	view := stripANSI(model.View())
	lines := strings.Split(view, "\n")
	if len(lines) != model.height {
		t.Fatalf("rendered line count = %d, want %d:\n%s", len(lines), model.height, view)
	}
	if !strings.Contains(lines[0], "《测试书》 第一章") {
		t.Fatalf("first line should contain title, got %q", lines[0])
	}
	if !strings.Contains(lines[len(lines)-1], "Pg 1/2") {
		t.Fatalf("last line should contain status, got %q", lines[len(lines)-1])
	}
}

func TestReaderStatusShowsChapterAndOverallProgress(t *testing.T) {
	store := &fakeReaderStore{
		book: domain.Book{ID: 1, Title: "测试书"},
		chapters: map[int]domain.Chapter{
			2: {BookID: 1, ChapterNo: 2, Title: "第二章", Content: strings.Join([]string{"一", "二", "三", "四", "五", "六"}, "\n")},
		},
		chapterCount: 4,
	}
	model := NewReaderModel(store, app.ChapterView{Book: store.book, Chapter: store.chapters[2]}, 0)
	model.width = 50
	model.height = 7
	model.repaginate()

	view := stripANSI(model.View())
	lines := strings.Split(view, "\n")
	status := lines[len(lines)-1]
	if !strings.Contains(status, "Ch 2/4") {
		t.Fatalf("status = %q, want chapter position", status)
	}
	if !strings.Contains(status, "25%") {
		t.Fatalf("status = %q, want overall progress percentage", status)
	}
}

func TestReaderStatusKeepsEssentialInfoOnNarrowWidth(t *testing.T) {
	got := readerStatusText(2, 4, 25, 3, 8, 30)
	for _, want := range []string{"Ch2/4", "25%", "Pg3/8", "?", "q"} {
		if !strings.Contains(got, want) {
			t.Fatalf("compact status = %q, want to contain %q", got, want)
		}
	}
	if strings.Contains(got, "scroll") || strings.Contains(got, "chapter") {
		t.Fatalf("compact status = %q, should omit long hints", got)
	}
}

func TestReaderModelUsesArrowPageKeysAndHomeEnd(t *testing.T) {
	store := &fakeReaderStore{
		book: domain.Book{ID: 1, Title: "测试书"},
		chapters: map[int]domain.Chapter{
			1: {BookID: 1, ChapterNo: 1, Title: "第一章", Content: strings.Join([]string{"一", "二", "三", "四", "五", "六", "七"}, "\n")},
		},
		chapterCount: 1,
	}
	model := NewReaderModel(store, app.ChapterView{Book: store.book, Chapter: store.chapters[1]}, 0)
	model.width = 50
	model.height = 5
	model.repaginate()

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRight})
	next := updated.(ReaderModel)
	if next.lineOffset == 0 {
		t.Fatalf("right arrow should move to next page")
	}

	updated, _ = next.Update(tea.KeyMsg{Type: tea.KeyLeft})
	prev := updated.(ReaderModel)
	if prev.lineOffset != 0 {
		t.Fatalf("left arrow line offset = %d, want 0", prev.lineOffset)
	}

	updated, _ = prev.Update(tea.KeyMsg{Type: tea.KeyEnd})
	end := updated.(ReaderModel)
	if end.lineOffset != end.paginator.LastPageOffset() {
		t.Fatalf("end line offset = %d, want %d", end.lineOffset, end.paginator.LastPageOffset())
	}

	updated, _ = end.Update(tea.KeyMsg{Type: tea.KeyHome})
	home := updated.(ReaderModel)
	if home.lineOffset != 0 {
		t.Fatalf("home line offset = %d, want 0", home.lineOffset)
	}
}

func TestReaderModelTogglesHelpView(t *testing.T) {
	store := &fakeReaderStore{
		book: domain.Book{ID: 1, Title: "测试书"},
		chapters: map[int]domain.Chapter{
			1: {BookID: 1, ChapterNo: 1, Title: "第一章", Content: "正文"},
		},
		chapterCount: 1,
	}
	model := NewReaderModel(store, app.ChapterView{Book: store.book, Chapter: store.chapters[1]}, 0)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	help := updated.(ReaderModel)
	if !help.help {
		t.Fatalf("help should be active")
	}
	view := stripANSI(help.View())
	if !strings.Contains(view, "ReaderX help") || !strings.Contains(view, "/ search") {
		t.Fatalf("help view missing expected shortcuts:\n%s", view)
	}

	updated, _ = help.Update(tea.KeyMsg{Type: tea.KeyEsc})
	closed := updated.(ReaderModel)
	if closed.help {
		t.Fatalf("help should close on esc")
	}
}

func TestOverallPercentageIncludesChapterAndPage(t *testing.T) {
	got := overallPercentage(2, 4, 2, 2)
	if got != 37.5 {
		t.Fatalf("percentage = %v, want 37.5", got)
	}
}

type fakeReaderStore struct {
	book          domain.Book
	chapters      map[int]domain.Chapter
	chapterCount  int
	saved         domain.Progress
	note          domain.Note
	searchResults []domain.SearchResult
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

func (f *fakeReaderStore) AddNote(note domain.Note) (int64, error) {
	f.note = note
	return 1, nil
}

func (f *fakeReaderStore) GetChapter(bookID int64, chapterNo int) (domain.Chapter, error) {
	return f.chapters[chapterNo], nil
}

func (f *fakeReaderStore) CountChapters(bookID int64) (int, error) {
	return f.chapterCount, nil
}

func (f *fakeReaderStore) SearchChapters(keyword string, bookID int64, limit int) ([]domain.SearchResult, error) {
	return f.searchResults, nil
}

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)

func stripANSI(value string) string {
	return ansiPattern.ReplaceAllString(value, "")
}
