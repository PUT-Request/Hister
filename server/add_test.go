package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/asciimoo/hister/server/document"
	"github.com/asciimoo/hister/server/indexer"
	"github.com/asciimoo/hister/server/testutil"
)

type statusCountingWriter struct {
	http.ResponseWriter
	statusCodes []int
}

func (w *statusCountingWriter) WriteHeader(statusCode int) {
	w.statusCodes = append(w.statusCodes, statusCode)
	w.ResponseWriter.WriteHeader(statusCode)
}

func TestAddFormAccessTokenDoesNotAuthenticate(t *testing.T) {
	cfg := testutil.Config(t)
	cfg.App.AccessToken = "secret"
	sessionStore = newSessionStore([]byte(strings.Repeat("x", 32)), cfg.BaseURL(""), sessionMaxAge)
	called := false
	h := endpointHandler(func(c *webContext) {
		called = true
		c.Response.WriteHeader(http.StatusNoContent)
	})
	h = withCSRF(h)
	h = withTokenAuth(h)
	handler := http.HandlerFunc(createHandler(cfg, nil, h))
	body := url.Values{
		"access_token": {"secret"},
		"url":          {"https://example.com"},
	}.Encode()

	rec := testutil.ServeHTTP(
		t,
		handler,
		http.MethodPost,
		"/api/add",
		strings.NewReader(body),
		map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
			"Origin":       "https://unrelated.example",
		},
	)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("POST /api/add status = %d, want %d", rec.Code, http.StatusForbidden)
	}
	if called {
		t.Fatal("POST /api/add with a form access token reached the protected handler")
	}
}

func TestServeAddFormWritesStatusOnce(t *testing.T) {
	cfg := testutil.Config(t)
	cfg.Server.BaseURL = "http://127.0.0.1:4433"
	body := url.Values{
		"url": {"http://127.0.0.1:4433/already-local"},
	}.Encode()
	req := httptest.NewRequest(http.MethodPost, "/api/add", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	writer := &statusCountingWriter{ResponseWriter: rec}

	serveAdd(&webContext{
		Request:  req,
		Response: writer,
		Config:   cfg,
	})

	if len(writer.statusCodes) != 1 {
		t.Fatalf("WriteHeader calls = %v, want one call", writer.statusCodes)
	}
	if writer.statusCodes[0] != http.StatusNotAcceptable {
		t.Fatalf("status = %d, want %d", writer.statusCodes[0], http.StatusNotAcceptable)
	}
}

func TestServeAddIndexesRemoteFileSnapshot(t *testing.T) {
	cfg := testutil.Config(t)
	idx, err := indexer.New(cfg)
	if err != nil {
		t.Fatalf("create indexer: %v", err)
	}
	t.Cleanup(idx.Close)

	d := &document.Document{
		URL:       "remote-file://workstation/home/alice/note.md",
		Text:      "Remote note\n\nExtracted body",
		Title:     "Remote note",
		Label:     "notes",
		Updated:   1234,
		Type:      document.RemoteFile,
		Domain:    "untrusted.example",
		Processed: true,
	}
	body, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	serveAdd(&webContext{Request: req, Response: rec, Config: cfg, Indexer: idx})

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	got := idx.GetByURLAndUser(d.URL, 0)
	if got == nil {
		t.Fatal("remote file was not indexed")
	}
	if got.Type != document.RemoteFile {
		t.Errorf("type = %v, want remote", got.Type)
	}
	if got.Domain != "workstation" {
		t.Errorf("domain = %q, want workstation", got.Domain)
	}
	if got.Title != "Remote note" || got.Text != d.Text {
		t.Errorf("stored content = %#v, want submitted title and text", got)
	}
	if got.Updated != 1234 || got.Label != "notes" {
		t.Errorf("stored metadata = %#v, want updated time and label", got)
	}
}

func TestServeAddValidatesRemoteFileSnapshot(t *testing.T) {
	tests := []struct {
		name string
		doc  document.Document
	}{
		{
			name: "remote type requires remote URL",
			doc:  document.Document{URL: "file:///tmp/note.txt", Text: "note", Type: document.RemoteFile},
		},
		{
			name: "remote URL requires remote type",
			doc:  document.Document{URL: "remote-file://client/tmp/note.txt", Text: "note", Type: document.Web},
		},
		{
			name: "source host is required",
			doc:  document.Document{URL: "remote-file:///tmp/note.txt", Text: "note", Type: document.RemoteFile},
		},
		{
			name: "path is required",
			doc:  document.Document{URL: "remote-file://client/", Text: "note", Type: document.RemoteFile},
		},
		{
			name: "content is required",
			doc:  document.Document{URL: "remote-file://client/tmp/note.txt", Type: document.RemoteFile},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := testutil.Config(t)
			body, err := json.Marshal(tt.doc)
			if err != nil {
				t.Fatal(err)
			}
			req := httptest.NewRequest(http.MethodPost, "/api/add", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			serveAdd(&webContext{Request: req, Response: rec, Config: cfg})

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusBadRequest, rec.Body.String())
			}
		})
	}
}

func TestServeAddFormRejectsRemoteFileURLWithoutType(t *testing.T) {
	cfg := testutil.Config(t)
	body := url.Values{
		"url":  {"remote-file://client/tmp/note.txt"},
		"text": {"note"},
	}.Encode()
	req := httptest.NewRequest(http.MethodPost, "/api/add", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	serveAdd(&webContext{Request: req, Response: rec, Config: cfg})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestServeAddRechecksSensitiveRemoteFileContent(t *testing.T) {
	cfg := testutil.Config(t)
	idx, err := indexer.New(cfg)
	if err != nil {
		t.Fatalf("create indexer: %v", err)
	}
	t.Cleanup(idx.Close)

	d := document.Document{
		URL:       "remote-file://client/tmp/key.txt",
		Text:      "-----BEGIN OPENSSH PRIVATE KEY-----",
		Type:      document.RemoteFile,
		Processed: true,
	}
	body, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	serveAdd(&webContext{Request: req, Response: rec, Config: cfg, Indexer: idx})

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusUnprocessableEntity, rec.Body.String())
	}
}
