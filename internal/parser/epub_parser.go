package parser

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/heybox/readerx/internal/domain"
)

type epubArchive struct {
	reader *zip.ReadCloser
	files  map[string]*zip.File
}

func ParseEPUBFile(filePath string) (domain.Book, []domain.Chapter, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return domain.Book{}, nil, fmt.Errorf("read epub file: %w", err)
	}
	archive, err := openEPUB(filePath)
	if err != nil {
		return domain.Book{}, nil, err
	}
	defer archive.reader.Close()

	opfPath, err := archive.containerOPFPath()
	if err != nil {
		return domain.Book{}, nil, err
	}
	pkg, err := archive.parseOPF(opfPath)
	if err != nil {
		return domain.Book{}, nil, err
	}

	book := domain.Book{
		SourceType:  "local_epub",
		Title:       firstNonEmpty(pkg.Title, strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))),
		Author:      pkg.Creator,
		FilePath:    filePath,
		ContentHash: epubHashBytes(data),
	}

	chapters, err := archive.spineChapters(opfPath, pkg)
	if err != nil {
		return domain.Book{}, nil, err
	}
	return book, chapters, nil
}

func openEPUB(filePath string) (*epubArchive, error) {
	reader, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, fmt.Errorf("open epub zip: %w", err)
	}
	files := make(map[string]*zip.File, len(reader.File))
	for _, file := range reader.File {
		files[file.Name] = file
	}
	return &epubArchive{reader: reader, files: files}, nil
}

func (a *epubArchive) containerOPFPath() (string, error) {
	data, err := a.readFile("META-INF/container.xml")
	if err != nil {
		return "", fmt.Errorf("read container.xml: %w", err)
	}
	var container struct {
		Rootfiles []struct {
			FullPath string `xml:"full-path,attr"`
		} `xml:"rootfiles>rootfile"`
	}
	if err := xml.Unmarshal(data, &container); err != nil {
		return "", fmt.Errorf("parse container.xml: %w", err)
	}
	for _, rootfile := range container.Rootfiles {
		if rootfile.FullPath != "" {
			return rootfile.FullPath, nil
		}
	}
	return "", fmt.Errorf("container.xml has no rootfile")
}

type opfPackage struct {
	Title    string
	Creator  string
	Manifest map[string]string
	Spine    []string
}

func (a *epubArchive) parseOPF(opfPath string) (opfPackage, error) {
	data, err := a.readFile(opfPath)
	if err != nil {
		return opfPackage{}, fmt.Errorf("read opf: %w", err)
	}
	decoder := xml.NewDecoder(bytes.NewReader(data))
	pkg := opfPackage{Manifest: map[string]string{}}
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return opfPackage{}, fmt.Errorf("parse opf: %w", err)
		}
		start, ok := token.(xml.StartElement)
		if !ok {
			continue
		}
		switch start.Name.Local {
		case "title":
			pkg.Title = strings.TrimSpace(readElementText(decoder, start))
		case "creator":
			pkg.Creator = strings.TrimSpace(readElementText(decoder, start))
		case "item":
			id, href := attr(start, "id"), attr(start, "href")
			if id != "" && href != "" {
				pkg.Manifest[id] = href
			}
		case "itemref":
			idref := attr(start, "idref")
			if idref != "" {
				pkg.Spine = append(pkg.Spine, idref)
			}
		}
	}
	if len(pkg.Spine) == 0 {
		return opfPackage{}, fmt.Errorf("opf spine is empty")
	}
	return pkg, nil
}

