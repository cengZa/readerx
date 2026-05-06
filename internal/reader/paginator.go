package reader

import (
	"strings"

	"github.com/mattn/go-runewidth"
)

type Paginator struct {
	lines      []string
	lineStarts []int
	pageHeight int
}

func NewPaginator(text string, width, pageHeight int) Paginator {
	if width < 1 {
		width = 1
	}
	if pageHeight < 1 {
		pageHeight = 1
	}
	lines := WrapText(text, width)
	starts := make([]int, len(lines))
	offset := 0
	for i, line := range lines {
		starts[i] = offset
		offset += len([]rune(line)) + 1
	}
	return Paginator{lines: lines, lineStarts: starts, pageHeight: pageHeight}
}

func (p Paginator) VisibleLines(lineOffset int) []string {
	lineOffset = p.ClampOffset(lineOffset)
	end := lineOffset + p.pageHeight
	if end > len(p.lines) {
		end = len(p.lines)
	}
	return p.lines[lineOffset:end]
}

func (p Paginator) NextLine(lineOffset int) int {
	return p.ClampOffset(lineOffset + 1)
}

func (p Paginator) PrevLine(lineOffset int) int {
	return p.ClampOffset(lineOffset - 1)
}

func (p Paginator) NextPage(lineOffset int) int {
	return p.ClampOffset(lineOffset + p.pageHeight)
}

func (p Paginator) PrevPage(lineOffset int) int {
	return p.ClampOffset(lineOffset - p.pageHeight)
}

func (p Paginator) ClampOffset(lineOffset int) int {
	maxOffset := len(p.lines) - 1
	if maxOffset < 0 {
		maxOffset = 0
	}
	if lineOffset < 0 {
		return 0
	}
	if lineOffset > maxOffset {
		return maxOffset
	}
	return lineOffset
}

func (p Paginator) CharOffsetForLine(lineOffset int) int {
	lineOffset = p.ClampOffset(lineOffset)
	if lineOffset >= len(p.lineStarts) {
		return 0
	}
	return p.lineStarts[lineOffset]
}

func (p Paginator) PageInfo(lineOffset int) (int, int) {
	if len(p.lines) == 0 {
		return 1, 1
	}
	page := p.ClampOffset(lineOffset)/p.pageHeight + 1
	total := (len(p.lines) + p.pageHeight - 1) / p.pageHeight
	return page, total
}

func WrapText(text string, width int) []string {
	if width < 1 {
		width = 1
	}
	paragraphs := strings.Split(text, "\n")
	lines := make([]string, 0, len(paragraphs))
	for _, paragraph := range paragraphs {
		if strings.TrimSpace(paragraph) == "" {
			lines = append(lines, "")
			continue
		}
		lines = append(lines, wrapLine(paragraph, width)...)
	}
	if len(lines) == 0 {
		return []string{""}
	}
	return lines
}

func wrapLine(line string, width int) []string {
	var result []string
	var builder strings.Builder
	currentWidth := 0
	for _, r := range line {
		rw := runewidth.RuneWidth(r)
		if currentWidth > 0 && currentWidth+rw > width {
			result = append(result, builder.String())
			builder.Reset()
			currentWidth = 0
		}
		builder.WriteRune(r)
		currentWidth += rw
	}
	result = append(result, builder.String())
	return result
}
