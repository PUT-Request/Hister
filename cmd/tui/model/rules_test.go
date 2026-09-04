package model

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/asciimoo/hister/client"
	"github.com/asciimoo/hister/config"
)

func TestPrioritizeRulePreservesRulesNotLoadedInTUI(t *testing.T) {
	var saved url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(client.RulesResponse{
				Skip:       []string{"private/*"},
				Priority:   []string{"existing/*"},
				Versioning: []string{"articles/*"},
			})
		case http.MethodPost:
			if err := r.ParseForm(); err != nil {
				t.Error(err)
			}
			saved = r.PostForm
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer server.Close()

	m := InitialModel(config.CreateDefaultConfig())
	m.Client = client.New(server.URL)
	msg := m.PrioritizeRuleCmd("new/*")().(RulesSavedMsg)
	if msg.Err != nil {
		t.Fatal(msg.Err)
	}
	if got := saved.Get("skip"); got != "private/*" {
		t.Fatalf("skip = %q", got)
	}
	if got := saved.Get("priority"); got != "existing/*\nnew/*" {
		t.Fatalf("priority = %q", got)
	}
	if got := saved.Get("versioning"); got != "articles/*" {
		t.Fatalf("versioning = %q", got)
	}
}
