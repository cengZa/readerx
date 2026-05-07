package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/heybox/readerx/internal/app"
	"github.com/heybox/readerx/internal/domain"
	"github.com/heybox/readerx/internal/reader"
	"github.com/mattn/go-runewidth"
)

type ProgressSaver interface {
	SaveProgress(progress domain.Progress) error
	UpdateLastRead(bookID int64) error
	AddBookmark(bookmark domain.Bookmark) (int64, error)
	AddNote(note domain.Note) (int64, error)
	GetChapter(bookID int64, chapterNo int) (domain.Chapter, error)
	CountChapters(bookID int64) (int, error)
	SearchChapters(keyword string, bookID int64, limit int) ([]domain.SearchResult, error)
}

type ReaderModel struct {
	store      ProgressSaver
	view       app.ChapterView
	paginator  reader.Paginator
	lineOffset int
	width      int
	height     int
	err        error
	jump       jumpState
	note       noteState
	search     searchState
	options    ReaderOptions
	styles     readerStyles
}

type ReaderOptions struct {
	MaxWidth int
	Theme    string
}

type readerStyles struct {
	title  lipgloss.Style
	body   lipgloss.Style
	status lipgloss.Style
}

type jumpState struct {
	active bool
	input  string
}

type noteState struct {
	active bool
	input  string
}

type searchState struct {
	active      bool
	input       string
	showResults bool
	results     []domain.SearchResult
	selected    int
}

func NewReaderModel(store ProgressSaver, view app.ChapterView, startLineOffset int) ReaderModel {
	return NewReaderModelWithOptions(store, view, startLineOffset, ReaderOptions{})
}

func NewReaderModelWithOptions(store ProgressSaver, view app.ChapterView, startLineOffset int, options ReaderOptions) ReaderModel {
	options = normalizeOptions(options)
	model := ReaderModel{
		store:      store,
		view:       view,
		lineOffset: startLineOffset,
		width:      80,
		height:     24,
		options:    options,
		styles:     stylesForTheme(options.Theme),
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
		if m.jump.active {
			return m.updateJump(msg), nil
		}
		if m.note.active {
			return m.updateNote(msg), nil
		}
		if m.search.active {
			return m.updateSearch(msg), nil
		}
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
		case key.Matches(msg, keys.jumpChapter):
			m.jump = jumpState{active: true}
			m.err = nil
		case key.Matches(msg, keys.addNote):
			m.note = noteState{active: true}
			m.err = nil
		case key.Matches(msg, keys.search):
			m.search = searchState{active: true}
			m.err = nil
		}
	}
	return m, nil
}

func (m ReaderModel) updateJump(msg tea.KeyMsg) ReaderModel {
	switch msg.Type {
	case tea.KeyEsc, tea.KeyCtrlC:
		m.jump = jumpState{}
	case tea.KeyEnter:
		m.commitJump()
	case tea.KeyBackspace:
		if len(m.jump.input) > 0 {
			m.jump.input = m.jump.input[:len(m.jump.input)-1]
		}
	case tea.KeyRunes:
		for _, r := range msg.Runes {
			if r >= '0' && r <= '9' {
				m.jump.input += string(r)
			}
		}
	}
	return m
}

func (m ReaderModel) updateSearch(msg tea.KeyMsg) ReaderModel {
	if m.search.showResults {
		return m.updateSearchResults(msg)
	}
	switch msg.Type {
	case tea.KeyEsc, tea.KeyCtrlC:
		m.search = searchState{}
	case tea.KeyEnter:
		m.commitSearch()
	case tea.KeyBackspace:
		if len(m.search.input) > 0 {
			m.search.input = m.search.input[:len(m.search.input)-1]
		}
	case tea.KeySpace:
		m.search.input += " "
	case tea.KeyRunes:
		m.search.input += string(msg.Runes)
	}
	return m
}

