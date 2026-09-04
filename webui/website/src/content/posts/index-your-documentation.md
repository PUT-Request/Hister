---
date: '2026-06-04T16:00:00+02:00'
draft: false
title: 'Index Your Entire Documentation Stack'
description: 'Go documentation, MDN, your project docs, Stack Overflow answers, GitHub issues. A practical guide to turning Hister into a fast offline reference for everything you work with.'
---

IT enthusiasts spend a significant portion of their time reading documentation. Language references, library APIs, framework guides, Stack Overflow answers, internal wikis, GitHub issues. Most of this is read once and either remembered imperfectly or looked up again the next time it is needed.

Hister can index all of it. Once it is indexed, a query returns results from all your sources simultaneously, with no distinction between what came from an official API reference, a blog post, or your own notes.

> We've started publishing [public datasets](/datasets) that can be imported to local Hister instances. Check them out and [suggest](https://github.com/asciimoo/hister) new datasets.

## Pre-indexing Documentation You Use Every Day

The fastest wins come from pre-indexing documentation for tools you use constantly. You do not have to wait until you visit a page to have it in your index.

For Go:

```bash
hister index --recursive \
  --allowed-domain=pkg.go.dev \
  --allowed-pattern="https://pkg.go.dev/github.com/yourdep/.*" \
  --max-depth=3 \
  https://pkg.go.dev/github.com/yourdep/package
```

For MDN (JavaScript, CSS, Web APIs):

```bash
hister index --recursive \
  --allowed-domain=developer.mozilla.org \
  --allowed-pattern="https://developer.mozilla.org/en-US/docs/.*" \
  --max-depth=3 \
  https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/
```

For Python:

```bash
hister index --recursive \
  --allowed-domain=docs.python.org \
  --allowed-pattern="https://docs.python.org/3/library/.*" \
  https://docs.python.org/3/library/
```

The `--allowed-domain` flag restricts the crawl so it does not follow links off the documentation site. The `--allowed-pattern` flag narrows it further to a specific section of the site. The `--max-depth` flag prevents the crawler from going arbitrarily deep into nested pages.

After running these commands you have full-text search over those documentation sets with no network dependency. Searching for `asyncio gather` returns the relevant MDN and Python pages instantly from your local index.

## Indexing Your Own Project Documentation

If your project has documentation in a repository or deployed as a static site, indexing it gives you searchable access to your own docs alongside everything else:

```bash
hister index --recursive \
  --allowed-domain=docs.yourcompany.com \
  https://docs.yourcompany.com/
```

For local documentation that is not served as a web page, Hister can index Markdown and text files directly. Add the directory to your indexer configuration:

```yaml
indexer:
  directories:
    - path: ~/projects/yourproject/docs
      filetypes: [md, txt, rst]
```

Hister watches the directory for changes and re-indexes modified files automatically. Your documentation search stays current as you write.

## GitHub and Stack Overflow

The browser extension handles these automatically as you browse, but you can also crawl targeted sections. For a library you are adopting, the issue tracker is often as useful as the documentation:

```bash
hister index --recursive \
  --allowed-pattern="https://github.com/yourorg/yourrepo/issues/.*" \
  --max-links=500 \
  https://github.com/yourorg/yourrepo/issues
```

The `--max-links` flag caps the crawl so you do not accidentally index an entire repository history. For active projects, the 500 most-linked issues typically cover the most discussed problems.

For Stack Overflow, the extension is more practical than crawling since you can index answers as you encounter them and they are immediately searchable.

## Using the Chromedp Backend for JavaScript-Heavy Sites

Some documentation sites are built as single-page applications and do not return useful content with a plain HTTP request. The Hister crawler has a `chromedp` backend that runs a headless Chrome instance and waits for JavaScript to render before capturing the page:

```yaml
crawler:
  backend: chromedp
  backend_options:
    exec_path: /usr/bin/chromium
```

