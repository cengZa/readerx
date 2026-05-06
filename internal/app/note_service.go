package app

import (
	"errors"
	"fmt"
	"strings"

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