func (m ReaderModel) updateSearchResults(msg tea.KeyMsg) ReaderModel {
	switch msg.Type {
	case tea.KeyEsc, tea.KeyCtrlC:
		m.search = searchState{}
	case tea.KeyEnter:
		m.jumpToSearchResult()
	case tea.KeyRunes:
		if len(msg.Runes) == 1 {
			switch msg.Runes[0] {
			case 'j':
				if m.search.selected < len(m.search.results)-1 {
					m.search.selected++
				}
			case 'k':
				if m.search.selected > 0 {
					m.search.selected--
				}
			}
		}
	case tea.KeyDown:
		if m.search.selected < len(m.search.results)-1 {
			m.search.selected++
		}
	case tea.KeyUp:
		if m.search.selected > 0 {
			m.search.selected--
		}
	}
	return m
}

func (m ReaderModel) updateNote(msg tea.KeyMsg) ReaderModel {
	switch msg.Type {
	case tea.KeyEsc, tea.KeyCtrlC:
		m.note = noteState{}
	case tea.KeyEnter:
		m.commitNote()
	case tea.KeyBackspace:
		if len(m.note.input) > 0 {
			m.note.input = m.note.input[:len(m.note.input)-1]
		}
	case tea.KeySpace:
		m.note.input += " "
	case tea.KeyRunes:
		m.note.input += string(msg.Runes)
	}
	return m
}

func (m ReaderModel) View() string {
	frameWidth := m.frameWidth()
	title := m.styles.title.Render(fitDisplayWidth(fmt.Sprintf("《%s》 %s", m.view.Book.Title, m.view.Chapter.Title), frameWidth))
	lines := m.paginator.VisibleLines(m.lineOffset)
	body := m.renderBody(lines)
	page, total := m.paginator.PageInfo(m.lineOffset)
	statusText := fmt.Sprintf("Page %d/%d | j/k scroll | Space/u page | n/p chapter | g jump | b bookmark | s save | q quit", page, total)
	if m.jump.active {
		statusText = fmt.Sprintf("Go to chapter: %s", m.jump.input)
	}
	if m.note.active {
		statusText = fmt.Sprintf("Add note: %s", m.note.input)
	}
	if m.search.active {
		statusText = fmt.Sprintf("Search: %s", m.search.input)
		if m.search.showResults {
			statusText = m.searchStatus()
		}
	}
	if m.err != nil {
		statusText = m.err.Error() + " | " + statusText
	}
	status := m.styles.status.Render(fitDisplayWidth(statusText, frameWidth))
	return lipgloss.JoinVertical(lipgloss.Left, title, body, status)
}

func (m ReaderModel) renderBody(lines []string) string {
	bodyLines := make([]string, 0, m.contentHeight())
	for i := 0; i < m.contentHeight(); i++ {
		line := ""
		if i < len(lines) {
			line = lines[i]
		}
		bodyLines = append(bodyLines, m.styles.body.Render(m.renderBodyLine(line)))
	}
	return strings.Join(bodyLines, "\n")
}

func (m ReaderModel) renderBodyLine(line string) string {
	content := fitDisplayWidth(line, m.contentWidth())
	return "  " + content + "  "
}

func (m ReaderModel) searchStatus() string {
	if len(m.search.results) == 0 {
		return fmt.Sprintf("No results for %q | Esc cancel", m.search.input)
	}
	result := m.search.results[m.search.selected]
	return fmt.Sprintf("Result %d/%d | Enter jump | j/k select | %s ch.%d %s",
		m.search.selected+1, len(m.search.results), result.BookTitle, result.ChapterNo, result.Snippet)
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
	if width > m.options.MaxWidth {
		return m.options.MaxWidth
	}
	return width
}

func (m ReaderModel) frameWidth() int {
	return m.contentWidth() + 4
}

