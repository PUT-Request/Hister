---
date: '2026-04-27T11:00:00+00:00'
draft: false
title: "Building a Personal Knowledge Base with Hister's Crawler"
description: "Hister's crawler turns any website into a searchable part of your personal index. Here is how to build a structured knowledge base from documentation, wikis, and reference sites."
---

A personal knowledge base is only as good as how quickly you can find something in it. Most solutions optimized for capture are terrible for retrieval. You dump notes and links into them, and six months later you cannot find what you are looking for unless you remember exactly where you put it.

Hister approaches this differently. Everything in your index is full-text searchable, instantly, with no folder hierarchy to navigate and no tag system to maintain.

Hister's crawler extends this to content you have never personally visited.

## The Core Idea

The browser extension indexes pages as you browse. The crawler indexes pages you tell it about, including entire documentation trees, wikis, and reference sites, without you having to visit each page manually.

Combine the two and your knowledge base covers everything: what you have read, what you have written locally, and what you know you will need to look up.

## Crawling Documentation

The most direct use case is pre-indexing documentation for libraries and tools you use regularly:

```bash
hister index --recursive \
  --allowed-domain=docs.example.com \
  --max-depth=4 \
  https://docs.example.com/
```

This crawls the documentation site up to four levels deep, adds every page to your index, and gives you instant full-text search over the entire documentation set.

For Go packages:

```bash
hister index --recursive \
  --allowed-pattern="https://pkg.go.dev/github.com/yourlibrary/.*" \
  https://pkg.go.dev/github.com/yourlibrary/package
```

For Python packages:

```bash
hister index --recursive \
  --allowed-pattern="https://docs.python.org/3/library/.*" \
  https://docs.python.org/3/library/
```

## Using Labels to Organize

Hister supports a label field that you can attach to documents at index time. Labels are searchable just like any other field, so you can filter by topic, project, or whatever taxonomy makes sense for you.

The easiest way to label a crawl job is with the `--label` flag:

```bash
hister index --recursive \
  --label=projectx \
  --allowed-domain=docs.example.com \
  --max-depth=4 \
  https://docs.example.com/
```

Every page crawled in that job will have the label "projectx" attached to it.

You can also set a label in the browser extension (applied to all pages you visit while it is set), or specify it command line for the `index` command.

Then search by label:

```
label:projectx authentication
```

This returns only documents tagged with "projectx" that match "authentication". Useful when you have documentation for multiple projects in the same index and want to scope a search.

## Crawl and Label Workflow

A practical workflow for onboarding a new project:

1. Crawl the project's documentation site with `--label=projectname`.
2. Open the pages you know are important so the extension captures them with your reading context attached (the browser extension uses its own global label setting).

After this, searching with `label:projectname` gives you a scoped knowledge base for that project. Searching without a label gives you everything.

## Keeping the Index Fresh

Documentation sites update. Run the crawler again to pick up changes:

```bash
hister index --recursive \
  --label=projectx \
  --allowed-domain=docs.example.com \
  https://docs.example.com/
```

Hister updates existing entries in the index when a URL is resubmitted. You do not end up with duplicates.

For frequently updated sites, you can automate this with a cron job or a scheduled task.

## Pairing with Semantic Search

If you have semantic search enabled, your crawled documentation becomes part of the meaning-aware index automatically. You can then find relevant pages by describing what you need rather than quoting exact API names.

Searching "how to handle connection timeouts" in a crawled networking library will surface the relevant pages even if they use terms like "socket deadline" and "read timeout" rather than "connection timeout".

## What This Replaces

This workflow replaces a combination of bookmarks, browser history, local markdown notes, and the habit of re-searching the same documentation pages. Everything lives in one place, searchable by full text and label.

The full configuration reference is at [hister.org/docs/configuration](/docs/configuration). The crawler options are documented under the `crawler` section.
