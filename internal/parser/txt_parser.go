package parser

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/heybox/readerx/internal/domain"
)

var chapterHeadingPattern = regexp.MustCompile(`^\s*((第\s*[0-9一二三四五六七八九十百千万零〇两]+\s*[章节回卷部篇].*)|(卷\s*[0-9一二三四五六七八九十百千万零〇两]+.*)|(Chapter\s+[0-9]+.*))\s*$`)

func ParseTXTFile(path string) (domain.Book, []domain.Chapter, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return domain.Book{}, nil, fmt.Errorf("read txt file: %w", err)
	}
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	book := domain.Book{
		SourceType:  "local_txt",
		Title:       strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)),
		FilePath:    path,
		ContentHash: hashBytes(data),
	}

	chapters := parseChapters(text)
	if len(chapters) == 0 {
		chapters = fallbackChapters(text, 4000)
	}
	for i := range chapters {
		chapters[i].ChapterNo = i + 1
		chapters[i].ContentStatus = "cached"
		chapters[i].WordCount = countWords(chapters[i].Content)
		chapters[i].ContentHash = hashString(chapters[i].Content)
	}
	return book, chapters, nil
}

func parseChapters(text string) []domain.Chapter {
	scanner := bufio.NewScanner(strings.NewReader(text))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var chapters []domain.Chapter
	var current *domain.Chapter
	var lines []string
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), " \t")
		if chapterHeadingPattern.MatchString(line) {
			if current != nil {
				current.Content = cleanText(strings.Join(lines, "\n"))
				chapters = append(chapters, *current)
			}
			current = &domain.Chapter{Title: strings.TrimSpace(line)}
			lines = nil
			continue
		}
		if current != nil {
			lines = append(lines, line)
		}
	}
	if current != nil {
		current.Content = cleanText(strings.Join(lines, "\n"))
		chapters = append(chapters, *current)
	}
	return chapters
}

func fallbackChapters(text string, maxRunes int) []domain.Chapter {
	cleaned := cleanText(text)
	if strings.TrimSpace(cleaned) == "" {
		return []domain.Chapter{{Title: "第 1 章", Content: ""}}
	}
	runes := []rune(cleaned)
	var chapters []domain.Chapter
	for start := 0; start < len(runes); start += maxRunes {
		end := start + maxRunes
		if end > len(runes) {
			end = len(runes)
		}
		chapters = append(chapters, domain.Chapter{
			Title:   fmt.Sprintf("第 %d 章", len(chapters)+1),
			Content: strings.TrimSpace(string(runes[start:end])),
		})
	}
	return chapters
}

func cleanText(text string) string {
	lines := strings.Split(text, "\n")
	out := make([]string, 0, len(lines))
	blank := false
	for _, line := range lines {
		trimmedRight := strings.TrimRight(line, " \t")
		if strings.TrimSpace(trimmedRight) == "" {
			if !blank && len(out) > 0 {
				out = append(out, "")
			}
			blank = true
			continue
		}
		out = append(out, trimmedRight)
		blank = false
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}

func countWords(text string) int {
	count := 0
	for _, r := range text {
		if !unicode.IsSpace(r) {
			count++
		}
	}
	return count
}

func hashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func hashString(text string) string {
	return hashBytes([]byte(text))
}