func (m ReaderModel) contentHeight() int {
	height := m.height - 2
	if height < 3 {
		return 3
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

func (m *ReaderModel) commitJump() {
	input := strings.TrimSpace(m.jump.input)
	m.jump = jumpState{}
	if input == "" {
		m.err = fmt.Errorf("chapter number required")
		return
	}
	chapterNo, err := strconv.Atoi(input)
	if err != nil || chapterNo <= 0 {
		m.err = fmt.Errorf("invalid chapter number: %s", input)
		return
	}
	m.jumpToChapter(chapterNo)
}

func (m *ReaderModel) jumpToChapter(chapterNo int) {
	chapterCount, err := m.store.CountChapters(m.view.Book.ID)
	if err != nil {
		m.err = err
		return
	}
	if chapterNo < 1 || chapterNo > chapterCount {
		m.err = fmt.Errorf("chapter %d is out of range 1-%d", chapterNo, chapterCount)
		return
	}
	chapter, err := m.store.GetChapter(m.view.Book.ID, chapterNo)
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

func (m *ReaderModel) commitNote() {
	input := strings.TrimSpace(m.note.input)
	m.note = noteState{}
	if input == "" {
		m.err = fmt.Errorf("note content required")
		return
	}
	m.save()
	if m.err != nil {
		return
	}
	charOffset := m.paginator.CharOffsetForLine(m.lineOffset)
	id, err := m.store.AddNote(domain.Note{
		BookID:     m.view.Book.ID,
		ChapterNo:  m.view.Chapter.ChapterNo,
		LineOffset: m.lineOffset,
		CharOffset: charOffset,
		Content:    input,
	})
	if err != nil {
		m.err = err
		return
	}
	m.err = fmt.Errorf("note saved: %d", id)
}

func (m *ReaderModel) commitSearch() {
	input := strings.TrimSpace(m.search.input)
	if input == "" {
		m.err = fmt.Errorf("search keyword required")
		m.search = searchState{}
		return
	}
	results, err := m.store.SearchChapters(input, m.view.Book.ID, 20)
	if err != nil {
		m.err = err
		m.search = searchState{}
		return
	}
	m.search.showResults = true
	m.search.results = results
	m.search.selected = 0
	m.err = nil
}

func (m *ReaderModel) jumpToSearchResult() {
	if len(m.search.results) == 0 {
		return
	}
	result := m.search.results[m.search.selected]
	m.search = searchState{}
	m.jumpToChapter(result.ChapterNo)
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

func normalizeOptions(options ReaderOptions) ReaderOptions {
	if options.MaxWidth <= 0 {
		options.MaxWidth = 100
	}
	if options.MaxWidth < 20 {
		options.MaxWidth = 20
	}
	switch options.Theme {
	case "dark", "light":
	default:
		options.Theme = "default"
	}
	return options
}

func stylesForTheme(theme string) readerStyles {
	styles := readerStyles{
		title:  lipgloss.NewStyle().Bold(true),
		body:   lipgloss.NewStyle(),
		status: lipgloss.NewStyle().Reverse(true),
	}
	switch theme {
	case "dark":
		styles.title = styles.title.Foreground(lipgloss.Color("229"))
		styles.body = styles.body.Foreground(lipgloss.Color("252"))
		styles.status = styles.status.Background(lipgloss.Color("238")).Foreground(lipgloss.Color("229"))
	case "light":
		styles.title = styles.title.Foreground(lipgloss.Color("18"))
		styles.body = styles.body.Foreground(lipgloss.Color("16"))
		styles.status = styles.status.Background(lipgloss.Color("254")).Foreground(lipgloss.Color("18"))
	}
	return styles
}

func fitDisplayWidth(value string, width int) string {
	if width < 1 {
		return ""
	}
	if runewidth.StringWidth(value) > width {
		value = runewidth.Truncate(value, width, "…")
	}
	return value + strings.Repeat(" ", width-runewidth.StringWidth(value))
}

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
	jumpChapter key.Binding
	addNote     key.Binding
	search      key.Binding
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
	jumpChapter: key.NewBinding(key.WithKeys("g")),
	addNote:     key.NewBinding(key.WithKeys("m")),
	search:      key.NewBinding(key.WithKeys("/")),
}

func RunReader(store ProgressSaver, view app.ChapterView, startLineOffset int, options ...ReaderOptions) error {
	readerOptions := ReaderOptions{}
	if len(options) > 0 {
		readerOptions = options[0]
	}
	program := tea.NewProgram(NewReaderModelWithOptions(store, view, startLineOffset, readerOptions), tea.WithAltScreen())
	_, err := program.Run()
	return err
}
