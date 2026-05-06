package source

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestImportLocalFileSupportsEPUB(t *testing.T) {
	path := parserTestEPUB(t)

	book, chapters, err := ImportLocalFile(path)
	if err != nil {
		t.Fatalf("ImportLocalFile returned error: %v", err)
	}
	if book.SourceType != "local_epub" {
		t.Fatalf("source type = %q, want local_epub", book.SourceType)
	}
	if len(chapters) != 2 {
		t.Fatalf("chapter count = %d, want 2", len(chapters))
	}
}

func parserTestEPUB(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "sample.epub")
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create epub: %v", err)
	}
	defer file.Close()
	writer := zip.NewWriter(file)
	defer writer.Close()
	writeSourceZipFile(t, writer, "META-INF/container.xml", `<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container"><rootfiles><rootfile full-path="content.opf"/></rootfiles></container>`)
	writeSourceZipFile(t, writer, "content.opf", `<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf"><metadata xmlns:dc="http://purl.org/dc/elements/1.1/"><dc:title>Source EPUB</dc:title></metadata><manifest><item id="c1" href="ch1.xhtml" media-type="application/xhtml+xml"/><item id="c2" href="ch2.xhtml" media-type="application/xhtml+xml"/></manifest><spine><itemref idref="c1"/><itemref idref="c2"/></spine></package>`)
	writeSourceZipFile(t, writer, "ch1.xhtml", `<html><body><h1>第一章</h1><p>内容一</p></body></html>`)
	writeSourceZipFile(t, writer, "ch2.xhtml", `<html><body><h1>第二章</h1><p>内容二</p></body></html>`)
	return path
}

func writeSourceZipFile(t *testing.T, writer *zip.Writer, name, content string) {
	t.Helper()
	file, err := writer.Create(name)
	if err != nil {
		t.Fatalf("create zip file %s: %v", name, err)
	}
	if _, err := file.Write([]byte(content)); err != nil {
		t.Fatalf("write zip file %s: %v", name, err)
	}
}
