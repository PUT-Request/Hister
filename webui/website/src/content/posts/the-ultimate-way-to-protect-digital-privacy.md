---
date: '2026-06-19T15:00:00+02:00'
draft: false
title: 'The Ultimate Way to Protect Your Digital Privacy'
description: 'The only guaranteed way to preserve your digital privacy requires no special tooling. This post explains why, and offers practical methods to minimize internet exposure.'
---

So, what snake oil are we selling here?

- A new browser extension?
- Yet another VPN provider?
- A third-party service that promises never to store your data?

None of these. The only guaranteed way to fully preserve your digital privacy is simple and requires no tooling:

> **Solution: Do not use the internet.**

I know, I know, this sounds ridiculous to most of us, but let me explain.

---

The goal of this post is not to scare you away from the internet, but to raise awareness and enable conscious decision making. **The internet is a fascinating place once you know what you are doing.**

## Conditioned to the "Always Online" Mindset

Tech companies are incentivized to keep users always online. Constant connections let them build subscription-based services and continuously collect data from their users, which in turn makes their systems more efficient at serving and monetizing those same users. The scale of this is not hypothetical. The 2018 [Facebook and Cambridge Analytica scandal](https://en.wikipedia.org/wiki/Facebook%E2%80%93Cambridge_Analytica_data_scandal) revealed that the personal data of up to 87 million users had been harvested and turned into psychological profiles for political advertising.

An incredible amount of resources goes into shaping the digital world so that user tracking feels as convenient and natural as possible. Even when a company's primary goal is not tracking, its systems, or the tools those systems depend on, are by design capable of detailed user tracking and profiling. The capability exists whether or not it is the stated intent. A good example is the Strava fitness app, whose global activity heatmap was never meant as a surveillance tool, yet by aggregating the routes of users wearing connected fitness trackers exposed the layout and patrol patterns of secret military bases.

It helps to be concrete about what "tracking" actually means, because it is not one technique but a stack of them. Cookies persist your identity across sessions. Tracking pixels embedded in pages and emails report back when you open something. Browser fingerprinting recognizes your device from the unique combination of its settings, fonts, and screen size, without storing anything on your machine at all. Analytics and advertising SDKs bundled into apps and websites stream behavioral data to third parties. Even with all of these removed, the plain pairing of your IP address with request timestamps is often enough to correlate your activity over time. Blocking trackers at the browser helps, but it is never complete, because much of the collection happens on servers you cannot see or influence. For example in 2021 the Federal Trade Commission found that the popular period-tracking app Flo had shared users' sensitive health data with Facebook, Google, and other third parties through embedded SDKs, and in 2022 an investigation by The Markup showed that several major tax-filing services were quietly sending customers' financial details to Meta through a single tracking pixel.

> Trying to preserve online privacy is a never-ending cat-and-mouse game, except here the cat commands a high-tech drone army equipped with every sensor and gadget imaginable.
>
> One thing is often overlooked, though: the mice have a trick up their sleeve. At any moment, they can simply decide not to participate in the chase.

## How to Minimize Internet Exposure

Going fully offline is nearly impossible for most people, but several practical methods can significantly reduce the time you spend on the web.

### Self-Hosting

Many online services have excellent offline or self-hosted alternatives. [Collections of self-hosted software](https://awesome-selfhosted.net/) are a good starting point for finding privacy-respecting tools across most popular service categories.

By hosting your own wiki, feed reader, email infrastructure, or document management platform, you can replace a surprising number of online services quickly. Each service you bring in-house reduce data leaks.

### Physical Multimedia

While it is getting harder, you can still buy physical copies of music and movies as vinyl records, DVDs, and Blu-rays. Creating a digital backup of these makes them conveniently accessible without tying playback to a streaming account that logs what you watch and when.

It is also worth noting that many smaller artists offer a direct digital purchase option for their music on their own websites, which keeps a large intermediary out of the transaction.

### Shopping

Use cash instead of card transactions, and visit physical stores instead of webshops. Bank transactions and online purchases require sensitive personal information that can be stored for decades. Before deciding to buy online, always consider whether an offline alternative exists. The convenience of one click doesn't always justify the permanent record it creates.

### Local Caching

Building your own knowledge base can be a major step toward reducing internet usage.

Knowledge workers spend a significant portion of their time consuming digital information, and they often revisit the same information more than once. Documentation, scientific papers, Wikipedia pages: they frequently look up things they already knew but no longer remember in full.

This is exactly where [Hister](https://hister.org/) becomes a convenient companion. Hister automatically saves every page your browser renders and provides a search interface with offline previews of the results. As your local index grows, you find yourself needing to revisit the live web less and less often, because the answer is already stored on your own machine.

### Prefer Decentralization, Federation and Peer-to-Peer

Social media is inherently internet-dependent. It cannot be replaced by offline tools when the internet is your only channel to reach your peers. To minimize the risk, prefer services where the server you connect through is operated by an entity you trust: a friend, an organization, a local hackerspace, and so on.

Most of these platforms are not strictly peer-to-peer but federated, and the distinction matters. Social platforms built on the ActivityPub protocol, such as [Mastodon](https://mastodon.social/), [Lemmy](https://join-lemmy.org/), [Pixelfed](https://pixelfed.social/), [Pleroma](https://pleroma.social/), and [PeerTube](https://joinpeertube.org/), run as many independent servers, called instances, that talk to one another, rather than one company owning the entire network. This removes the single central operator, but it does not remove trust entirely: the administrator of the instance you join can still read your data, including private messages, because those messages are processed on their server. Choosing whose instance you join is therefore the real privacy decision. Running your own instance, or joining one operated by someone you actually know, keeps that trust close to home.

### Smart Devices and the Cloud

A growing category of hardware does not work at all without the internet. Smart speakers, video doorbells, security cameras, robot vacuums, smart TVs, thermostats, and countless other "smart home" gadgets route their core functionality through a vendor's cloud. The device in your home is often little more than a sensor and a network adapter, while the actual logic runs on servers you do not control. This design has real privacy consequences: many of these products carry always-on microphones or cameras, continuously report telemetry about how and when you use them.

The practical defense is to prefer devices that work locally and treat cloud connectivity as optional rather than mandatory. Look for hardware that supports local-control protocols like Zigbee, Z-Wave, or Matter, and pair it with a self-hosted hub such as [Home Assistant](https://www.home-assistant.io/) so that automations run on your own machine and never leave your network. Where a local option does not exist, the simplest answer is often the best one: a "dumb" appliance that does its job without an account, an app, or a connection to anyone's servers.

## Do Not Trust

Many services promise privacy or publish a detailed privacy and data handling policy, but in practice there is usually no way to validate those claims. The data processing happens server-side, inside closed systems you have no access to, so you cannot confirm whether data is genuinely deleted, whether logging is truly disabled, or whether your records are quietly shared with third-party processors and data brokers. Even data described as "anonymized" is frequently re-identifiable once it is combined with other datasets. You are effectively asked to trust a black box whose commercial incentives usually favor retaining as much data about you as possible.

Your data can still end up in the wrong hands even if a service is honest, either because it fails to honor its own statements, whether deliberately or by accident, or because it falls victim to a data breach. Worse, data rarely has a real expiry date. Copies live in backups, analytics pipelines, and downstream brokers, which is why a deletion request does not always result in true deletion, and why a breach years from now can still expose what you share today. In most cases you will not even know that your data was compromised.

Always be extra cautious with any service that handles your data. Counterintuitively, privacy-promoting services can sometimes cause more harm, because they encourage users to develop a false sense of security and to share more than they otherwise would.

## Closing Words

Do not let the internet's potential harms intimidate you. The physical world can be tough and unsafe too, yet by exploring it and learning its rules we minimize the risks and free ourselves to focus on the good things it offers. Apply the same principles to your online presence: stay aware, make deliberate choices, and remember that not participating is always an option.
