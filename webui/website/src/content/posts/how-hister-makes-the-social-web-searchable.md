---
date: '2026-08-18T12:00:00+02:00'
draft: false
title: 'How Hister Makes the Social Web Searchable'
description: 'Turn posts and discussions from Bluesky, Mastodon, Reddit, X, and Discourse into a private full text archive you can search across platforms.'
---

Social networks are excellent places to discover useful information and terrible
places to find it again.

You may remember that someone explained a subtle PostgreSQL problem, linked to a
useful security report, or posted the perfect answer to a programming question.
You may even remember which community it came from. But after a few days of new
posts, notifications, and open tabs, the original discussion has disappeared
into the timeline.

Bookmarks help when you remember to create one. Platform search helps when it
supports the words and time range you need. Neither gives you one place to
search across everything you have read.

Hister takes a different approach. It indexes the content that reaches your
browser or crawler and stores it in a search engine you control. With dedicated
extractors for Bluesky, Mastodon, Reddit, X, and Discourse, it can understand
the difference between a timeline, an individual post, and a complete
discussion.

The result is a private, searchable memory of the social web you actually saw.

## Why a Social Feed Is Not a Normal Web Page

A traditional article is usually one page with one title and one body. A social
feed is a container that may hold dozens of unrelated posts. New items appear as
you scroll, old items disappear, and the page changes without navigating to a
new URL.

If a search engine treats that feed as one large document, several problems
appear. A result points back to the feed rather than to the post. Text from
different authors is mixed together. The next update can replace the content
that was indexed previously. A search result may tell you that a phrase existed
somewhere in a timeline without identifying who wrote it.

Hister extractors sit between page capture and indexing. They recognize the
structure of supported sites and turn the page into documents that match the
way people think about its content.

Bluesky, Mastodon, and X feeds are split into individual posts. Each post gets
its own URL and searchable text. Reddit and Discourse use a different model:
the complete discussion is kept together so the original post and its replies
remain useful as a unit.

This distinction is small, but it changes search from “these words appeared on
a feed” into “this person wrote this post in this conversation.”

## Bluesky Profiles, Feeds, and Threads

The Bluesky extractor handles profile pages, custom feeds, and post threads.
Every available post becomes a separate Hister document with its canonical
`bsky.app` URL, author, handle, publication time, and a
`type: bluesky` metadata value.

Bluesky publishes structured information for profile and individual post pages,
so Hister can often understand those pages with its normal HTTP crawler. Custom
feeds are rendered inside the browser and require either the `chromedp` or
`bidi` crawler backend. The [browser extension](/docs/browser-extension) can
also submit the rendered page you are already viewing.

When structured data and rendered content describe the same post, Hister
combines them using the canonical post URL. Rendered text is preferred when it
contains richer content, while missing author or publication details can still
come from the structured data.

To search only Bluesky posts, use:

```textplain
metadata.type:bluesky
```

Add remembered words just as you would in any other Hister search:

```textplain
metadata.type:bluesky "database migration"
```

## Mastodon Across the Fediverse

Mastodon creates an extra challenge because there is no single Mastodon domain.
The same timeline structure appears across thousands of independently operated
instances.

The Mastodon extractor detects the software from information in the page rather
than relying on a fixed hostname. It splits every toot found on the page into an
individual document and marks it with `type: toot` metadata.

Federation also means that a toot from one server may appear through another
server. When possible, Hister rewrites that federated link to the original
account server. Search results then point to the canonical source instead of an
instance that happened to display a copy.

The browser extension watches for page changes, so newly loaded toots can be
submitted while you use your normal timeline. You do not have to remember to
save each interesting post manually.

Search Mastodon content with:

```textplain
metadata.type:toot
```

For a detailed walkthrough, see
[Never Lose a Toot Again](/posts/never-lose-a-toot-again).

## Tweets Without the Timeline Noise

The X and Twitter extractor follows the same document model as Bluesky and
Mastodon. It recognizes feeds, profiles, and individual tweet pages, then
creates one document for every visible tweet. Each document can include its
canonical X URL, author, handle, and publication time, together with
`type: tweet` metadata.

Links inside tweets require special care. The visible page may use a `t.co`
address even when the browser already knows the original destination. Hister
uses the expanded destination when it is available. If a link cannot be
resolved safely, Hister keeps it as plain text instead of inventing a target.

This makes the indexed text more useful. Searching for a domain or product name
can find the original link rather than only finding Twitter's redirect service.

Search tweets with:

```textplain
metadata.type:tweet
```

## Reddit Threads as Complete Documents

A Reddit post is usually valuable because of the conversation below it. The
best explanation, correction, or link may be several replies deep. Splitting
every comment into an isolated result would lose that context.

The Reddit extractor therefore keeps the post and all comments already present
on the page in one document. It preserves reply depth in both indexed text and
the preview. Metadata can include the author, subreddit, publication time,
score, flair, post identifier, and number of extracted comments.

Only individual Reddit post pages are matched. Listings, profiles, and search
pages are ignored because they are temporary collections rather than complete
discussions. Hister does not request hidden comments or expand content that was
not present in the submitted page.

Search indexed Reddit discussions with:

```textplain
metadata.type:reddit
```

You can combine that filter with the words you remember:

```textplain
metadata.type:reddit "garbage collector" latency
```

## Discourse Topics With Their Solutions Intact

Discourse powers support forums and community sites across a wide range of
projects. These topics often contain information that never reaches the formal
documentation: migration fixes, configuration details, design explanations,
and solutions confirmed by maintainers.

The Discourse extractor keeps each topic as one searchable document. It
combines the original post with the regular replies already available on the
page, while preserving post numbers, authors, timestamps, reply targets, likes,
reactions, categories, tags, and accepted solutions when that information is
available.

System actions and hidden posts are excluded. Category lists, tag pages,
profiles, and search results are also ignored. The extractor focuses on the
topic page because that is the stable unit someone will want to retrieve later.

Search Discourse topics with:

```textplain
metadata.type:discourse
```

## One Search Across Every Community

The real value appears when you stop remembering which platform contained the
answer.

You can search every supported social document type with one query:

```textplain
metadata.type:(bluesky|tweet|toot|reddit|discourse)
```

Add a subject to search across all five sources:

```textplain
metadata.type:(bluesky|tweet|toot|reddit|discourse) sqlite backup
```

Add a date filter and sorting when you remember seeing it recently:

```textplain
metadata.type:(bluesky|tweet|toot|reddit|discourse) added:<7d sort:date
```

If that expression is too long for daily use, create an alias such as `!social`
in the Rules page and map it to the metadata expression. Your query then becomes:

```textplain
!social sqlite backup
```

Normal phrase matching, exclusions, labels, dates, and semantic search continue
to work. Semantic search is particularly helpful when you remember the idea of
a post but none of its exact vocabulary.

## Start Building Your Searchable Social Memory

Install [Hister](/docs/installing), follow the
[quickstart guide](/docs/quickstart), and connect the browser extension. Then
use Bluesky, Mastodon, Reddit, X, and Discourse as you normally would. Supported
pages are recognized automatically by the extractor chain.

You can inspect every available extractor and its exact matching behavior in
the [extractor documentation](/docs/extractors). If there is another community
site you rely on, extractor contributions are welcome in the
[Hister repository](https://github.com/asciimoo/hister).
