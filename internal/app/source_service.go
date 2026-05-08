package app

import (
	"context"
	"encoding/xml"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultSourceSearchLimit = 10
	projectGutenbergSource   = "gutenberg"
	projectGutenbergOPDSURL  = "https://www.gutenberg.org/ebooks/search.opds/"
)

var sourceHTTPClient = &http.Client{Timeout: 30 * time.Second}

type SourceService struct{}

type SourceSearchOptions struct {
	Source string
	Query  string
	Limit  int
}

type SourceSearchResult struct {
	Source  string
	ID      string
	Title   string
	Author  string
	Summary string
	EPUBURL string
	TextURL string
}

func NewSourceService() *SourceService {
	return &SourceService{}
}

func (s *SourceService) ListSources() []string {
	return []string{projectGutenbergSource}
}

func (s *SourceService) Search(ctx context.Context, options SourceSearchOptions) ([]SourceSearchResult, error) {
	sourceName := options.Source
	if sourceName == "" {
		sourceName = projectGutenbergSource
	}
	if sourceName != projectGutenbergSource {
		return nil, fmt.Errorf("unknown source %q; supported: gutenberg", sourceName)
	}
	if strings.TrimSpace(options.Query) == "" {
		return nil, fmt.Errorf("query is required")
	}
	limit := options.Limit
	if limit <= 0 {
		limit = defaultSourceSearchLimit
	}
	return searchProjectGutenberg(ctx, options.Query, limit)
}

func searchProjectGutenberg(ctx context.Context, query string, limit int) ([]SourceSearchResult, error) {
	endpoint, err := url.Parse(projectGutenbergOPDSURL)
	if err != nil {
		return nil, err
	}
	values := endpoint.Query()
	values.Set("query", query)
	endpoint.RawQuery = values.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "readerx/source-search")

	response, err := sourceHTTPClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("search source: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("search source: HTTP %d", response.StatusCode)
	}

	var feed opdsFeed
	if err := xml.NewDecoder(response.Body).Decode(&feed); err != nil {
		return nil, fmt.Errorf("parse OPDS feed: %w", err)
	}
	baseURL := endpoint
	if response.Request != nil && response.Request.URL != nil {
		baseURL = response.Request.URL
	}
	return opdsResults(projectGutenbergSource, baseURL, feed, limit), nil
}

type opdsFeed struct {
	Entries []opdsEntry `xml:"entry"`
}

type opdsEntry struct {
	ID      string     `xml:"id"`
	Title   string     `xml:"title"`
	Summary string     `xml:"summary"`
	Authors []opdsName `xml:"author"`
	Links   []opdsLink `xml:"link"`
}

type opdsName struct {
	Name string `xml:"name"`
}

type opdsLink struct {
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
	Href string `xml:"href,attr"`
}

func opdsResults(sourceName string, baseURL *url.URL, feed opdsFeed, limit int) []SourceSearchResult {
	results := make([]SourceSearchResult, 0, len(feed.Entries))
	for _, entry := range feed.Entries {
		result := SourceSearchResult{
			Source:  sourceName,
			ID:      strings.TrimSpace(entry.ID),
			Title:   strings.TrimSpace(entry.Title),
			Author:  opdsAuthor(entry.Authors),
			Summary: strings.TrimSpace(entry.Summary),
		}
		for _, link := range entry.Links {
			if !strings.Contains(link.Rel, "acquisition") {
				continue
			}
			mediaType, _, err := mime.ParseMediaType(link.Type)
			if err != nil {
				mediaType = strings.ToLower(strings.TrimSpace(link.Type))
			}
			switch strings.ToLower(mediaType) {
			case "application/epub+zip":
				if result.EPUBURL == "" {
					result.EPUBURL = absoluteOPDSURL(baseURL, link.Href)
				}
			case "text/plain":
				if result.TextURL == "" {
					result.TextURL = absoluteOPDSURL(baseURL, link.Href)
				}
			}
		}
		if result.EPUBURL == "" && result.TextURL == "" {
			continue
		}
		results = append(results, result)
		if len(results) >= limit {
			break
		}
	}
	return results
}

func opdsAuthor(authors []opdsName) string {
	names := make([]string, 0, len(authors))
	for _, author := range authors {
		if name := strings.TrimSpace(author.Name); name != "" {
			names = append(names, name)
		}
	}
	return strings.Join(names, ", ")
}

func absoluteOPDSURL(baseURL *url.URL, href string) string {
	parsed, err := url.Parse(strings.TrimSpace(href))
	if err != nil {
		return ""
	}
	return baseURL.ResolveReference(parsed).String()
}
