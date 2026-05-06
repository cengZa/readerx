package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/heybox/readerx/internal/app"
	"github.com/heybox/readerx/internal/domain"
	"github.com/heybox/readerx/internal/reader"
)

type ProgressSaver interface {
	SaveProgress(progress domain.Progress) error
	UpdateLastRead(bookID int64) error
	AddBookmark(bookmark domain.Bookmark) (int64, error)
	GetChapter(bookID int64, chapterNo int) (domain.Chapter, error)
	CountChapters(bookID int64) (int, error)
}

type ReaderModel struct {
	store      ProgressSaver
	view       app.ChapterView
	paginator  reader.Paginator
	lineOffset int
	width      int
	height     int
	err        error
}

func NewReaderModel(store ProgressSaver, view app.ChapterView, startLineOffset int) ReaderModel {
	model := ReaderModel{
		store:      store,
		view:       view,
		lineOffset: startLineOffset,
		width:      80,
		height:     24,
	}
	model.repaginate()
	return model
}

func (m ReaderModel) Init() tea.Cmd {
	return nil
}

func (m ReaderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.repaginate()
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.quit):
			m.save()
			return m, tea.Quit
		case key.Matches(msg, keys.down):
			m.lineOffset = m.paginator.NextLine(m.lineOffset)
		case key.Matches(msg, keys.up):
			m.lineOffset = m.paginator.PrevLine(m.lineOffset)
		case key.Matches(msg, keys.nextPage):
			m.lineOffset = m.paginator.NextPage(m.lineOffset)
		case key.Matches(msg, keys.prevPage):
			m.lineOffset = m.paginator.PrevPage(m.lineOffset)
		case key.Matches(msg, keys.save):
			m.save()
		case key.Matches(msg, keys.bookmark):
			m.addBookmark()
		case key.Matches(msg, keys.nextChapter):
			m.moveChapter(1)
		case key.Matches(msg, keys.prevChapter):
			m.moveChapter(-1)
		}
	}
	return m, nil
}

func (m ReaderModel) View() string {
	title := titleStyle.Render(fmt.Sprintf("《%s》 %s", m.view.Book.Title, m.view.Chapter.Title))
	lines := m.paginator.VisibleLines(m.lineOffset)
	body := bodyStyle.Width(m.contentWidth()).Render(strings.Join(lines, "\n"))
	page, total := m.paginator.PageInfo(m.lineOffset)
	statusText := fmt.Sprintf("Page %d/%d | j/k scroll | Space/u page | n/p chapter | b bookmark | s save | q quit", page, total)
	if m.err != nil {
		statusText = m.err.Error() + " | " + statusText
	}
	status := statusStyle.Width(m.contentWidth()).Render(statusText)
	return lipgloss.JoinVertical(lipgloss.Left, title, body, status)
}

func (m *ReaderModel) repaginate() {
	m.paginator = reader.NewPaginator(m.view.Chapter.Content, m.contentWidth(), m.contentHeight())
	m.lineOffset = m.paginator.ClampOffset(m.lineOffset)
}

func (m ReaderModel) contentWidth() int {
	width := m.width - 4
	if width < 20 {
		return 20
	}
	if width > 100 {
		return 100
	}
	return width
}

func (m ReaderModel) contentHeight() int {
	height := m.height - 4
	if height < 5 {
		return 5
	}
	return height
}

func (m *ReaderModel) save() {
	currentPage, totalPages := m.paginator.PageInfo(m.lineOffset)
	chapterCount, err := m.store.CountChapters(m.view.Book.ID)
	if err != nil {
		m.err = err
		return
	}
	err = m.store.SaveProgress(domain.Progress{
		BookID:     m.view.Book.ID,
		ChapterNo:  m.view.Chapter.ChapterNo,
		LineOffset: m.lineOffset,
		CharOffset: m.paginator.CharOffsetForLine(m.lineOffset),
		Percentage: overallPercentage(m.view.Chapter.ChapterNo, chapterCount, currentPage, totalPages),
	})
	if err != nil {
		m.err = err
		return
	}
	if err := m.store.UpdateLastRead(m.view.Book.ID); err != nil {
		m.err = err
		return
	}
	m.err = nil
}

