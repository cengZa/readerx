package parser

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestParseEPUBReadsMetadataAndSpineChapters(t *testing.T) {
	path := writeTestEPUB(t, "sample.epub")

	book, chapters, err := ParseEPUBFile(path)
	if err != nil {
		t.Fatalf("ParseEPUBFile returned error: %v", err)
	}
	if book.Title != "测试 EPUB" {
		t.Fatalf("book title = %q, want 测试 EPUB", book.Title)
	}
	if book.Author != "林澈" {
		t.Fatalf("book author = %q, want 林澈", book.Author)
	}
	if book.SourceType != "local_epub" {
		t.Fatalf("source type = %q, want local_epub", book.SourceType)
	}
	if len(chapters) != 2 {
		t.Fatalf("chapter count = %d, want 2", len(chapters))
	}
	if chapters[0].Title != "第一章 山雨欲来" || chapters[1].Title != "第二章 白鹿渡" {
		t.Fatalf("chapter titles = %#v", []string{chapters[0].Title, chapters[1].Title})
	}
	if chapters[0].Content != "雨落在青石巷里。\n一道剑气自山巅而起。" {
		t.Fatalf("first chapter content = %q", chapters[0].Content)
	}
}

func writeTestEPUB(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create epub: %v", err)
	}
	defer file.Close()
	writer := zip.NewWriter(file)
	defer writer.Close()

	writeZipFile(t, writer, "META-INF/container.xml", `<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`)
	writeZipFile(t, writer, "OEBPS/content.opf", `<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" unique-identifier="BookId" version="2.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>测试 EPUB</dc:title>
    <dc:creator>林澈</dc:creator>
  </metadata>
  <manifest>
    <item id="c1" href="chapters/ch1.xhtml" media-type="application/xhtml+xml"/>
    <item id="c2" href="chapters/ch2.xhtml" media-type="application/xhtml+xml"/>
  </manifest>
  <spine>
    <itemref idref="c1"/>
    <itemref idref="c2"/>
  </spine>
</package>`)
	writeZipFile(t, writer, "OEBPS/chapters/ch1.xhtml", `<?xml version="1.0"?>
<html xmlns="http://www.w3.org/1999/xhtml"><head><title>ignored</title></head><body>
<h1>第一章 山雨欲来</h1>
<p>雨落在青石巷里。</p>
<p>一道剑气自山巅而起。</p>
</body></html>`)
	writeZipFile(t, writer, "OEBPS/chapters/ch2.xhtml", `<?xml version="1.0"?>
<html xmlns="http://www.w3.org/1999/xhtml"><body>
<h2>第二章 白鹿渡</h2>
<p>渡船老人在船头煮茶。</p>
</body></html>`)
	return path
}

func writeZipFile(t *testing.T, writer *zip.Writer, name, content string) {
	t.Helper()
	file, err := writer.Create(name)
	if err != nil {
		t.Fatalf("create zip file %s: %v", name, err)
	}
	if _, err := file.Write([]byte(content)); err != nil {
		t.Fatalf("write zip file %s: %v", name, err)
	}
}
