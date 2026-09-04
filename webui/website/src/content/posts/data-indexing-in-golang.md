---
date: '2026-03-18T14:23:40+00:00'
draft: false
title: 'Data Indexing in Golang'
description: 'A practical guide to fast, content-based document retrieval in Go using Bleve. From simple indexing to custom query languages and performance fine-tuning'
---

If you need fast, content-based retrieval of large amounts of documents, your best option is to use a full-text indexer. Popular solutions like Elasticsearch and Meilisearch are more than capable of getting the job done. But what if you don't want to depend on an external service, or if you need a higher level of control over how your data is stored and searched?

Luckily, Go has an excellent library for exactly this purpose: [Bleve](https://blevesearch.com/). Bleve lets you quickly index any Go struct with sensible defaults and a built-in Google-like query language. Or you can go further and build your own query language and customize every single detail of the indexer.

Bleve is a file-based indexer that can handle millions of records. It supports concurrent reads and writes, hot-swapping of indexes, match highlighting, and much more.

[Hister](https://github.com/asciimoo/hister) is built on top of Bleve and uses a wide range of its features: custom field mappings with language-specific analyzers, a hand-crafted query language with per-field boosting, cursor-based pagination, multi-language index aliases, and fine-grained Scorch tuning. The examples through this post are inspired from our codebase and the knowledge we collected during the development.

## Creating a Simple Indexer

Getting started with Bleve only takes a few lines of code. The two core operations are **indexing** (storing a document so it can be searched later) and **querying** (retrieving ranked documents that match a search expression).

```go
package main

import (
	"fmt"
	"log"

	bleve "github.com/blevesearch/bleve/v2"
)

// Document represents the data we want to index and search.
type Document struct {
	Title string
	URL   string
	Text  string
}

func main() {
	// Create a new index on disk. If one already exists at that path, open it.
	mapping := bleve.NewIndexMapping()
	index, err := bleve.New("example.bleve", mapping)
	if err != nil {
		index, err = bleve.Open("example.bleve")
		if err != nil {
			log.Fatal(err)
		}
	}
	defer index.Close()

	// Index a handful of documents. The first argument is a unique ID;
	// the second is any Go value, Bleve will reflect over its fields.
	docs := map[string]Document{
		"1": {Title: "Go Programming", URL: "https://go.dev", Text: "Go is an open source programming language that makes it easy to build reliable software."},
		"2": {Title: "Bleve Search", URL: "https://blevesearch.com", Text: "Bleve is a full-text search and indexing library for Go."},
		"3": {Title: "Hister - Your own search engine", URL: "https://hister.org/", Text: "Full-text search across your files, browsing history and beyond."},
	}

	for id, doc := range docs {
		if err := index.Index(id, doc); err != nil {
			log.Printf("failed to index %s: %v", id, err)
		}
	}

	// Query the index. NewMatchQuery performs a full-text search across
	// all indexed fields and ranks results by relevance score.
	query := bleve.NewMatchQuery("Hister search engine")
	req := bleve.NewSearchRequest(query)
	req.Fields = []string{"Title", "URL"} // which stored fields to return
	req.Size = 10                         // maximum number of hits

	results, err := index.Search(req)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d result(s):\n", results.Total)
	for _, hit := range results.Hits {
		fmt.Printf("  [%.4f] %s  %s\n", hit.Score, hit.Fields["Title"], hit.Fields["URL"])
	}
}
```

A few things to note:

- **`bleve.New` vs `bleve.Open`** `New` creates a fresh index at the given path; `Open` opens an existing one. The pattern shown above. Try `New`, fall back to `Open` on error is the idiomatic way to handle both the first run and subsequent runs with the same index directory.
- **`bleve.NewIndexMapping()`** Returns a default mapping that works well out of the box: text fields are tokenized, lowercased, and stop-word filtered using the English analyzer. You can replace this with a custom mapping when you need more control (see the _Mappings_ section below).
- **Automatic field discovery** Bleve uses reflection to inspect your struct. Every exported field is automatically tokenized and made searchable with no extra configuration. Unexported fields are silently skipped.
- **Unique document IDs** The string ID you pass to `Index` is how you identify documents for updates and deletes. Calling `Index` with an ID that already exists replaces the previous document in place, making it safe to re-index pages that have changed.
- **`SearchRequest.Fields`** By default Bleve returns only document IDs and relevance scores to keep responses lean. Specify the field names you want returned in `Fields`, or pass `[]string{"*"}` to get every stored field.
- **`hit.Score`** Each result carries a floating-point relevance score computed by Bleve's BM25-based scorer. Higher scores indicate a stronger match. You can influence scores with boost values (covered in the _Querying_ section).

## Mappings

The default mapping works well for a quick start, but real applications usually need more control over how Bleve analyzes and stores each field. A **mapping** tells Bleve what type each field is, which analyzer to use when tokenizing it, whether to store the original value, and whether to include it in the index at all.

Mappings can:

- **Control tokenization** split text into terms using whitespace, language rules, edge n-grams, etc.
- **Filter input data** lowercase terms, strip HTML, apply stop-word lists, or run a stemmer so that "running" and "runs" match the same root token
- **Exclude fields from the index** omit sensitive or irrelevant fields to save disk space and keep the index lean
- **Define custom analyzers** combine any tokenizer with any chain of token filters to get exactly the behavior you need

Here is a concrete example that applies language-based stemming to the `Text` and `Title` fields, and excludes a raw HTML field from being indexed at all:

```go
import (
	"github.com/blevesearch/bleve/v2/analysis/analyzer/en"
	"github.com/blevesearch/bleve/v2/mapping"
)

func buildIndexMapping() *mapping.IndexMappingImpl {
	// English analyzer: tokenizes, lowercases, removes stop words, and stems.
	// "running" and "runs" will both match a search for "run".
	englishField := bleve.NewTextFieldMapping()
	englishField.Analyzer = en.AnalyzerName

	// A keyword analyzer treats the entire field value as a single token
	// useful for exact-match fields like URLs or tags.
	keywordField := bleve.NewTextFieldMapping()
	keywordField.Analyzer = "keyword"

	// Disable indexing for a field we only want to store, not search.
	storedOnlyField := bleve.NewTextFieldMapping()
	storedOnlyField.Index = false

	docMapping := bleve.NewDocumentMapping()
	docMapping.AddFieldMappingsAt("title", englishField)
	docMapping.AddFieldMappingsAt("text", englishField)
	docMapping.AddFieldMappingsAt("url", keywordField)
	docMapping.AddFieldMappingsAt("raw_html", storedOnlyField) // stored but not indexed

	indexMapping := bleve.NewIndexMapping()
	indexMapping.AddDocumentMapping("document", docMapping)
	indexMapping.DefaultAnalyzer = en.AnalyzerName

	return indexMapping
}
```

Pass the result of `buildIndexMapping()` to `bleve.New` or `bleve.NewUsing` when creating the index. Mappings are baked into the index at creation time and cannot be changed afterwards. To apply a new mapping you need to create a fresh index and re-index all documents.

## Querying

Bleve provides a powerful built-in text query processor called `QueryStringQuery`. It supports field filters (`title:golang`), quoted phrases (`"error handling"`), term exclusion (`go -python`), wildcard patterns (`auth*`), and boolean operators (`go AND concurrency`). Its syntax closely mirrors Google's search syntax. You can read more about it [here](https://blevesearch.com/docs/Query-String-Query/).

But where Bleve really shines is in providing composable building blocks for constructing your own domain-specific query language. The [query package](https://pkg.go.dev/github.com/blevesearch/bleve/v2@v2.5.7/search/query) exposes a wide variety of primitives. Match queries, wildcard queries, range queries, boolean combinators, and more that you can wire together however you like.

Here's a simplified example from our app to demonstrate how powerful this can be:

```go
queries := []query.Query{}
negatedQueries := []query.Query{}
for _, keyword := range strings.Fields(queryString) {
    negated := false
    // negate the term if it starts with "-"
    if cut, ok := strings.CutPrefix(keyword, "-"); ok {
        keyword = cut
        negated = true
    }

    // WildcardQuery matches the keyword anywhere inside the URL string.
    // The 10x boost means a URL match raises the document's score
    // significantly compared to a plain text match.
    //
    // The boost number 10 is arbitrary, adjust it to your needs
    urlq := bleve.NewWildcardQuery("*" + keyword + "*")
    urlq.SetField("url")
    urlq.SetBoost(10)

    // MatchQuery tokenizes the keyword with the field's analyzer before
    // matching, so stemming and stop-word removal apply automatically.
    textq := bleve.NewMatchQuery(keyword)
    textq.SetField("text")

    // Title matches are given 50x weight. A keyword found in the title
    // is a very strong signal of relevance.
    //
    // The boost number 50 is arbitrary, adjust it to your needs
    titleq := bleve.NewMatchQuery(keyword)
    titleq.SetField("title")
    titleq.SetBoost(50)

    // DisjunctionQuery is an OR combinator: the document scores as a match
    // if it satisfies *any* of the sub-queries. The final score is taken
    // from whichever sub-query scored highest.
    disjq := bleve.NewDisjunctionQuery(
        urlq,
        textq,
        titleq,
    )

    if negated {
        negatedQueries = append(negatedQueries, disjq)
    } else {
        queries = append(queries, disjq)
    }
}

// BooleanQuery is an AND/OR/NOT combinator at the keyword level:
//   - must    (first arg):  document must satisfy every query in this list
//   - should  (second arg): optional queries that boost score when matched
//   - mustNot (third arg):  document must satisfy none of these queries
//
// The result: every non-negated keyword must appear somewhere in the
// document, while negated keywords disqualify a document entirely.
fullQuery := query.NewBooleanQuery(
    queries,
    nil,
    negatedQueries,
)
```

Each keyword in the input string becomes its own `DisjunctionQuery` that spans all three fields. The `BooleanQuery` then requires that _all_ keyword disjunctions are satisfied, giving us an implicit AND between keywords and a per-field OR within each keyword. Negated keywords (prefixed with `-`) are placed in the `mustNot` list and disqualify any document that matches them.

This structure is easy to extend: you could add date-range filters, weight fields dynamically based on user preferences, or introduce special syntax for field-scoped searches.

Take a look at our [query builder](https://github.com/asciimoo/hister/blob/master/server/indexer/querybuilder/builder.go) for a more complete real-world example.

## Paging

Bleve's [SearchRequest](https://pkg.go.dev/github.com/blevesearch/bleve/v2#SearchRequest) controls both the page size (`Size`) and the starting offset of results. A natural first instinct is to use the `From` field, set it to `0` for the first page, `20` for the second, and so on. This works, but it has a serious problem: Bleve must score and sort _all_ matching documents up to `From + Size` on every request, making deep pages increasingly expensive in both memory and CPU. Worse, if new documents are indexed between two page requests, the offset shifts and users see duplicate or missing results.

The correct approach is to use cursor-based pagination via [SearchAfter](https://pkg.go.dev/github.com/blevesearch/bleve/v2#SearchRequest.SetSearchAfter) and [SearchBefore](https://pkg.go.dev/github.com/blevesearch/bleve/v2#SearchRequest.SetSearchBefore). These functions resume the result stream from a known position rather than re-scanning from the beginning, which is both accurate and efficient. We learned to prefer them the [hard way](https://github.com/asciimoo/hister/issues/173).

```go
const pageSize = 20

// First page, no cursor needed.
req := bleve.NewSearchRequest(myQuery)
req.Size = pageSize
req.SortBy([]string{"_score", "_id"}) // stable sort is required for cursors

results, _ := index.Search(req)

// Subsequent pages, pass the sort key of the last hit as the cursor.
if len(results.Hits) == pageSize {
    lastHit := results.Hits[len(results.Hits)-1]
    cursor := lastHit.Sort // []string, one element per sort field

    nextReq := bleve.NewSearchRequest(myQuery)
    nextReq.Size = pageSize
    nextReq.SortBy([]string{"_score", "_id"})
    nextReq.SetSearchAfter(cursor)

    nextResults, _ := index.Search(nextReq)
    // ...
}
```

A few things to keep in mind:

- **Stable sort is required.** `SearchAfter` uses the sort key of the last result as its cursor. If the sort key is changing the cursor become invalid.
- **`Sort` is always a `[]string`.** Even when sorting by a numeric field, Bleve serializes the sort key as a string. Read the cursor from `hit.Sort[0]` (or whichever index corresponds to your primary sort field) and pass it directly to `SetSearchAfter`.
- **`SearchBefore` works the same way** but moves in the opposite direction, which is useful for implementing a "previous page" button.

## Handling Multiple Indexes

Bleve can transparently manage multiple indexes at the same time through [IndexAlias](https://pkg.go.dev/github.com/blevesearch/bleve/v2#IndexAlias). An alias is a virtual index that fans a query out to several real indexes and merges their results back into a single ranked list.

This is particularly useful when you want to maintain separate indexes for different languages. Each language gets its own index with a tailored analyzer (English stemming, French stop-words, custom tokenization, etc.), but a single alias lets you search all of them at once:

```go
enIndex, _ := bleve.Open("index_en.bleve")
frIndex, _ := bleve.Open("index_fr.bleve")
deIndex, _ := bleve.Open("index_de.bleve")

// Combine all language indexes behind a single alias.
alias := bleve.NewIndexAlias(enIndex, frIndex, deIndex)

// Query the alias exactly as you would a regular index.
req := bleve.NewSearchRequest(bleve.NewMatchQuery("Hister search engine"))
req.Size = 20
results, _ := alias.Search(req)
```

Aliases also make hot-swapping painless. When you need to rebuild an index (for example, to apply a new mapping), you can build the new index in the background, then atomically swap it into the alias with `alias.Swap(newIndexes, oldIndexes)`. In-flight queries complete against the old index while new queries immediately use the new one, with zero downtime.

## Fine-tuning

Bleve's performance knobs are not prominently documented, but they make a real difference under load. Configuration is passed as a `map[string]any` to [NewUsing](https://pkg.go.dev/github.com/blevesearch/bleve/v2#NewUsing) or [OpenUsing](https://pkg.go.dev/github.com/blevesearch/bleve/v2#OpenUsing) instead of the regular `New`/`Open` functions.

```go
config := map[string]any{
	// How long the BoltDB storage layer will wait for a write lock
	// before returning an error. Increase this if you see timeout
	// errors under concurrent write load.
	"bolt_timeout": "2s",

	"scorchPersisterOptions": map[string]any{
		// Number of goroutines that flush in-memory segments to disk
		// in parallel. More workers help throughput on multi-core machines
		// at the cost of higher memory usage during flushing.
		"NumPersisterWorkers": 4,

		// Maximum bytes each persister worker holds in memory before
		// flushing. Larger values reduce I/O by writing bigger segments,
		// but increase peak memory consumption.
		"MaxSizeInMemoryMergePerWorker": 80 * 1024 * 1024, // 80 MB

		// The persister pauses merging when the number of on-disk segment
		// files is below this threshold, reducing unnecessary write
		// amplification when the index is small or lightly loaded.
		"PersisterNapUnderNumFiles": 100,
	},

	"scorchMergePlanOptions": map[string]any{
		// Segments smaller than this size are candidates for merging.
		// Raising this value reduces the total number of segments (and
		// therefore read latency) at the cost of more merge I/O.
		"FloorSegmentFileSize": 20 * 1024 * 1024, // 20 MB
	},
}

index, err := bleve.OpenUsing("my.bleve", config)
```

These settings live in the Scorch storage backend, which is Bleve's default. Consult the [persister source](https://github.com/blevesearch/bleve/blob/master/index/scorch/persister.go#L67) for the full list of available options and their default values.

## Conclusion

Bleve is one of Go's hidden gems that deserves more attention. It lets you add full-text search to your application, without complex infrastructure. The default configuration gets you up and running in minutes, while the custom mapping system, composable query primitives, performance, debugging options and deep custimzation provides a great toolset to solve specific problems optimally.

The official documentation has gaps, but the GitHub issues and real-life open-source projects fill them in well. Check out our [indexer package](https://github.com/asciimoo/hister/tree/master/server/indexer) to see all of the above concepts working together in a production codebase.

Happy indexing.
