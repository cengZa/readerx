package app

import (
	"errors"
	"fmt"

	"github.com/heybox/readerx/internal/domain"
	"github.com/heybox/readerx/internal/storage"
)

type ReadService struct {
	store storage.Store
}

type ChapterView struct {
	Book     domain.Book
	Chapter  domain.Chapter
	Progress domain.Progress
}

func NewReadService(store storage.Store) *ReadService {
	return &ReadService{store: store}
}

func (s *ReadService) ReadChapter(bookID int64, chapterNo int) (ChapterView, error) {
	book, err := s.store.GetBook(bookID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return ChapterView{}, fmt.Errorf("未找到 Book ID=%d 的书籍，请先使用 readerx list 查看可用书籍", bookID)
		}
		return ChapterView{}, err
	}
	progress := domain.Progress{BookID: bookID, ChapterNo: 1}
	savedProgress, savedProgressErr := s.store.GetProgress(bookID)
	if chapterNo <= 0 {
		if savedProgressErr == nil {
			progress = savedProgress
			chapterNo = progress.ChapterNo
		} else if errors.Is(savedProgressErr, storage.ErrNotFound) {
			chapterNo = 1
		} else {
			return ChapterView{}, savedProgressErr
		}
	} else if savedProgressErr == nil && savedProgress.ChapterNo == chapterNo {
		progress = savedProgress
	} else if savedProgressErr != nil && !errors.Is(savedProgressErr, storage.ErrNotFound) {
		return ChapterView{}, savedProgressErr
	}

	chapter, err := s.store.GetChapter(bookID, chapterNo)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return ChapterView{}, fmt.Errorf("未找到 Book ID=%d 的第 %d 章", bookID, chapterNo)
		}
		return ChapterView{}, err
	}
	count, err := s.store.CountChapters(bookID)
	if err != nil {
		return ChapterView{}, err
	}
	percentage := 0.0
	if count > 0 {
		percentage = float64(chapterNo-1) / float64(count) * 100
	}
	progress.BookID = bookID
	progress.ChapterNo = chapterNo
	progress.Percentage = percentage
	if err := s.store.SaveProgress(progress); err != nil {
		return ChapterView{}, err
	}
	if err := s.store.UpdateLastRead(bookID); err != nil {
		return ChapterView{}, err
	}
	return ChapterView{Book: book, Chapter: chapter, Progress: progress}, nil
}

func (s *ReadService) Continue() (ChapterView, error) {
	progress, err := s.store.GetRecentProgress()
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return ChapterView{}, fmt.Errorf("暂无阅读进度，请先使用 readerx list 查看书籍或 readerx read <book-id> 开始阅读")
		}
		return ChapterView{}, err
	}
	return s.ReadChapter(progress.BookID, progress.ChapterNo)
}
