package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/asciimoo/hister/client"
	"github.com/asciimoo/hister/server/document"
)

func TestServiceImportBufferDownloadsMissingFavicon(t *testing.T) {
	var received *document.Document
	faviconDownloads := 0
	targetHTTPClient := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		var body struct {
			Ops []struct {
				*document.Document
			} `json:"ops"`
		}
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			return nil, err
		}
		if len(body.Ops) != 1 {
			return nil, fmt.Errorf("received %d batch operations, want one", len(body.Ops))
		}
		received = body.Ops[0].Document
		return jsonHTTPResponse(req, http.StatusOK, `{"results":[{"status":201}]}`), nil
	})}
	target := client.New("http://hister.example", client.WithHTTPClient(targetHTTPClient), client.WithMaxBatchBodyBytes(40<<20))
	buffer, err := newServiceImportBuffer(
		"test",
		target,
		document.NewNullLanguageDetector(),
		nil,
		serviceImportOptions{
			BatchSize: 1,
			FaviconDownloader: func(d *document.Document) error {
				faviconDownloads++
				d.Favicon = "data:image/png;base64,ZGVmYXVsdCBpY29u"
				return nil
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	d := &document.Document{
		URL:       "https://example.com/article",
		Title:     "Article",
		Text:      "Contents",
		Processed: true,
	}
	buffer.Add(context.Background(), d, nil)

	if received == nil {
		t.Fatal("no document was submitted")
	}
	if received.Favicon != "data:image/png;base64,ZGVmYXVsdCBpY29u" {
		t.Errorf("favicon = %q, want downloaded default icon", received.Favicon)
	}
	if faviconDownloads != 1 {
		t.Errorf("favicon downloads = %d, want one", faviconDownloads)
	}
	if buffer.stats.Imported != 1 || buffer.stats.Errors != 0 {
		t.Errorf("stats = %+v, want one import without errors", buffer.stats)
	}
}

func TestApplyServiceContentPreservesFetchedFavicon(t *testing.T) {
	d := &document.Document{URL: "https://example.com/article"}
	fetched := &document.Document{
		URL:     "https://example.com/article",
		HTML:    `<html><head><title>Article</title></head><body><main><p>Downloaded contents.</p></main></body></html>`,
		Favicon: "data:image/png;base64,bGlua2VkIGljb24=",
	}
	if err := applyServiceContent(context.Background(), d, fetched, "", "", document.NewNullLanguageDetector()); err != nil {
		t.Fatal(err)
	}
	if d.Favicon != fetched.Favicon {
		t.Errorf("favicon = %q, want fetched favicon %q", d.Favicon, fetched.Favicon)
	}
}