func (m *ReaderModel) moveChapter(delta int) {
	nextChapterNo := m.view.Chapter.ChapterNo + delta
	if nextChapterNo < 1 {
		m.err = fmt.Errorf("already at first chapter")
		return
	}
	chapterCount, err := m.store.CountChapters(m.view.Book.ID)
	if err != nil {
		m.err = err
		return
	}
	if nextChapterNo > chapterCount {
		m.err = fmt.Errorf("already at last chapter")
		return
	}
	chapter, err := m.store.GetChapter(m.view.Book.ID, nextChapterNo)
	if err != nil {
		m.err = err
		return
	}
	m.view.Chapter = chapter
	m.lineOffset = 0
	m.repaginate()
	m.save()
}

func overallPercentage(chapterNo, chapterCount, currentPage, totalPages int) float64 {
	if chapterCount <= 0 {
		return 0
	}
	if chapterNo < 1 {
		chapterNo = 1
	}
	if chapterNo > chapterCount {
		chapterNo = chapterCount
	}
	pageFraction := 0.0
	if totalPages > 0 && currentPage > 0 {
		pageFraction = float64(currentPage-1) / float64(totalPages)
	}
	return (float64(chapterNo-1) + pageFraction) / float64(chapterCount) * 100
}

func (m *ReaderModel) addBookmark() {
	m.save()
	if m.err != nil {
		return
	}
	charOffset := m.paginator.CharOffsetForLine(m.lineOffset)
	id, err := m.store.AddBookmark(domain.Bookmark{
		BookID:     m.view.Book.ID,
		ChapterNo:  m.view.Chapter.ChapterNo,
		LineOffset: m.lineOffset,
		CharOffset: charOffset,
		Excerpt:    excerptFrom(m.view.Chapter.Content, charOffset, 36),
	})
	if err != nil {
		m.err = err
		return
	}
	m.err = fmt.Errorf("bookmark saved: %d", id)
}

func excerptFrom(content string, charOffset, length int) string {
	runes := []rune(content)
	if len(runes) == 0 {
		return ""
	}
	if charOffset < 0 {
		charOffset = 0
	}
	if charOffset > len(runes) {
		charOffset = len(runes)
	}
	end := charOffset + length
	if end > len(runes) {
		end = len(runes)
	}
	return strings.TrimSpace(string(runes[charOffset:end]))
}

var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Padding(0, 1)
	bodyStyle   = lipgloss.NewStyle().Padding(1, 2)
	statusStyle = lipgloss.NewStyle().Reverse(true).Padding(0, 1)
)

var keys = struct {
	quit        key.Binding
	down        key.Binding
	up          key.Binding
	nextPage    key.Binding
	prevPage    key.Binding
	save        key.Binding
	bookmark    key.Binding
	nextChapter key.Binding
	prevChapter key.Binding
}{
	quit:        key.NewBinding(key.WithKeys("q", "ctrl+c")),
	down:        key.NewBinding(key.WithKeys("j", "down")),
	up:          key.NewBinding(key.WithKeys("k", "up")),
	nextPage:    key.NewBinding(key.WithKeys(" ", "pgdown")),
	prevPage:    key.NewBinding(key.WithKeys("u", "pgup")),
	save:        key.NewBinding(key.WithKeys("s")),
	bookmark:    key.NewBinding(key.WithKeys("b")),
	nextChapter: key.NewBinding(key.WithKeys("n")),
	prevChapter: key.NewBinding(key.WithKeys("p")),
}

func RunReader(store ProgressSaver, view app.ChapterView, startLineOffset int) error {
	program := tea.NewProgram(NewReaderModel(store, view, startLineOffset), tea.WithAltScreen())
	_, err := program.Run()
	return err
}
