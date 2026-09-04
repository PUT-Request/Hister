---
date: '2026-06-30T14:00:00+02:00'
draft: false
title: 'Give Your AI Assistant a Private Memory'
description: 'Use Hister MCP to let AI assistants search pages you have read, retrieve stored previews, and answer from your own index.'
---

AI assistants are very good at explaining things, but they usually have a weak
memory of your own context.

They do not know which documentation page you read last week. They do not know
which article convinced you to try a library. They do not know which migration
guide you followed, which GitHub issue had the workaround, or which security
writeup you already trusted.

That context often lives in your browser history, bookmarks, local notes, and
half-remembered search queries.

Hister turns that context into a private search index. With MCP support, it can
also make that index available to AI assistants.

## A Quick Introduction to Hister

Hister is a personal search engine for pages and files you care about.

The browser extension automatically indexes pages as you browse. The command line
tool can import browser history, crawl documentation sites, index URLs manually,
and add local files. The web interface lets you search everything with full-text
search, field filters, labels, facets, priority results, and optional semantic
search.

The important part is where the data lives.

Your browsing history, indexed pages, search queries, and local files stay in
your Hister instance. For most people, that means they stay on the same machine
or on a server they control.

This also means you do not need to set up a separate assistant skill for every
site you use. You do not need to give your assistant individual API access to
GitHub, documentation sites, issue trackers, forums, wikis, and other sites
that require authentication. If the useful page is already in Hister, the
assistant can search Hister instead.

> **Instead of wiring every private service into your assistant, Hister gives it
> one search interface over material you already indexed.**

That makes Hister useful as a daily search tool by itself. MCP adds another
layer: an assistant can ask Hister for relevant pages instead of guessing from
general model knowledge or fetching random web results.

## Practical Hister MCP Workflows

Here are a few ways Hister MCP can be useful in everyday work. The common
pattern is simple: the assistant asks Hister first, then works from pages and
files that are already part of your own index.

### Find the Article You Remember Vaguely

You know the situation: you read a good explanation of a bug, a design pattern,
or a deployment problem, but you cannot remember the title or the site.

Without Hister, you ask an assistant a broad question and get a broad answer.
It might be helpful, but it is not grounded in the source you remember.

With Hister MCP configured, you can ask:

```text
Search my Hister index for the article I read about PostgreSQL migration
locking and summarize the most relevant result.
```

The assistant can search Hister for `PostgreSQL migration locking`, inspect the
stored result, and answer based on the pages in your own index.

If semantic search is enabled on your Hister server, the assistant can also ask
for semantic matching. That helps when you remember the idea but not the exact
words. If semantic search is not available on the server, Hister falls back to
normal keyword search.

### Explain Code with Documentation You Already Indexed

AI coding assistants often answer from broad training data. That is useful, but
it can be wrong for the exact version of a library or framework you use.

Hister gives the assistant a more specific source of context.

You can crawl the documentation for a project:

```bash
hister index --recursive \
  --allowed-domain=docs.example.com \
  --max-depth=4 \
  https://docs.example.com/
```

You can also start from Hister's prepared documentation datasets instead of
crawling everything yourself. The [datasets page](/datasets) includes imports
for common reference material such as language and platform documentation, so
your assistant can search those docs through Hister immediately after import.

Then, inside your AI assistant, ask:

```text
Using my Hister index, find the official docs for configuring connection
timeouts in this library. Then explain which option applies to this code.
```

The assistant can search your indexed docs, retrieve previews, and use those
pages when explaining the code.

This is especially useful for internal documentation, private wikis, older API
versions, or documentation that is hard to find through a normal web search.

### Turn Research into a Source-Backed Brief

When you research a topic over several days, the useful material usually ends
up scattered across tabs and history.

With Hister, browsing becomes capture. With MCP, the assistant can turn that
captured material into a brief.

Example prompt:

```text
Search my Hister history for pages about supply chain attacks from the last
month. Group the findings by incident, list the affected projects, and cite the
pages you used.
```

The assistant can use date filters in the Hister search tool, then retrieve
stored previews for the most relevant pages. The result is not just a generic
summary of supply chain attacks. It is a summary of the material you actually
read or indexed.

