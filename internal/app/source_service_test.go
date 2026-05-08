package app

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestSourceServiceSearchesProjectGutenbergOPDS(t *testing.T) {
	withSourceHTTPClient(t, roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if !strings.Contains(r.URL.RawQuery, "alice") {
			t.Fatalf("query = %q, want alice", r.URL.RawQuery)
		}
		if got := r.Header.Get("User-Agent"); got == "" {
			t.Fatalf("User-Agent is empty")
		}
		body := `<?xml version="1.0" encoding="utf-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <entry>
    <id>ebooks/11</id>
    <title>Alice's Adventures in Wonderland</title>
    <author><name>Lewis Carroll</name></author>
    <summary>A girl follows a white rabbit.</summary>
    <link rel="http://opds-spec.org/acquisition/open-access" type="application/epub+zip" href="/ebooks/11.epub.images" />
    <link rel="http://opds-spec.org/acquisition/open-access" type="text/plain; charset=utf-8" href="/files/11/11-0.txt" />
  </entry>
  <entry>
    <id>ebooks/999</id>
    <title>Unsupported</title>
    <link rel="alternate" type="text/html" href="/ebooks/999" />
  </entry>
</feed>`
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/atom+xml"}},
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    r,
		}, nil
	}))

	results, err := NewSourceService().Search(context.Background(), SourceSearchOptions{
		Source: "gutenberg",
		Query:  "alice",
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("result count = %d, want 1: %#v", len(results), results)
	}
	result := results[0]
	if result.Source != "gutenberg" || result.Title != "Alice's Adventures in Wonderland" || result.Author != "Lewis Carroll" {
		t.Fatalf("result metadata = %#v", result)
	}
	if result.EPUBURL != "https://www.gutenberg.org/ebooks/11.epub.images" {
		t.Fatalf("EPUBURL = %q", result.EPUBURL)
	}
	if result.TextURL != "https://www.gutenberg.org/files/11/11-0.txt" {
		t.Fatalf("TextURL = %q", result.TextURL)
	}
}

func TestSourceServiceRejectsUnknownSource(t *testing.T) {
	_, err := NewSourceService().Search(context.Background(), SourceSearchOptions{Source: "unknown", Query: "alice"})
	if err == nil {
		t.Fatalf("expected unknown source error")
	}
}

func withSourceHTTPClient(t *testing.T, transport http.RoundTripper) {
	t.Helper()
	previousClient := sourceHTTPClient
	sourceHTTPClient = &http.Client{Transport: transport}
	t.Cleanup(func() {
		sourceHTTPClient = previousClient
	})
}
