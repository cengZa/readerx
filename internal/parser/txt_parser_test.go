package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseTXTRecognizesChineseChapters(t *testing.T) {
	path := writeTempTXT(t, "剑来.txt", "序言\n\n第一章 山雨欲来\n第一段内容。\n\n第二章 剑气\n第二段内容。")

	book, chapters, err := ParseTXTFile(path)
	if err != nil {
		t.Fatalf("ParseTXTFile returned error: %v", err)
	}
	if book.Title != "剑来" {
		t.Fatalf("book title = %q, want 剑来", book.Title)
	}
	if len(chapters) != 2 {
		t.Fatalf("chapter count = %d, want 2", len(chapters))
	}
	if chapters[0].Title != "第一章 山雨欲来" {
		t.Fatalf("first chapter title = %q", chapters[0].Title)
	}
	if chapters[0].Content == "" || chapters[1].Content == "" {
		t.Fatalf("chapter content should be preserved")
	}
}

func TestParseTXTRecognizesEnglishChapters(t *testing.T) {
	path := writeTempTXT(t, "story.txt", "Chapter 1 Start\nhello\n\nChapter 2 End\nworld")

	_, chapters, err := ParseTXTFile(path)
	if err != nil {
		t.Fatalf("ParseTXTFile returned error: %v", err)
	}
	if len(chapters) != 2 {
		t.Fatalf("chapter count = %d, want 2", len(chapters))
	}
	if chapters[1].Title != "Chapter 2 End" {
		t.Fatalf("second chapter title = %q", chapters[1].Title)
	}
}

func TestParseTXTFallsBackWhenNoChapterHeadings(t *testing.T) {
	path := writeTempTXT(t, "notes.txt", "没有章节标题。\n这里只是一段很短的文本。")

	_, chapters, err := ParseTXTFile(path)
	if err != nil {
		t.Fatalf("ParseTXTFile returned error: %v", err)
	}
	if len(chapters) != 1 {
		t.Fatalf("chapter count = %d, want 1", len(chapters))
	}
	if chapters[0].Title != "第 1 章" {
		t.Fatalf("fallback title = %q, want 第 1 章", chapters[0].Title)
	}
}

func writeTempTXT(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp txt: %v", err)
	}
	return path
}
