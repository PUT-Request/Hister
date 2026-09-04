// SPDX-License-Identifier: AGPL-3.0-or-later

package vectorstore

import "testing"

func TestSearchCandidateLimit(t *testing.T) {
	if got := searchCandidateLimit(10); got != 40 {
		t.Fatalf("unexpected candidate limit: got %d, want 40", got)
	}
	if got := searchCandidateLimit(0); got != 0 {
		t.Fatalf("unexpected disabled candidate limit: got %d, want 0", got)
	}
}

func TestDiversifySearchResults(t *testing.T) {
	results := []Result{
		{DocID: "long", ChunkIdx: 0, Similarity: 0.99},
		{DocID: "long", ChunkIdx: 1, Similarity: 0.98},
		{DocID: "long", ChunkIdx: 2, Similarity: 0.97},
		{DocID: "second", ChunkIdx: 0, Similarity: 0.96},
		{DocID: "long", ChunkIdx: 3, Similarity: 0.95},
		{DocID: "third", ChunkIdx: 0, Similarity: 0.94},
		{DocID: "second", ChunkIdx: 1, Similarity: 0.93},
		{DocID: "fourth", ChunkIdx: 0, Similarity: 0.92},
	}

	got := diversifySearchResults(results, 3, 2)
	want := []Result{
		{DocID: "long", ChunkIdx: 0, Similarity: 0.99},
		{DocID: "long", ChunkIdx: 1, Similarity: 0.98},
		{DocID: "second", ChunkIdx: 0, Similarity: 0.96},
		{DocID: "third", ChunkIdx: 0, Similarity: 0.94},
		{DocID: "second", ChunkIdx: 1, Similarity: 0.93},
	}

	if len(got) != len(want) {
		t.Fatalf("unexpected result count: got %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unexpected result at %d: got %#v, want %#v", i, got[i], want[i])
		}
	}
}

func TestDiversifySearchResultsRejectsInvalidLimits(t *testing.T) {
	results := []Result{{DocID: "doc", Similarity: 1}}
	if got := diversifySearchResults(results, 0, 2); got != nil {
		t.Fatalf("expected nil results for zero document limit, got %#v", got)
	}
	if got := diversifySearchResults(results, 1, 0); got != nil {
		t.Fatalf("expected nil results for zero chunk limit, got %#v", got)
	}
}
