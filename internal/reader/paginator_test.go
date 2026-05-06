package reader

import "testing"

func TestWrapTextRespectsCJKDisplayWidth(t *testing.T) {
	lines := WrapText("你好世界", 4)
	if len(lines) != 2 {
		t.Fatalf("line count = %d, want 2: %#v", len(lines), lines)
	}
	if lines[0] != "你好" || lines[1] != "世界" {
		t.Fatalf("lines = %#v", lines)
	}
}

func TestPaginatorMovesByPageAndClamps(t *testing.T) {
	p := NewPaginator("a\nb\nc\nd\ne", 20, 2)

	if got := p.VisibleLines(0); len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("visible at top = %#v", got)
	}
	if next := p.NextPage(0); next != 2 {
		t.Fatalf("next page offset = %d, want 2", next)
	}
	if last := p.NextPage(2); last != 4 {
		t.Fatalf("last page offset = %d, want 4", last)
	}
	if clamped := p.NextPage(4); clamped != 4 {
		t.Fatalf("clamped next page offset = %d, want 4", clamped)
	}
	page, total := p.PageInfo(4)
	if page != 3 || total != 3 {
		t.Fatalf("page info at last page = %d/%d, want 3/3", page, total)
	}
	if prev := p.PrevPage(1); prev != 0 {
		t.Fatalf("previous page offset = %d, want 0", prev)
	}
}

func TestPaginatorMapsLineOffsetToCharOffset(t *testing.T) {
	p := NewPaginator("alpha\nbeta\ngamma", 20, 2)

	if got := p.CharOffsetForLine(1); got != len("alpha\n") {
		t.Fatalf("char offset = %d, want %d", got, len("alpha\n"))
	}
}
