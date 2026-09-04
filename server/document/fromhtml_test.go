package document

import (
	"errors"
	"testing"
)

func TestFromHTML(t *testing.T) {
	cases := []struct {
		name      string
		html      string
		wantURL   string
		wantTitle string
	}{
		{
			name:      "canonical link",
			html:      `<html><head><link rel="canonical" href="https://example.com/a"><title>Title A</title></head></html>`,
			wantURL:   "https://example.com/a",
			wantTitle: "Title A",
		},
		{
			name:      "canonical preferred over og:url",
			html:      `<html><head><link rel="canonical" href="https://example.com/canon"><meta property="og:url" content="https://example.com/og"></head></html>`,
			wantURL:   "https://example.com/canon",
			wantTitle: "",
		},
		{
			name:      "og:url and og:title",
			html:      `<html><head><meta property="og:url" content="https://example.com/b"><meta property="og:title" content="OG Title"><title>HTML Title</title></head></html>`,
			wantURL:   "https://example.com/b",
			wantTitle: "OG Title",
		},
		{
			name:      "og:url via name attribute",
			html:      `<html><head><meta name="og:url" content="https://example.com/c"></head></html>`,
			wantURL:   "https://example.com/c",
			wantTitle: "",
		},
		{
			name:      "twitter:url fallback",
			html:      `<html><head><meta name="twitter:url" content="https://example.com/d"><title>Twitter Page</title></head></html>`,
			wantURL:   "https://example.com/d",
			wantTitle: "Twitter Page",
		},
		{
			name:      "title element fallback and trimming",
			html:      `<html><head><link rel="canonical" href="https://example.com/e"><title>  Spaced Title  </title></head></html>`,
			wantURL:   "https://example.com/e",
			wantTitle: "Spaced Title",
		},
		{
			name:      "blank canonical falls back to og:url",
			html:      `<html><head><link rel="canonical" href="   "><meta property="og:url" content="https://example.com/f"></head></html>`,
			wantURL:   "https://example.com/f",
			wantTitle: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d, err := FromHTML(tc.html)
			if err != nil {
				t.Fatalf("FromHTML() unexpected error: %v", err)
			}
			if d.URL != tc.wantURL {
				t.Errorf("URL = %q, want %q", d.URL, tc.wantURL)
			}
			if d.Title != tc.wantTitle {
				t.Errorf("Title = %q, want %q", d.Title, tc.wantTitle)
			}
			if d.HTML != tc.html {
				t.Errorf("HTML = %q, want %q", d.HTML, tc.html)
			}
		})
	}
}

func TestFromHTMLNoURL(t *testing.T) {
	cases := []struct {
		name string
		html string
	}{
		{"no url at all", `<html><head><title>Just a title</title></head><body>content</body></html>`},
		{"empty input", ``},
		{"blank meta content", `<html><head><meta property="og:url" content=""></head></html>`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d, err := FromHTML(tc.html)
			if !errors.Is(err, ErrNoURL) {
				t.Fatalf("FromHTML() error = %v, want ErrNoURL", err)
			}
			if d != nil {
				t.Errorf("expected nil document, got %+v", d)
			}
		})
	}
}

func TestFromHTMLWithFallbackURL(t *testing.T) {
	html := `<html><head><title>Saved page</title></head><body>content</body></html>`
	d, err := FromHTMLWithFallbackURL(html, "file:///tmp/Saved%20page.html")
	if err != nil {
		t.Fatalf("FromHTMLWithFallbackURL() unexpected error: %v", err)
	}
	if d.URL != "file:///tmp/Saved%20page.html" {
		t.Errorf("URL = %q, want file fallback URL", d.URL)
	}
	if d.Title != "Saved page" {
		t.Errorf("Title = %q, want Saved page", d.Title)
	}
}

func TestFromHTMLWithFallbackURLPrefersHTMLURL(t *testing.T) {
	html := `<html><head><link rel="canonical" href="https://example.com/page"></head></html>`
	d, err := FromHTMLWithFallbackURL(html, "file:///tmp/page.html")
	if err != nil {
		t.Fatalf("FromHTMLWithFallbackURL() unexpected error: %v", err)
	}
	if d.URL != "https://example.com/page" {
		t.Errorf("URL = %q, want canonical URL", d.URL)
	}
}