func (a *epubArchive) spineChapters(opfPath string, pkg opfPackage) ([]domain.Chapter, error) {
	baseDir := path.Dir(opfPath)
	if baseDir == "." {
		baseDir = ""
	}
	chapters := make([]domain.Chapter, 0, len(pkg.Spine))
	for _, idref := range pkg.Spine {
		href := pkg.Manifest[idref]
		if href == "" {
			continue
		}
		chapterPath := path.Clean(path.Join(baseDir, href))
		data, err := a.readFile(chapterPath)
		if err != nil {
			return nil, fmt.Errorf("read spine item %s: %w", chapterPath, err)
		}
		title, content := extractXHTMLText(data)
		title, content = promoteFirstHeadingLine(title, content)
		wordCount := countEPUBWords(content)
		if wordCount == 0 {
			continue
		}
		if title == "" {
			title = fmt.Sprintf("第 %d 章", len(chapters)+1)
		}
		chapter := domain.Chapter{
			ChapterNo:       len(chapters) + 1,
			SourceChapterID: chapterPath,
			Title:           title,
			Content:         content,
			ContentStatus:   "cached",
			ContentHash:     epubHashString(content),
			WordCount:       wordCount,
		}
		chapters = append(chapters, chapter)
	}
	if len(chapters) == 0 {
		return nil, fmt.Errorf("no readable chapters found in spine")
	}
	return chapters, nil
}

func (a *epubArchive) readFile(name string) ([]byte, error) {
	file := a.files[name]
	if file == nil {
		return nil, fmt.Errorf("file not found: %s", name)
	}
	reader, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}

func readElementText(decoder *xml.Decoder, start xml.StartElement) string {
	var builder strings.Builder
	depth := 1
	for depth > 0 {
		token, err := decoder.Token()
		if err != nil {
			break
		}
		switch token := token.(type) {
		case xml.CharData:
			builder.Write([]byte(token))
		case xml.StartElement:
			depth++
		case xml.EndElement:
			depth--
		}
	}
	return builder.String()
}

func attr(start xml.StartElement, name string) string {
	for _, attr := range start.Attr {
		if attr.Name.Local == name {
			return attr.Value
		}
	}
	return ""
}

var (
	headingRE = regexp.MustCompile(`(?is)<h[1-6][^>]*>(.*?)</h[1-6]>`)
	bodyRE    = regexp.MustCompile(`(?is)<body[^>]*>(.*?)</body>`)
	blockRE   = regexp.MustCompile(`(?i)</?(p|div|section|article|br|h[1-6])\b[^>]*>`)
	tagRE     = regexp.MustCompile(`(?is)<[^>]+>`)
	spaceRE   = regexp.MustCompile(`[ \t]+`)
	newlineRE = regexp.MustCompile(`\n{3,}`)
)

func extractXHTMLText(data []byte) (string, string) {
	raw := string(data)
	title := ""
	if match := headingRE.FindStringSubmatch(raw); len(match) > 1 {
		title = cleanInlineHTML(match[1])
	}
	if match := bodyRE.FindStringSubmatch(raw); len(match) > 1 {
		raw = match[1]
	}
	raw = blockRE.ReplaceAllString(raw, "\n")
	raw = tagRE.ReplaceAllString(raw, "")
	text := html.UnescapeString(raw)
	lines := strings.Split(text, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(spaceRE.ReplaceAllString(line, " "))
		if line != "" && line != title {
			out = append(out, line)
		}
	}
	return title, strings.TrimSpace(newlineRE.ReplaceAllString(strings.Join(out, "\n"), "\n\n"))
}

func promoteFirstHeadingLine(title, content string) (string, string) {
	if title != "" {
		return title, content
	}
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		candidate := strings.TrimSpace(line)
		if candidate == "" {
			continue
		}
		if !chapterHeadingPattern.MatchString(candidate) {
			return title, content
		}
		remaining := append([]string{}, lines[:i]...)
		remaining = append(remaining, lines[i+1:]...)
		return candidate, strings.TrimSpace(strings.Join(remaining, "\n"))
	}
	return title, content
}

func cleanInlineHTML(raw string) string {
	raw = tagRE.ReplaceAllString(raw, "")
	return strings.TrimSpace(html.UnescapeString(raw))
}

func countEPUBWords(text string) int {
	count := 0
	for _, r := range text {
		if !unicode.IsSpace(r) {
			count++
		}
	}
	return count
}

func epubHashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func epubHashString(text string) string {
	return epubHashBytes([]byte(text))
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