With this configuration, `hister index --recursive` uses Chromedp for every page fetch. This is slower than the HTTP backend but handles sites that require JavaScript rendering. Switch back to the default HTTP backend for plain HTML documentation sites.

## Labeling Documentation by Project or Type

When you index documentation from many sources, finding everything related to a specific project or technology can require creative query construction. Labels give you a direct solution: attach a short tag to documents at index time and retrieve them later with a single `label:` filter.

Labels can be set in the extension to automatically label each visited website. It can be useful for tagging research sessions - don't forget to clear the label from the extension when you end your research.

Attaching a label is also supported when running the `index` command:

```bash
hister index --recursive \
  --allowed-domain=pkg.go.dev \
  --label=golang-stdlib \
  https://pkg.go.dev/context
```

Every page captured by this crawl is stored with the label `golang-stdlib`. Repeat the pattern for each source you want to group:

```bash
hister index --recursive \
  --allowed-domain=docs.yourcompany.com \
  --label=internal-api \
  https://docs.yourcompany.com/

hister index --recursive \
  --allowed-pattern="https://github.com/yourorg/yourrepo/issues/.*" \
  --label=projectx-issues \
  https://github.com/yourorg/yourrepo/issues
```

Retrieve any group instantly:

```
label:golang-stdlib context deadline
label:internal-api authentication
label:projectx-issues panic
```

Combine labels with domain or content filters to narrow further:

```
label:internal-api domain:docs.yourcompany.com timeout
label:golang-stdlib Marshal*
```

You can also assign or update a label after a document is already indexed. Click the three-dot menu next to any search result, type a label into the "Label" input, and click Save. This is useful for tagging pages you discovered through regular browsing with the browser extension.

Labeled documents display a small tag badge next to their URL in search results, so you can see at a glance which group a result belongs to without opening it.

## Query Syntax for Documentation Searches

Once your documentation is indexed, a few query patterns are particularly useful for developer workflows.

Search within a specific domain:

```
domain:pkg.go.dev context deadline
```

Search for a function or method name:

```
title:"http.NewRequest" example
```

Combine domain and content filters:

```
domain:developer.mozilla.org "fetch api" cors
```

Exclude a domain to search everything except one source:

```
authentication middleware -domain:stackoverflow.com
```

Use wildcards for partial function names:

```
domain:pkg.go.dev Marshal*
```

The [query language reference](/docs/query-language) covers all available operators.

## Priority Results for Your Most-Used Pages

Some pages you look up constantly: the error handling section of a language spec, the configuration reference for a framework, the specific API you call in every project. Set these as priority results so they always appear first for the relevant queries.

Click the three-dot menu next to any search result and choose "Set as priority result". Enter the query string you use to find it. The next time you run that search, the page appears at the top of the results list before the regular ranked results.

## Keeping Documentation Current

Documentation changes. When a new version of a library is released its API reference changes. The `hister reindex` command re-processes all documents in your index through the current extractor chain with the current configuration, but it does not re-fetch pages from the internet.

For documentation you want to keep current, set up a periodic crawl:

```bash
# Add to cron or a systemd timer
hister index --recursive \
  --allowed-domain=pkg.go.dev \
  https://pkg.go.dev/yourfrequentlyupdatedlib
```

Hister performs deduplication at index time: if a page's content has not changed since the last crawl, the existing index entry is left in place and no new file is written to the HTML store. Recrawling a site that has not changed is fast and cheap.

## The Result

A quick setup gives you a local search engine that covers every documentation source you work with regularly, returns results in milliseconds, works without an internet connection, and does not require switching between browser tabs and documentation sites. Searching `context.WithTimeout example` returns results from Go's pkg.go.dev, from blog posts you have read about context usage, from Stack Overflow answers you visited, and from your own project's notes, all in a single ranked list.

The [quickstart guide](/docs/quickstart) gets you running. The [configuration docs](/docs/configuration) cover the advanced setup posibilities in detail.
