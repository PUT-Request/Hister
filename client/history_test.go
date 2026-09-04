package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchHistoryRequestsAndDecodesOpenedHistory(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/history" {
			t.Errorf("request = %s %s", r.Method, r.URL.Path)
		}
		if opened := r.URL.Query().Get("opened"); opened != "true" {
			t.Errorf("opened query parameter = %q, want true", opened)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"documents": []map[string]any{{
				"id":        42,
				"query":     "history query",
				"title":     "Opened result",
				"url":       "https://example.com/result",
				"added":     1_723_456_789,
				"add_count": 3,
			}},
			"last_id":         42,
			"last_updated_at": "2026-08-16T12:00:00Z",
		})
	}))
	defer server.Close()

	items, err := New(server.URL).FetchHistory()
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("history items = %#v, want one item", items)
	}
	item := items[0]
	if item.Query != "history query" || item.Title != "Opened result" || item.URL != "https://example.com/result" {
		t.Fatalf("history item = %#v", item)
	}
}
