---
date: '2026-04-12T10:00:00+00:00'
draft: false
title: 'Extractors: An Ecosystem for Better Search'
description: 'How Hister extractors turn raw HTML into rich, searchable content and why we need more of them'
---

Every page you add to Hister arrives as raw HTML. Raw HTML is noisy. A Stack
Overflow question, a Go package reference, a GitHub issue, and a news article
all contain the same structural clutter: navigation bars, cookie banners, sidebars,
footers, and advertising wrappers. A generic text extractor discards all of that,
but it also discards the structure that makes each content type useful.

Extractors are the layer between raw HTML and the indexed, searchable content you
actually want to find.

## Extractors: The Secret Ingredients of Hister

Hister indexes your data and makes it searchable. The quality of what
you can find depends almost entirely on the quality of the text that was indexed.
A mediocre extractor means mediocre search results. A great extractor means that
when you search for something you read six months ago, you actually find it.

More extractors, covering more of the sites people actually use, make Hister more
useful for everyone. This is one of the highest leverage contributions you can
make to the project, and it does not require touching the indexer, the crawler, or
any of the more complex parts of the codebase.

## How You Can Contribute

The extractor ecosystem is young. There are two specialist extractors in the
codebase today. The list of sites people use every day is considerably longer.

We have an open issue at [github.com/asciimoo/hister/issues/305](https://github.com/asciimoo/hister/issues/305)
collecting ideas and tracking who is working on what. The current shortlist
includes:

- A video information extractor using `yt-dlp`
- A GitHub project and repository extractor
- A general Stack Exchange question and answer extractor (which would cover
  dozens of communities beyond Stack Overflow)
- A Reddit post extractor

None of these require deep knowledge of the Hister internals. The interface is
small, the existing extractors are short enough to read in a few minutes, and
the full reference documentation is at [hister.org/docs/extractors](https://hister.org/docs/extractors).

If you have a site you visit constantly and wish Hister understood better, that
is the right motivation to write an extractor for it. Comment on the issue to
claim one, or open a new issue if you have a site in mind that is not on the list.
Pull requests are very welcome.

## What Extractors Do

When Hister receives a document, it runs the content through a chain of
extractors. Each extractor has two jobs.

The first job is **indexing**: turning the raw HTML into clean, well structured
plain text so that the full text search index contains only the content that
matters. A generic parser might include sidebar text, repeated navigation labels,
or boilerplate footers in the indexed text. A domain specific extractor can strip
all of that and surface only the real content.

The second job is **preview**: when you click on a search result to read it
inside Hister, the extractor produces the rendered output you see in the preview
panel. This is where a specialist extractor really shines. It can apply a custom
layout, format code blocks, structure questions and answers, and link related
sections in a way that is far more useful than a plain text dump.

## The Chain

Extractors are tried in registration order. Each one examines the document and
returns one of three signals:

- **Stop** means success: the extractor handled the document and the chain halts.
- **Continue** means the extractor was inconclusive: the next matching extractor
  in the chain gets a turn.
- **Abort** means a fatal error occurred: the chain stops immediately and the
  error is returned to the caller.

Today the chain looks like this, from most specific to most general:

1. **Stackoverflow** matches `stackoverflow.com/questions/*` and renders the
   question body together with all answers, formatted as a clean sequence of
   sections without the voting widgets, copy buttons, or profile cards that
   surround them on the live site.

2. **GoDoc** matches `pkg.go.dev/*` and extracts the `Documentation` content
   section of a Go package page, rewriting every relative link to an absolute URL
   so navigation within the preview panel works correctly.

3. **Readability** is a general purpose article extractor. It uses a readability
   algorithm to identify the main content block of any article or blog post and
   discard the surrounding page furniture.

4. **Default** is the final safety net. It walks the full HTML token stream and
   collects all visible text, with no heuristics at all.

The first two extractors are domain specific. The last two are generic fallbacks.
This means that right now, the vast majority of pages you index are handled by the
Readability or Default extractor, which know nothing about the structure of the
content they are processing.

Every major site you visit regularly is an opportunity for a specialist extractor
to do a much better job.

## The Interface

Writing a new extractor is straightforward. You implement a single Go interface:

```go
type Extractor interface {
    Name() string
    Description() string
    Capabilities() sdk.Capabilities
    Match(*sdk.Document) bool
    Extract(*sdk.Document) sdk.ExtractResult
    Preview(*sdk.Document) sdk.PreviewResult
    GetConfig() *sdk.Config
    SetConfig(*sdk.Config) error
}
```

`Match` is called for every document before anything else. If it returns false,
the extractor is skipped entirely for that document. This is the right place to
check whether the URL belongs to a specific domain or matches a particular path
pattern.

`Capabilities` declares whether an implementation enriches metadata, extracts
searchable content, renders previews, or combines those roles.

`Extract` populates or modifies the `Document` object before it goes into the
search index. Return `ExtractFallback` when another content extractor should
try the document. Custom key value data can be added to the
`Document.Metadata` field for later use in previews.

`Preview` returns the HTML (or plain text) that Hister renders in the preview
panel. You can also return a custom `Template` name in the `PreviewResponse` to
apply a Svelte template specifically designed for this content type.

Result constructors make control flow explicit. Extraction uses `Extracted`,
`ExtractFallback`, and `AbortExtraction`. Preview rendering uses `Previewed`,
`PreviewFallback`, and `AbortPreview`.

Here is what a minimal extractor looks like in practice:

```go
// SPDX-License-Identifier: AGPL-3.0-or-later

package mySite

import (
    "fmt"
    "strings"

    "github.com/PuerkitoBio/goquery"
    "github.com/asciimoo/hister/server/extractor/sdk"
    "github.com/asciimoo/hister/server/sanitizer"
)

const matchPrefix = "https://example.com/posts/"

type MySiteExtractor struct {
    cfg *sdk.Config
}

func (e *MySiteExtractor) Name() string { return "MySite" }

func (e *MySiteExtractor) Description() string {
    return "Renders example.com posts"
}

func (e *MySiteExtractor) Capabilities() sdk.Capabilities {
    return sdk.Capabilities{Preview: true}
}

func (e *MySiteExtractor) Match(d *sdk.Document) bool {
    return strings.HasPrefix(d.URL, matchPrefix)
}

func (e *MySiteExtractor) Extract(d *sdk.Document) sdk.ExtractResult {
    // Let the generic Readability extractor handle indexing.
    return sdk.ExtractFallback(nil)
}

func (e *MySiteExtractor) Preview(d *sdk.Document) sdk.PreviewResult {
    doc, err := goquery.NewDocumentFromReader(strings.NewReader(d.HTML))
    if err != nil {
        return sdk.PreviewFallback(err)
    }
    content, err := doc.Find("article.post-body").Html()
    if err != nil {
        return sdk.PreviewFallback(err)
    }
    return sdk.Previewed(sdk.PreviewResponse{
        Content: sanitizer.SanitizeHTML(content),
    })
}

func (e *MySiteExtractor) GetConfig() *sdk.Config {
    if e.cfg == nil {
        return &sdk.Config{Enable: true, Options: map[string]any{}}
    }
    return e.cfg
}

func (e *MySiteExtractor) SetConfig(c *sdk.Config) error {
    for k := range c.Options {
        return fmt.Errorf("unknown option %q", k)
    }
    e.cfg = c
    return nil
}
```

Once you have the struct, place it in a new subdirectory under
`server/extractor/extractors/` and add an instance to `DefaultExtractors` in
`server/extractor/registry.go`, before the generic fallbacks. That is all it
takes.

## A Few Things Worth Knowing

**Work with what you have.** The `Document` struct already contains the raw HTML,
the URL, the title, and the detected language. There is rarely a reason to make
additional HTTP requests inside an extractor. Extra requests add latency, can fail
silently, and expose the user's browsing activity to external servers. In a
privacy focused tool like Hister, that matters.

**Strip third party content.** Embedded iframes, remote images, and tracking
pixels cause the browser to contact external servers every time a preview is
opened. Sanitize the HTML with `sanitizer.SanitizeHTML` before returning it, and
if you need to surface a video or embed, prefer a placeholder that the user clicks
to load on demand.

**Use a custom template when the content has structure.** A question and answer
thread, a recipe, an API reference, or a commit message all benefit from a layout
that matches the shape of the content. Return a non empty `Template` name in
`PreviewResponse` and write a matching Svelte template in the web UI.
