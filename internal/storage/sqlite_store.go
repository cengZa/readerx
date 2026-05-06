package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/heybox/readerx/internal/domain"
	_ "modernc.org/sqlite"
)

var ErrNotFound = errors.New("not found")

type SQLiteStore struct {
	db *sql.DB
}

func OpenSQLite(path string) (*SQLiteStore, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create data directory: %w", err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if _, err := db.Exec(`PRAGMA busy_timeout = 5000; PRAGMA journal_mode = WAL; PRAGMA foreign_keys = ON;`); err != nil {
		db.Close()
		return nil, fmt.Errorf("configure sqlite: %w", err)
	}
	store := &SQLiteStore{db: db}
	if err := store.InitSchema(); err != nil {
		db.Close()
		return nil, err
	}
	return store, nil
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

func (s *SQLiteStore) InitSchema() error {
	_, err := s.db.Exec(schemaSQL)
	if err != nil {
		return fmt.Errorf("init schema: %w", err)
	}
	return nil
}

func (s *SQLiteStore) CreateBook(book domain.Book) (int64, error) {
	now := time.Now().Unix()
	result, err := s.db.Exec(`
INSERT INTO books (source_type, source_book_id, title, author, description, cover_url, file_path, content_hash, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		defaultString(book.SourceType, "local"), book.SourceBookID, book.Title, book.Author, book.Description,
		book.CoverURL, book.FilePath, book.ContentHash, now, now)
	if err != nil {
		return 0, fmt.Errorf("create book: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("read book id: %w", err)
	}
	return id, nil
}

func (s *SQLiteStore) InsertChapters(bookID int64, chapters []domain.Chapter) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin insert chapters: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
INSERT INTO chapters (book_id, chapter_no, source_chapter_id, title, content, content_status, content_hash, word_count, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare insert chapters: %w", err)
	}
	defer stmt.Close()

	now := time.Now().Unix()
	for _, chapter := range chapters {
		_, err := stmt.Exec(bookID, chapter.ChapterNo, chapter.SourceChapterID, chapter.Title, chapter.Content,
			defaultString(chapter.ContentStatus, "cached"), chapter.ContentHash, chapter.WordCount, now, now)
		if err != nil {
			return fmt.Errorf("insert chapter %d: %w", chapter.ChapterNo, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit chapters: %w", err)
	}
	return nil
}

func (s *SQLiteStore) ListBooks() ([]domain.Book, error) {
	rows, err := s.db.Query(`
SELECT b.id, b.source_type, b.title, b.author, b.file_path, b.content_hash, b.created_at, b.updated_at, b.last_read_at,
       COUNT(c.id) AS chapter_count,
       COALESCE(p.percentage, 0) AS percentage
FROM books b
LEFT JOIN chapters c ON c.book_id = b.id
LEFT JOIN reading_progress p ON p.book_id = b.id
GROUP BY b.id
ORDER BY b.last_read_at DESC, b.created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list books: %w", err)
	}
	defer rows.Close()

	var books []domain.Book
	for rows.Next() {
		var book domain.Book
		if err := rows.Scan(&book.ID, &book.SourceType, &book.Title, &book.Author, &book.FilePath, &book.ContentHash,
			&book.CreatedAt, &book.UpdatedAt, &book.LastReadAt, &book.ChapterCount, &book.ProgressPercentage); err != nil {
			return nil, fmt.Errorf("scan book: %w", err)
		}
		books = append(books, book)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate books: %w", err)
	}
	return books, nil
}

func (s *SQLiteStore) GetBook(bookID int64) (domain.Book, error) {
	var book domain.Book
	err := s.db.QueryRow(`
SELECT id, source_type, source_book_id, title, author, description, cover_url, file_path, content_hash, created_at, updated_at, last_read_at
FROM books WHERE id = ?`, bookID).Scan(&book.ID, &book.SourceType, &book.SourceBookID, &book.Title, &book.Author,
		&book.Description, &book.CoverURL, &book.FilePath, &book.ContentHash, &book.CreatedAt, &book.UpdatedAt, &book.LastReadAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Book{}, ErrNotFound
	}
	if err != nil {
		return domain.Book{}, fmt.Errorf("get book: %w", err)
	}
	return book, nil
}