This works well for security research, academic literature review, market
research, project planning, and any workflow where source traceability matters.

### Continue from Recent History

Sometimes you do not even know what to search for. You only know that you were
reading something relevant earlier.

The `get_history` tool lets an assistant inspect recently indexed pages or
recently opened Hister results.

You can ask:

```text
Look at my recently indexed Hister pages and tell me which ones are relevant to
the caching problem we are debugging.
```

The assistant can start from recent history, identify likely sources, then use
search and preview retrieval to dig deeper.

This is different from giving the assistant access to the whole open web. It is
looking first at the trail of pages you created while working.

### Create a Worklog from Your Browsing History

Worklogs are useful, but writing them manually is easy to forget. Hister already
knows which pages you indexed, searched, and opened while working, so an
assistant can use that trail to draft a daily or weekly log.

You can ask:

```text
Look at my recently indexed Hister pages from today. Group them by project,
summarize what I worked on, and create a short worklog with source URLs.
```

This works best when the assistant treats the result as a draft. You still
decide what belongs in the final worklog, but the boring part of reconstructing
the day from browser history is done for you.

For more details on keeping worklogging lightweight, see
[kvch's post, Worklogging for the Lazy Dev](https://kvch.me/worklogging-for-the-lazy-dev/).

## Privacy and Control

Hister MCP does not require sending your whole history to an AI provider.

The assistant asks for specific searches. Hister returns matching results and
previews. You can protect the endpoint with an access token or user token. You
can also control what enters the index with skip rules, manual indexing,
automatic indexing settings, labels, and domain rules.

Warning: if you connect Hister MCP to an assistant backed by an external AI
provider, the search results and previews returned by Hister may be sent to
that provider as part of the conversation context. That can expose private
browsing data, indexed documents, page titles, URLs, snippets, and stored page
content.

Indexed pages are also untrusted input. A malicious page can contain prompt
injection instructions aimed at the assistant. Hister marks source fields as
untrusted structured content, returns HTML only when requested, strips invisible
control characters, and labels the exact untrusted data boundary. The consuming
assistant must still treat every returned field as data, sanitize HTML before
rendering it, and request confirmation before using another tool to act on it.

Use a local model when you want to avoid sending that material to an external
AI provider. This keeps both the assistant and Hister on infrastructure you
control. You can also add skip rules to keep private or sensitive pages out of
Hister in the first place, so they cannot be returned through MCP later.

This matters because a personal search index is sensitive. It reflects what you
read, what you work on, what you investigate, and what you forget.

Hister is designed so you can decide what gets indexed, where it is stored, and
which clients can read it.

## Getting Started

First, follow the [Obtaining Hister guide](/docs/installing), [run Hister](/docs/quickstart),
and make sure you have some content indexed. The browser extension is the
easiest way to build the index during normal browsing. You can also import
existing browser history or crawl documentation sites.

Then configure your MCP client to use:

```text
http://127.0.0.1:4433/mcp
```

If authentication is enabled, pass your Hister access token as a bearer token.

The full setup guide is available in the [MCP integration docs](/docs/mcp).

### What MCP Adds

[MCP](https://modelcontextprotocol.io) is a protocol that lets AI tools connect
to external data sources and tools. Hister exposes an MCP endpoint at `/mcp`.

Through that endpoint, an assistant can:

1. Search your indexed pages and files with the `search` tool.
2. Retrieve the stored preview of a specific page with `get_preview`.
3. Inspect recently indexed or recently opened Hister history with `get_history`.

This changes the assistant from a general answer machine into something closer
to a research partner that can work with the material you have already read.

#### Better Answers Without Refetching Pages

Many assistant workflows rely on fetching URLs again. That can fail because of
login walls, rate limits, removed pages, bot protection, or changed content.

Hister already stores the text and metadata of indexed pages. The MCP
`get_preview` tool can retrieve that stored content directly.

That means an assistant can work from the version of the page you indexed,
instead of whatever the site returns later.

This is useful when:

1. The page changed after you read it.
2. The page is no longer available.
3. The page requires a session in your browser.
4. The page blocks automated fetches.
5. You want the assistant to work from your personal archive.
