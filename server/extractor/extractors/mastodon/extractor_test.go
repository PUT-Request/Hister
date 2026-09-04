// SPDX-License-Identifier: AGPL-3.0-or-later

package mastodon

import (
	"testing"

	"github.com/asciimoo/hister/config"
	"github.com/asciimoo/hister/server/document"
	"github.com/asciimoo/hister/server/extractor/sdk"
)

func TestSetConfigRejectsUnknownOptions(t *testing.T) {
	e := &MastodonExtractor{}
	err := e.SetConfig(&config.Extractor{
		Enable:  true,
		Options: map[string]any{"unknown": true},
	})
	if err == nil {
		t.Fatal("SetConfig accepted an unknown option")
	}
}

func TestOriginalStatusURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "remote status",
			url:  "https://chaos.social/@hovav@infosec.exchange/116976714107843531",
			want: "https://infosec.exchange/@hovav/116976714107843531",
		},
		{
			name: "local status",
			url:  "https://chaos.social/@alice/123",
			want: "https://chaos.social/@alice/123",
		},
		{
			name: "unrelated path",
			url:  "https://chaos.social/tags/hister",
			want: "https://chaos.social/tags/hister",
		},
		{
			name: "invalid remote host",
			url:  "https://chaos.social/@alice@example.com%3Fredirect=attacker.example/123",
			want: "https://chaos.social/@alice@example.com%3Fredirect=attacker.example/123",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := originalStatusURL(test.url); got != test.want {
				t.Fatalf("originalStatusURL(%q) = %q, want %q", test.url, got, test.want)
			}
		})
	}
}

func TestExtractUsesOriginalRemoteStatusURL(t *testing.T) {
	d := &document.Document{
		URL: "https://chaos.social/public/local",
		HTML: `<div class="status">
			<a class="status__relative-time" href="/@hovav@infosec.exchange/116976714107843531"></a>
			<span class="display-name">Hovav</span>
			<div class="status__content"><p>Remote toot</p></div>
		</div>`,
	}

	state, err := (&MastodonExtractor{}).Extract(d).Unpack()
	if err != nil {
		t.Fatalf("Extract returned an error: %v", err)
	}
	if state != sdk.ExtractorSuccess {
		t.Fatalf("Extract state = %v, want %v", state, sdk.ExtractorSuccess)
	}
	if len(d.ExtraDocuments) != 1 {
		t.Fatalf("ExtraDocuments length = %d, want 1", len(d.ExtraDocuments))
	}
	if got, want := d.ExtraDocuments[0].URL, "https://infosec.exchange/@hovav/116976714107843531"; got != want {
		t.Fatalf("toot URL = %q, want %q", got, want)
	}
}

func TestExtractRenderedDetailedStatus(t *testing.T) {
	d := &document.Document{
		URL: "https://mastodon.social/@Mastodon/114818205314251502",
		HTML: `<meta content='{"repository":"mastodon/mastodon"}'>
		<div class="detailed-status">
			<span class="display-name">Mastodon @Mastodon@mastodon.social</span>
			<div class="status__content"><p>Mastodon 4.4 is now available.</p></div>
			<a class="detailed-status__datetime" href="/@Mastodon/114818205314251502">
				<time datetime="2025-07-08T14:59:35.443Z">Jul 8, 2025</time>
			</a>
		</div>`,
	}

	state, err := (&MastodonExtractor{}).Extract(d).Unpack()
	if err != nil {
		t.Fatalf("Extract returned an error: %v", err)
	}
	if state != sdk.ExtractorSuccess {
		t.Fatalf("Extract state = %v, want %v", state, sdk.ExtractorSuccess)
	}
	if len(d.ExtraDocuments) != 1 {
		t.Fatalf("ExtraDocuments length = %d, want 1", len(d.ExtraDocuments))
	}
	toot := d.ExtraDocuments[0]
	if got, want := toot.URL, d.URL; got != want {
		t.Fatalf("toot URL = %q, want %q", got, want)
	}
	if got, want := toot.Text, "Mastodon 4.4 is now available."; got != want {
		t.Fatalf("toot text = %q, want %q", got, want)
	}
}

func TestExtractSkipsDocumentWhenNoTootsFound(t *testing.T) {
	d := &document.Document{
		URL:  "https://chaos.social/public/local",
		HTML: `<meta content='{"repository":"mastodon/mastodon"}'>`,
	}

	state, err := (&MastodonExtractor{}).Extract(d).Unpack()
	if err != nil {
		t.Fatalf("Extract returned an error: %v", err)
	}
	if state != sdk.ExtractorSuccess {
		t.Fatalf("Extract state = %v, want %v", state, sdk.ExtractorSuccess)
	}
	if !d.SkipIndexing {
		t.Fatal("source document was not marked to skip indexing")
	}
	if len(d.ExtraDocuments) != 0 {
		t.Fatalf("ExtraDocuments length = %d, want 0", len(d.ExtraDocuments))
	}
}