func (s *SQLiteStore) ListChapters(bookID int64) ([]domain.Chapter, error) {
	rows, err := s.db.Query(`
SELECT id, book_id, chapter_no, source_chapter_id, title, content_status, content_hash, word_count, created_at, updated_at
FROM chapters
WHERE book_id = ?
ORDER BY chapter_no ASC`, bookID)
	if err != nil {
		return nil, fmt.Errorf("list chapters: %w", err)
	}
	defer rows.Close()

	var chapters []domain.Chapter
	for rows.Next() {
		var chapter domain.Chapter
		if err := rows.Scan(&chapter.ID, &chapter.BookID, &chapter.ChapterNo, &chapter.SourceChapterID, &chapter.Title,
			&chapter.ContentStatus, &chapter.ContentHash, &chapter.WordCount, &chapter.CreatedAt, &chapter.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan chapter: %w", err)
		}
		chapters = append(chapters, chapter)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate chapters: %w", err)
	}
	return chapters, nil
}

func (s *SQLiteStore) GetChapter(bookID int64, chapterNo int) (domain.Chapter, error) {
	var chapter domain.Chapter
	err := s.db.QueryRow(`
SELECT id, book_id, chapter_no, source_chapter_id, title, content, content_status, content_hash, word_count, created_at, updated_at
FROM chapters WHERE book_id = ? AND chapter_no = ?`, bookID, chapterNo).Scan(&chapter.ID, &chapter.BookID, &chapter.ChapterNo,
		&chapter.SourceChapterID, &chapter.Title, &chapter.Content, &chapter.ContentStatus, &chapter.ContentHash,
		&chapter.WordCount, &chapter.CreatedAt, &chapter.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Chapter{}, ErrNotFound
	}
	if err != nil {
		return domain.Chapter{}, fmt.Errorf("get chapter: %w", err)
	}
	return chapter, nil
}

