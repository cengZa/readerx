package domain

type Book struct {
	ID                 int64
	SourceType         string
	SourceBookID       string
	Title              string
	Author             string
	Description        string
	CoverURL           string
	FilePath           string
	ContentHash        string
	CreatedAt          int64
	UpdatedAt          int64
	LastReadAt         int64
	ChapterCount       int
	ProgressPercentage float64
}

type Chapter struct {
	ID              int64
	BookID          int64
	ChapterNo       int
	SourceChapterID string
	Title           string
	Content         string
	ContentStatus   string
	ContentHash     string
	WordCount       int
	CreatedAt       int64
	UpdatedAt       int64
}

type Progress struct {
	BookID     int64
	ChapterNo  int
	LineOffset int
	CharOffset int
	Percentage float64
	UpdatedAt  int64
}

type Bookmark struct {
	ID           int64
	BookID       int64
	BookTitle    string
	ChapterNo    int
	ChapterTitle string
	LineOffset   int
	CharOffset   int
	Excerpt      string
	Note         string
	CreatedAt    int64
}

type SearchResult struct {
	BookID       int64
	BookTitle    string
	ChapterNo    int
	ChapterTitle string
	Snippet      string
}
