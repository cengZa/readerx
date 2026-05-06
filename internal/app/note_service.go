package app

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/heybox/readerx/internal/domain"
	"github.com/heybox/readerx/internal/storage"
)

type NoteService struct {
	store storage.Store
}

func NewNoteService(store storage.Store) *NoteService {
	return &NoteService{store: store}
}

func (s *NoteService) AddCurrent(bookID int64, content string) (domain.Note, error) {
	progress, err := s.store.GetProgress(bookID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			progress = domain.Progress{BookID: bookID, ChapterNo: 1}
		} else {
			return domain.Note{}, err
		}
	}
	return s.AddAt(domain.Note{
		BookID:     bookID,
		ChapterNo:  progress.ChapterNo,
		LineOffset: progress.LineOffset,
		CharOffset: progress.CharOffset,
		Content:    content,
	})
}

func (s *NoteService) AddAt(note domain.Note) (domain.Note, error) {
	if note.BookID <= 0 {
		return domain.Note{}, fmt.Errorf("book id must be positive")
	}
	if note.ChapterNo <= 0 {
		note.ChapterNo = 1
	}
	if strings.TrimSpace(note.Content) == "" {
		return domain.Note{}, fmt.Errorf("note content must not be empty")
	}
	id, err := s.store.AddNote(note)
	if err != nil {
		return domain.Note{}, err
	}
	note.ID = id
	return note, nil
}

func (s *NoteService) List(bookID int64) ([]domain.Note, error) {
	return s.store.ListNotes(bookID)
}

func (s *NoteService) Remove(noteID int64) error {
	return s.store.RemoveNote(noteID)
}

func (s *NoteService) ExportMarkdown(bookID int64) (string, error) {
	notes, err := s.store.ListNotes(bookID)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	buf.WriteString("# Notes\n\n")
	if len(notes) == 0 {
		buf.WriteString("_No notes._\n")
		return buf.String(), nil
	}
	for _, note := range notes {
		fmt.Fprintf(&buf, "## %s / Chapter %d %s\n\n", note.BookTitle, note.ChapterNo, note.ChapterTitle)
		if note.UpdatedAt > 0 {
			fmt.Fprintf(&buf, "- Updated: %s\n", time.Unix(note.UpdatedAt, 0).Format("2006-01-02 15:04"))
		}
		fmt.Fprintf(&buf, "- Location: line %d, char %d\n\n", note.LineOffset, note.CharOffset)
		fmt.Fprintf(&buf, "%s\n\n", strings.TrimSpace(note.Content))
	}
	return buf.String(), nil
}