func (s *SQLiteStore) CountChapters(bookID int64) (int, error) {
	var count int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM chapters WHERE book_id = ?`, bookID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count chapters: %w", err)
	}
	return count, nil
}

func (s *SQLiteStore) SaveProgress(progress domain.Progress) error {
	now := time.Now().Unix()
	_, err := s.db.Exec(`
INSERT INTO reading_progress (book_id, chapter_no, line_offset, char_offset, percentage, updated_at)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(book_id) DO UPDATE SET
chapter_no = excluded.chapter_no,
line_offset = excluded.line_offset,
char_offset = excluded.char_offset,
percentage = excluded.percentage,
updated_at = excluded.updated_at`,
		progress.BookID, progress.ChapterNo, progress.LineOffset, progress.CharOffset, progress.Percentage, now)
	if err != nil {
		return fmt.Errorf("save progress: %w", err)
	}
	return nil
}

func (s *SQLiteStore) GetProgress(bookID int64) (domain.Progress, error) {
	var progress domain.Progress
	err := s.db.QueryRow(`
SELECT book_id, chapter_no, line_offset, char_offset, percentage, updated_at
FROM reading_progress WHERE book_id = ?`, bookID).Scan(&progress.BookID, &progress.ChapterNo, &progress.LineOffset,
		&progress.CharOffset, &progress.Percentage, &progress.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Progress{}, ErrNotFound
	}
	if err != nil {
		return domain.Progress{}, fmt.Errorf("get progress: %w", err)
	}
	return progress, nil
}

func (s *SQLiteStore) GetRecentProgress() (domain.Progress, error) {
	var progress domain.Progress
	err := s.db.QueryRow(`
SELECT p.book_id, p.chapter_no, p.line_offset, p.char_offset, p.percentage, p.updated_at
FROM reading_progress p
JOIN books b ON b.id = p.book_id
ORDER BY b.last_read_at DESC, p.updated_at DESC
LIMIT 1`).Scan(&progress.BookID, &progress.ChapterNo, &progress.LineOffset, &progress.CharOffset, &progress.Percentage, &progress.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Progress{}, ErrNotFound
	}
	if err != nil {
		return domain.Progress{}, fmt.Errorf("get recent progress: %w", err)
	}
	return progress, nil
}

func (s *SQLiteStore) UpdateLastRead(bookID int64) error {
	_, err := s.db.Exec(`UPDATE books SET last_read_at = ?, updated_at = ? WHERE id = ?`, time.Now().Unix(), time.Now().Unix(), bookID)
	if err != nil {
		return fmt.Errorf("update last read: %w", err)
	}
	return nil
}

func (s *SQLiteStore) AddBookmark(bookmark domain.Bookmark) (int64, error) {
	now := time.Now().Unix()
	result, err := s.db.Exec(`
INSERT INTO bookmarks (book_id, chapter_no, line_offset, char_offset, excerpt, note, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?)`,
		bookmark.BookID, bookmark.ChapterNo, bookmark.LineOffset, bookmark.CharOffset, bookmark.Excerpt, bookmark.Note, now)
	if err != nil {
		return 0, fmt.Errorf("add bookmark: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("read bookmark id: %w", err)
	}
	return id, nil
}

func (s *SQLiteStore) ListBookmarks(bookID int64) ([]domain.Bookmark, error) {
	query := `
SELECT bm.id, bm.book_id, b.title, bm.chapter_no, c.title, bm.line_offset, bm.char_offset, bm.excerpt, bm.note, bm.created_at
FROM bookmarks bm
JOIN books b ON b.id = bm.book_id
LEFT JOIN chapters c ON c.book_id = bm.book_id AND c.chapter_no = bm.chapter_no`
	args := []any{}
	if bookID > 0 {
		query += ` WHERE bm.book_id = ?`
		args = append(args, bookID)
	}
	query += ` ORDER BY bm.created_at DESC, bm.id DESC`
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list bookmarks: %w", err)
	}
	defer rows.Close()

	var bookmarks []domain.Bookmark
	for rows.Next() {
		var bookmark domain.Bookmark
		if err := rows.Scan(&bookmark.ID, &bookmark.BookID, &bookmark.BookTitle, &bookmark.ChapterNo, &bookmark.ChapterTitle,
			&bookmark.LineOffset, &bookmark.CharOffset, &bookmark.Excerpt, &bookmark.Note, &bookmark.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan bookmark: %w", err)
		}
		bookmarks = append(bookmarks, bookmark)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bookmarks: %w", err)
	}
	return bookmarks, nil
}

func (s *SQLiteStore) RemoveBookmark(bookmarkID int64) error {
	result, err := s.db.Exec(`DELETE FROM bookmarks WHERE id = ?`, bookmarkID)
	if err != nil {
		return fmt.Errorf("remove bookmark: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read removed bookmark count: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *SQLiteStore) SearchChapters(keyword string, bookID int64) ([]domain.SearchResult, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return nil, fmt.Errorf("keyword must not be empty")
	}
	query := `
SELECT b.id, b.title, c.chapter_no, c.title, c.content
FROM chapters c
JOIN books b ON b.id = c.book_id
WHERE (c.title LIKE ? OR c.content LIKE ?)`
	pattern := "%" + keyword + "%"
	args := []any{pattern, pattern}
	if bookID > 0 {
		query += ` AND b.id = ?`
		args = append(args, bookID)
	}
	query += ` ORDER BY b.last_read_at DESC, b.created_at DESC, c.chapter_no ASC LIMIT 50`
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("search chapters: %w", err)
	}
	defer rows.Close()

	var results []domain.SearchResult
	for rows.Next() {
		var content string
		var result domain.SearchResult
		if err := rows.Scan(&result.BookID, &result.BookTitle, &result.ChapterNo, &result.ChapterTitle, &content); err != nil {
			return nil, fmt.Errorf("scan search result: %w", err)
		}
		result.Snippet = makeSnippet(content, keyword, 24)
		results = append(results, result)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate search results: %w", err)
	}
	return results, nil
}

func makeSnippet(content, keyword string, radius int) string {
	runes := []rune(content)
	keyRunes := []rune(keyword)
	if len(runes) == 0 {
		return ""
	}
	index := strings.Index(content, keyword)
	if index < 0 {
		if len(runes) <= radius*2 {
			return string(runes)
		}
		return string(runes[:radius*2]) + "..."
	}
	prefixRunes := []rune(content[:index])
	start := len(prefixRunes) - radius
	if start < 0 {
		start = 0
	}
	end := len(prefixRunes) + len(keyRunes) + radius
	if end > len(runes) {
		end = len(runes)
	}
	snippet := string(runes[start:end])
	if start > 0 {
		snippet = "..." + snippet
	}
	if end < len(runes) {
		snippet += "..."
	}
	return snippet
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

const schemaSQL = `
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS books (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source_type TEXT NOT NULL DEFAULT 'local',
    source_book_id TEXT DEFAULT '',
    title TEXT NOT NULL,
    author TEXT DEFAULT '',
    description TEXT DEFAULT '',
    cover_url TEXT DEFAULT '',
    file_path TEXT DEFAULT '',
    content_hash TEXT DEFAULT '',
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    last_read_at INTEGER DEFAULT 0
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_books_content_hash_unique
ON books(content_hash)
WHERE content_hash <> '';

CREATE TABLE IF NOT EXISTS chapters (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    book_id INTEGER NOT NULL,
    chapter_no INTEGER NOT NULL,
    source_chapter_id TEXT DEFAULT '',
    title TEXT NOT NULL,
    content TEXT DEFAULT '',
    content_status TEXT NOT NULL DEFAULT 'cached',
    content_hash TEXT DEFAULT '',
    word_count INTEGER DEFAULT 0,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    UNIQUE(book_id, chapter_no),
    FOREIGN KEY(book_id) REFERENCES books(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS reading_progress (
    book_id INTEGER PRIMARY KEY,
    chapter_no INTEGER NOT NULL,
    line_offset INTEGER NOT NULL DEFAULT 0,
    char_offset INTEGER NOT NULL DEFAULT 0,
    percentage REAL DEFAULT 0,
    updated_at INTEGER NOT NULL,
    FOREIGN KEY(book_id) REFERENCES books(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS bookmarks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    book_id INTEGER NOT NULL,
    chapter_no INTEGER NOT NULL,
    line_offset INTEGER DEFAULT 0,
    char_offset INTEGER DEFAULT 0,
    excerpt TEXT DEFAULT '',
    note TEXT DEFAULT '',
    created_at INTEGER NOT NULL,
    FOREIGN KEY(book_id) REFERENCES books(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS notes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    book_id INTEGER NOT NULL,
    chapter_no INTEGER NOT NULL,
    line_offset INTEGER DEFAULT 0,
    char_offset INTEGER DEFAULT 0,
    content TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    FOREIGN KEY(book_id) REFERENCES books(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS sources (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    config_json TEXT DEFAULT '{}',
    enabled INTEGER NOT NULL DEFAULT 1,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_books_last_read_at ON books(last_read_at);
CREATE INDEX IF NOT EXISTS idx_chapters_book_no ON chapters(book_id, chapter_no);
CREATE INDEX IF NOT EXISTS idx_bookmarks_book ON bookmarks(book_id);
CREATE INDEX IF NOT EXISTS idx_notes_book ON notes(book_id);
`
