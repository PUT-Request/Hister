---
date: '2026-04-20T12:00:00+00:00'
draft: false
title: 'AI-Powered Semantic Search in Hister'
description: 'Hister now supports semantic search: an AI-powered layer that understands the meaning behind your query and surfaces relevant pages even when they use completely different words.'
---

I'm happy to announce that Hister now has an **optional** semantic search layer.

Keyword search is very good at one thing: finding pages that contain the exact
words you type. That works well when you remember the precise terminology. It
works less well when you remember the idea but not the phrasing.

You know the feeling. You read a great article about handling failures in
distributed systems last month. You search for "distributed failure handling"
and get nothing, because the article used the phrase "fault tolerance" and
"resilience patterns" throughout. You try "fault tolerance" and find it, but
only after a few attempts.

Semantic search handles exactly this kind of situation.

## What Semantic Search Does

Semantic search understands what you are looking for rather than just matching
characters. When it is enabled, Hister reads every document you index,
identifies the core ideas, and stores a compact mathematical representation of
those ideas alongside the regular search index.

When you search, Hister converts your query into the same kind of
representation and finds the pages whose content is closest in meaning, even
when the words do not overlap at all.

In practice this means:

- Searching "fault tolerance" can surface a page that only mentions "resilience" and "graceful degradation"
- Searching "how to speed up my build" can find your notes about "compilation caching and incremental builds"
- Searching "that recipe with the roasted garlic" can find a page titled "Mediterranean chickpea stew" that happens to start with roasted garlic

The keyword search you already know is still there. Semantic search does not
replace it. Hister merges both sets of results and re-ranks them, so you
benefit from both at the same time.

## When It Helps Most

Semantic search is particularly useful when:

- You are looking for something you read a while ago and can only remember the
  general topic
- You are researching an unfamiliar area and do not know the exact terminology
  yet
- You indexed pages in another language and want to find them with a query in
  your own language
- The author used different vocabulary than you would use to describe the same
  idea

Regular keyword search remains better when you know the exact term you are
looking for. Both modes are available on every search, so you do not have to
choose in advance.

## Privacy

When you use a local model through Ollama, your documents and queries never
leave your machine. The embedding model runs on your hardware, producing
vectors locally. No page content, no search query, and no personal data is
sent anywhere.

If you choose to use a hosted provider, your document text and queries are
sent to that provider to generate embeddings. That is the same trade-off you
make when using any cloud AI service. The local option is there specifically
for people who want the capability without that trade-off.

## Tuning the Balance

Semantic search adds a weight slider to the search interface. It controls how
much influence the semantic results have relative to the keyword results.

At one end of the slider, you get pure keyword search: fast, precise, and
exactly what Hister has always done. At the other end, you get pure semantic
search: better at concepts and synonyms, but occasionally more speculative. In
the middle you get both signals blended together.

The default is a moderate mix that favors keyword matches slightly. You can
drag the slider toward semantic for exploratory searches ("something about that
deployment pattern I read about") and leave it in the middle for everyday use.

## A Word on Performance

Generating embeddings takes a more time than adding it to the index depending on what embedding model you choose.

For ongoing use the overhead is small. A document lands in the index at similar speed as before. The embedding step runs asynchronously, so you never wait for it when browsing. However the reindex time and resource requirements can increase.

## Getting Started

The full configuration reference is at [hister.org/docs/configuration](/docs/configuration).
If you have questions or run into issues, open an issue on
[GitHub](https://github.com/asciimoo/hister).
