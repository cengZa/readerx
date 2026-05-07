package app

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"

	"github.com/heybox/readerx/internal/domain"
)

var chapterNumberPattern = regexp.MustCompile(`第\s*([0-9]+)\s*[章节回]`)

// ChapterQualityWarnings reports likely chapter ordering issues without changing imported content.
func ChapterQualityWarnings(chapters []domain.Chapter) []string {
	type numberedChapter struct {
		index  int
		number int
	}

	numbered := make([]numberedChapter, 0, len(chapters))
	counts := map[int]int{}
	for i, chapter := range chapters {
		number, ok := chapterTitleNumber(chapter.Title)
		if !ok {
			continue
		}
		numbered = append(numbered, numberedChapter{index: i + 1, number: number})
		counts[number]++
	}
	if len(numbered) == 0 {
		return nil
	}

	var warnings []string
	for _, number := range sortedDuplicateNumbers(counts) {
		warnings = append(warnings, fmt.Sprintf("章节编号可能重复：第%d章出现%d次。", number, counts[number]))
	}

	for i := 1; i < len(numbered); i++ {
		previous := numbered[i-1]
		current := numbered[i]
		if current.number <= previous.number {
			warnings = append(warnings, fmt.Sprintf("章节编号顺序可能异常：第%d个章节标题编号为%d，前一个编号为%d。", current.index, current.number, previous.number))
		}
		if current.number > previous.number+1 {
			warnings = append(warnings, fmt.Sprintf("章节编号可能缺失：从第%d章后跳到第%d章。", previous.number, current.number))
		}
	}
	return warnings
}

func chapterTitleNumber(title string) (int, bool) {
	matches := chapterNumberPattern.FindStringSubmatch(title)
	if len(matches) != 2 {
		return 0, false
	}
	number, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, false
	}
	return number, true
}

func sortedDuplicateNumbers(counts map[int]int) []int {
	var numbers []int
	for number, count := range counts {
		if count > 1 {
			numbers = append(numbers, number)
		}
	}
	sort.Ints(numbers)
	return numbers
}
