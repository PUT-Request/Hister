---
date: '2026-05-20T10:00:00+02:00'
draft: false
title: 'Lessons I Learned from Creating Searx'
description: 'Seven years of maintaining a popular open source metasearch engine taught me a great deal about privacy, community, and the fundamental limits of the metasearch approach. Here is what led me to build Hister.'
---

Some of you may already know that I am the original author of [SearX](https://github.com/searx/searx), a self-hosted metasearch engine that is nowadays continued as the popular fork [SearXNG](https://github.com/searxng/searxng). I stepped away from active development some time ago, but the project shaped how I think about search, privacy, and open source software in ways that are still very much with me today.

Over seven years I reviewed and merged nearly 1500 pull requests from more than 150 contributors. The project significantly shaped how I approach search and privacy related issues. I want to share my biggest take aways here, along with the thinking that led me to start building [Hister](https://github.com/asciimoo/hister).

## A Community Worth Celebrating

Before anything else, I want to say this clearly: the SearX community was the finest I have ever had the privilege of being part of. The number of people who cared deeply about privacy, who gave their time to improve a tool they believed in, and who brought genuine creativity and passion to the work was unexpected and humbling.

It is still a big motivation to me to continue developing free software in this field and to try to come up with new ideas to further protect our online privacy.

## Where SearX Fell Short

The community was exceptional. The architecture however had hard limits baked in from the start.

SearX is a metasearch engine. Every search query you type is forwarded to third-party engines (Google, Bing, and others), and the results are aggregated and returned to you. This model can solve some privacy problems, but it introduces others that cannot be patched away.

### Inherent Metasearch Limitations

**Results are only as good as what upstream engines expose.** You are entirely dependent on what Google or Bing decide to surface. If a page has been deindexed, filtered, or simply ranked poorly by those engines, you will not see it regardless of how good your instance is.

**Privacy guarantees are partial.** Every query still leaves your machine and travels to an external provider. Even if your own IP is hidden behind a server, the query itself, its timing, and a source IP are visible to the upstream engine. If a provider correlates queries from the same IP over time, or if your instance is the sole source of traffic from that address, the anonymisation can decline.

**Relevance ranking is out of your hands.** You can reorder scraped results, but the fundamental signal comes from somewhere else. You cannot boost a website you trust or suppress one you find unhelpful.

**Querying is fragmented.** Different backends support different search syntaxes. Trying to write a precise query that works consistently across multiple engines is genuinely difficult.

**There is no memory.** A metasearch engine has no idea what you have already read, what was useful to you, or what you searched for last week. Every session starts from zero.

### Broader Problems with Web Search

Beyond the metasearch-specific issues, I also kept running into limitations that affect any search engine that relies on fetching live pages from the web.

There is **no offline access**. If you lose your internet connection, or if a website goes down, the content is simply gone.

**Visiting webpages is required to access information**. Even with a privacy-preserving front end, the act of clicking a result can be a privacy/security risk.

There is **no way to control which domains appear in your results** without writing custom filtering rules outside the search tool itself.

**Content that requires authentication, or that simply never gets indexed by major engines is permanently out of reach.**

> These limitations are not bugs that can be fixed with better code inside the current (meta)search model. They are consequences of the model itself. Solving them requires a different starting point.

## What Hister Does Differently

Hister takes a fundamentally different approach: instead of forwarding queries to other engines, it indexes content you choose. Web pages you have visited, local files on your machine, your browser history. Everything goes into a fully local, self-hosted full-text search index that lives on your own hardware and never leaves it.

This removes most of the conceptual weaknesses that come with metasearch. Your queries never leave your machine. The index reflects what you actually care about, not what an algorithm decided to surface. You can boost or suppress any domain. The index persists across sessions, so the tool learns what you have already found.

One consequence I find particularly valuable: Hister can store a full copy of every page you index and serve it back to you offline. You can read the content without ever visiting the original site again. That is a big boost for both privacy and reliability.

## Combining the Two

Hister cannot provide a complete replacement to existing web search engines yet. There are times when you need to search beyond your own collection, and for those moments SearX can provide a fallback solution.

The recommended approach is to use Hister as your primary interface and fall back to SearX when your local index comes up empty. Hister has a built-in hotkey that opens your current query in whatever external search engine you have configured, making the transition seamless.

You get the privacy and depth of a fully local index for the things you have already encountered, and the breadth of a metasearch engine for everything else.

## Looking Back

SearX was absolutely worth building. The experience taught me things I could not have learned any other way, and the connections I made through it have been some of the most rewarding of my professional life. I am genuinely grateful to everyone who contributed code, filed issues, wrote documentation, translated strings, or simply ran an instance and told others about it.

If the ideas behind SearX resonated with you, I think Hister is worth a look. It grew out of the same values and the same frustrations, and it tries to address the things that always felt just out of reach.
