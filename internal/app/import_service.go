package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/heybox/readerx/internal/domain"
	"github.com/heybox/readerx/internal/source"
	"github.com/heybox/readerx/internal/storage"
)

const (
	DefaultMaxDownloadBytes int64 = 100 * 1024 * 1024
	defaultHTTPTimeout            = 30 * time.Second
)

var ErrDownloadTooLarge = errors.New("download too large")

var importURLHTTPClient = &http.Client{Timeout: defaultHTTPTimeout}

type ImportService struct {
	store storage.Store
}

type ImportResult struct {
	BookID       int64
	Title        string
	ChapterCount int
	WordCount    int
	Existing     bool
	Replaced     bool
	Warnings     []string
}

type ImportOptions struct {
	Replace          bool
	TitleOverride    string
	MaxDownloadBytes int64
}

func NewImportService(store storage.Store) *ImportService {
	return &ImportService{store: store}
}

func (s *ImportService) ImportFile(path string) (ImportResult, error) {
	return s.ImportFileWithOptions(path, ImportOptions{})
}

func (s *ImportService) ImportFileWithOptions(path string, options ImportOptions) (ImportResult, error) {
	book, chapters, err := source.ImportLocalFile(path)
	if err != nil {
		return ImportResult{}, err
	}
	if len(chapters) == 0 {
		return ImportResult{}, fmt.Errorf("no chapters parsed from %s", path)
	}
	if options.TitleOverride != "" {
		book.Title = options.TitleOverride
	}
	warnings := ChapterQualityWarnings(chapters)
	existing, err := s.store.GetBookByContentHash(book.ContentHash)
	if err == nil {
		if options.Replace {
			return s.replaceExistingBook(existing.ID, book, chapters, warnings)
		}
		result, err := s.importResultForExistingBook(existing)
		if err != nil {
			return ImportResult{}, err
		}
		result.Warnings = warnings
		return result, nil
	}
	if !errors.Is(err, storage.ErrNotFound) {
		return ImportResult{}, err
	}
	bookID, err := s.store.CreateBook(book)
	if err != nil {
		return ImportResult{}, err
	}
	if err := s.store.InsertChapters(bookID, chapters); err != nil {
		return ImportResult{}, err
	}
	totalWords := 0
	for _, chapter := range chapters {
		totalWords += chapter.WordCount
	}
	return ImportResult{BookID: bookID, Title: book.Title, ChapterCount: len(chapters), WordCount: totalWords, Warnings: warnings}, nil
}

func (s *ImportService) ImportPathWithOptions(path string, options ImportOptions) ([]ImportResult, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		result, err := s.ImportFileWithOptions(path, options)
		if err != nil {
			return nil, err
		}
		return []ImportResult{result}, nil
	}

	files, err := supportedFilesInDirectory(path)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no supported TXT or EPUB files found in %s", path)
	}
	results := make([]ImportResult, 0, len(files))
	for _, file := range files {
		result, err := s.ImportFileWithOptions(file, options)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

func (s *ImportService) ImportURLWithOptions(rawURL string, options ImportOptions) (ImportResult, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return ImportResult{}, fmt.Errorf("parse URL: %w", err)
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return ImportResult{}, fmt.Errorf("unsupported URL scheme %q; supported: http, https", parsedURL.Scheme)
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultHTTPTimeout)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, parsedURL.String(), nil)
	if err != nil {
		return ImportResult{}, fmt.Errorf("create request: %w", err)
	}
	request.Header.Set("User-Agent", "readerx/online-import")

	response, err := importURLHTTPClient.Do(request)
	if err != nil {
		return ImportResult{}, fmt.Errorf("download URL: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return ImportResult{}, fmt.Errorf("download URL: HTTP %d", response.StatusCode)
	}

	filename, err := remoteFilename(parsedURL, response.Header.Get("Content-Type"))
	if err != nil {
		return ImportResult{}, err
	}
	maxBytes := options.MaxDownloadBytes
	if maxBytes <= 0 {
		maxBytes = DefaultMaxDownloadBytes
	}
	if response.ContentLength > maxBytes {
		return ImportResult{}, fmt.Errorf("%w: content length %d exceeds limit %d", ErrDownloadTooLarge, response.ContentLength, maxBytes)
	}

	tempDir, err := os.MkdirTemp("", "readerx-import-url-*")
	if err != nil {
		return ImportResult{}, err
	}
	defer os.RemoveAll(tempDir)

	tempPath := filepath.Join(tempDir, filename)
	file, err := os.Create(tempPath)
	if err != nil {
		return ImportResult{}, err
	}
	written, copyErr := io.Copy(file, io.LimitReader(response.Body, maxBytes+1))
	closeErr := file.Close()
	if closeErr != nil && copyErr == nil {
		copyErr = closeErr
	}
	if copyErr != nil {
		return ImportResult{}, copyErr
	}
	if written > maxBytes {
		return ImportResult{}, fmt.Errorf("%w: downloaded bytes exceed limit %d", ErrDownloadTooLarge, maxBytes)
	}

	return s.ImportFileWithOptions(tempPath, options)
}

func (s *ImportService) replaceExistingBook(bookID int64, book domain.Book, chapters []domain.Chapter, warnings []string) (ImportResult, error) {
	if err := s.store.ReplaceBook(bookID, book, chapters); err != nil {
		return ImportResult{}, err
	}
	totalWords := 0
	for _, chapter := range chapters {
		totalWords += chapter.WordCount
	}
	return ImportResult{BookID: bookID, Title: book.Title, ChapterCount: len(chapters), WordCount: totalWords, Replaced: true, Warnings: warnings}, nil
}

func (s *ImportService) importResultForExistingBook(book domain.Book) (ImportResult, error) {
	chapters, err := s.store.ListChapters(book.ID)
	if err != nil {
		return ImportResult{}, err
	}
	totalWords := 0
	for _, chapter := range chapters {
		totalWords += chapter.WordCount
	}
	return ImportResult{BookID: book.ID, Title: book.Title, ChapterCount: len(chapters), WordCount: totalWords, Existing: true}, nil
}

func supportedFilesInDirectory(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if source.SupportedLocalFile(path) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

func remoteFilename(parsedURL *url.URL, contentType string) (string, error) {
	name := filepath.Base(parsedURL.Path)
	ext := strings.ToLower(filepath.Ext(name))
	if source.SupportedLocalFile(name) {
		return safeRemoteFilename(name), nil
	}

	if mediaType, _, err := mime.ParseMediaType(contentType); err == nil {
		switch strings.ToLower(mediaType) {
		case "text/plain":
			ext = ".txt"
		case "application/epub+zip":
			ext = ".epub"
		}
	}
	if ext == "" {
		return "", fmt.Errorf("unsupported remote file type; supported: .txt, .epub")
	}
	if ext != ".txt" && ext != ".epub" {
		return "", fmt.Errorf("unsupported remote file type %q; supported: .txt, .epub", ext)
	}
	base := strings.TrimSuffix(name, filepath.Ext(name))
	if base == "." || base == "/" || base == "" {
		base = "download"
	}
	return safeRemoteFilename(base + ext), nil
}

func safeRemoteFilename(name string) string {
	name = filepath.Base(name)
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, string(os.PathSeparator), "_")
	if name == "" || name == "." || name == "/" {
		return "download.txt"
	}
	return name
}
