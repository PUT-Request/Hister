package querybuilder

import (
	"testing"

	"github.com/blevesearch/bleve/v2/search/query"
)

func TestParseSearch(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		text    string
		sort    string
		hasSort bool
	}{
		{
			name:    "directive among terms",
			input:   "golang sort:date documentation",
			text:    "golang documentation",
			sort:    "date",
			hasSort: true,
		},
		{
			name:    "directive only",
			input:   "sort:visits",
			text:    "*",
			sort:    "visits",
			hasSort: true,
		},
		{
			name:    "reverse directive",
			input:   "golang sort:-date",
			text:    "golang",
			sort:    "-date",
			hasSort: true,
		},
		{
			name:    "last directive wins",
			input:   "sort:date golang sort:domain",
			text:    "golang",
			sort:    "domain",
			hasSort: true,
		},
		{
			name:    "relevance clears legacy sorting",
			input:   "golang sort:relevance",
			text:    "golang",
			hasSort: true,
		},
		{
			name:  "invalid directive remains text",
			input: "golang sort:unknown",
			text:  "golang sort:unknown",
		},
		{
			name:  "quoted directive remains text",
			input: `golang "sort:date reference"`,
			text:  `golang "sort:date reference"`,
		},
		{
			name:    "quoted text is preserved when directive is removed",
			input:   `"golang documentation" sort:date`,
			text:    `"golang documentation"`,
			sort:    "date",
			hasSort: true,
		},
		{
			name:  "field value remains text",
			input: `text:"sort:date"`,
			text:  `text:"sort:date"`,
		},
		{
			name:  "negated directive remains text",
			input: "golang -sort:date",
			text:  "golang -sort:date",
		},
		{
			name:    "unicode text preserves source ranges",
			input:   "سلام sort:date világ",
			text:    "سلام világ",
			sort:    "date",
			hasSort: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := ParseSearch(test.input)
			if got.Text != test.text || got.Sort != test.sort || got.HasSort != test.hasSort {
				t.Fatalf("ParseSearch(%q) = %#v, want text %q, sort %q, hasSort %v", test.input, got, test.text, test.sort, test.hasSort)
			}
		})
	}
}

func TestParseReverseSortDirectives(t *testing.T) {
	for _, sortMode := range []string{"-relevance", "-date", "-domain", "-visits"} {
		expression := ParseSearch("sort:" + sortMode)
		if expression.Text != "*" || expression.Sort != sortMode || !expression.HasSort {
			t.Errorf("ParseSearch for %q = %#v", sortMode, expression)
		}
	}
}

func TestBuildDoesNotTreatSortAsASearchDirective(t *testing.T) {
	if _, ok := Build("sort:date").(*query.MatchAllQuery); ok {
		t.Fatal("Build treated sort:date as a match all query")
	}
}
