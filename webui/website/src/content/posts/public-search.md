---
date: '2026-06-25T10:00:00+02:00'
draft: false
title: 'Public Search Comes to Hister'
description: 'Public search lets admins run Hister as a public search engine, making shared discovery possible while keeping write access protected.'
---

Hister started as a private search engine. The first goal was simple: make your own browsing history, saved pages, files, documentation, and crawled knowledge searchable without sending queries to someone else's infrastructure.

That goal still matters. Private recall search is still one of the strongest reasons to run Hister. But the project has grown beyond that original shape. The crawler, extractors, document storage, semantic search, public datasets, API, and MCP endpoint all point in a broader direction:

> Hister can become a search engine for any corpus an admin decides to build.
>
> Public search is a major step in that direction.

With public search enabled, a Hister instance can serve anonymous visitors. Admins can host a traditional public search interface over a shared index while keeping indexing, deletion, rule changes, profile APIs, cleanup, and other write operations behind authentication. In practice this means a Hister server can now be more than a personal tool. It can be a public search engine for a documentation collection, a community archive, a research dataset, an organization knowledge base, a topical web index, or a general purpose public crawl.

## From Recall Search to Discovery Search

I have written before that part of everyday search is recall search. You are not always trying to discover something new. Often you are trying to refind an article, documentation page, bug report, or note you have already seen.

Hister has been strong at that use case from the beginning because the index is built from content you control.

The limitation has always been coverage. A private index is excellent when the thing you need is already inside it. It is weaker when you are doing a discovery search and you do not yet know where the answer might live.

Public search changes the shape of that problem. If an admin runs a public Hister instance with a broad enough index, visitors can use it for discovery type searches.

This does not magically make Hister a Google scale search engine. It does make the path toward a general purpose online search engine much more concrete.

## Why Public Search Matters

Public search mode is the bridge between Hister as a private recall engine and Hister as a possible general purpose search engine. Hister can still be your private search engine, that remains the default and the safest setup for personal data.

## What Public Search Does

Public search lets anonymous visitors use the search surface of a Hister instance. That includes search, suggestions, document reads, previews, file serving, API documentation, and MCP search.

Write access remains authenticated. Anonymous users cannot add documents, edit labels, delete results, modify rules, reindex, clean up data, or access profile endpoints. Hister also refuses to start with public search unless either an access token or user handling is configured, so admins have to keep a protected path for management.

You can enable it in the config:

```yaml
app:
  public: true
  access_token: 'your-secret-token-here'
```

Or start a configured authenticated server with:

```bash
hister listen --public
```

When user handling is enabled, anonymous visitors see only global documents stored under user ID `0`. Documents owned by named users stay private to those authenticated users.

This is intentionally conservative. Public search is not "make everything open". It is "let anonymous users query the public part of this instance". Admins still decide what goes into that public index.

> If a public instance indexes local files, previews, or pages with sensitive content, anonymous visitors may be able to read that content. Public search is for material that is meant to be public.

## The Bigger Direction

There is an open discussion about [result and index sharing](https://github.com/asciimoo/hister/discussions/432). The core idea is that Hister users and admins should eventually be able to share parts of their indexes with others. That could let people contribute pages they index into a shared search engine over time, or publish reusable datasets for other instances.

The discussion also raises the right hard questions:

1. How can users submit data in a secure and privacy respecting way?
2. Where should shared data live?
3. Should the system be centralized, peer to peer, or admin hosted?
4. How can submitted data be verified?
5. How do we prevent malicious or low quality contributions from poisoning the index?

Public search does not answer all of those questions. It gives us the first useful deployment model: an admin can host a public Hister instance using the existing crawler, extractors, indexer, previews, search API, and user controls.

### Distributed Crawling and Federated Instances

The long term direction I am interested in is distributed crawling and federated Hister instances.

Federated or interconnected instances would let Hister servers exchange selected public results, datasets, or index metadata. **The goal is not to force every instance into one global service. The goal is to let admins decide who they trust, what they share, and what they import.**

---

There is still a lot to build: better distributed crawl scheduling, index sharing, trust and verification, result exchange, quality scoring, abuse controls, and admin tools for curation. But public search makes those problems practical.

I'm looking forward to further enhance this aspect of Hister, and I would love to hear your ideas, use cases, and contributions.
