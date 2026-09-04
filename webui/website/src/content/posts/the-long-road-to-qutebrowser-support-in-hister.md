---
date: '2026-07-24T12:00:00+02:00'
draft: false
title: 'The Long Road to Qutebrowser Support in Hister'
description: 'How several failed attempts to submit rendered qutebrowser pages led to a safer and simpler DevTools companion for Hister.'
---

Adding qutebrowser support to Hister sounded simple.

Hister already had browser extension code that could extract the title, text,
HTML, and favicon from the current page. However, qutebrowser does not support
browser addons or the WebExtensions API, so users cannot simply install the
existing Hister extension. We had to find another way to integrate with the
browser.

Qutebrowser supports Greasemonkey scripts, and Hister exposes an API for adding
documents. The apparent plan was to connect those two pieces with the existing
extraction code and call it a day.

The extraction part worked. Sending the result to Hister did not.

What followed was a tour through qutebrowser userscripts, CORS, preflight
requests, Content Security Policy, opaque responses, form submission, CSRF
protection, and several increasingly uncomfortable server exceptions. The
final solution uses none of those workarounds. Hister now talks to qutebrowser
through the Qt WebEngine DevTools endpoint.

This post explains how we got there.

## The One Requirement We Could Not Compromise

The point of browser integration is not merely to save the URL of the page.
Hister should receive the page as the user actually saw it.

Modern sites often render most of their useful content with JavaScript. The
initial HTTP response may contain little more than an empty application shell.
The browser then loads data, applies authentication state, expands components,
and changes the DOM as the user interacts with the page.

A script running inside qutebrowser can read that final DOM:

```javascript
const page = {
  title: document.title,
  text: document.body?.innerText ?? '',
  html: document.documentElement?.innerHTML ?? '',
  url: window.location.href,
};
```

That is the valuable part. Replacing this with `curl`, a Go HTTP client, or a
server side crawler would mean downloading the page again. The second request
might receive different content, miss the authenticated state, trigger rate
limits, or fail behind an anti bot page. More importantly, it would throw away
the rendered document that was already available in the browser.

Any acceptable design therefore had to preserve this flow:

```text
rendered browser DOM → Hister document
```

No second page download.

## Attempt One: Native Qutebrowser Userscripts

Qutebrowser has a native userscript system that can run external programs with
information about the current page. This looked promising because an external
program could submit data to Hister without being restricted by browser CORS
or page CSP.

The problem was not delivery. It was triggering the script.

There is no reliable native userscript hook that runs automatically when a page
finishes loading or when its DOM changes. A user can invoke a script manually
through a command or key binding, but Hister browser integration is expected to
capture pages automatically while the user browses.

Polling from another process would not solve this cleanly either. That process
would still need a way to discover tabs, detect navigations, observe DOM
changes, and retrieve the current rendered content. At that point it would be
reimplementing the browser event interface we did not yet have.

Native userscripts could support a useful manual "save this page" command, but
they could not provide automatic browsing history capture.

## Attempt Two: Fetch From Greasemonkey

The first implementation generated a qutebrowser Greasemonkey script from the
page extraction code used by the regular browser extension. The script could
read the rendered DOM, so the obvious next step was a `fetch` call to
`http://127.0.0.1:4433/api/add`.

This is where qutebrowser differs from browser extension environments and from
userscript managers that provide privileged request functions.

The qutebrowser Greasemonkey implementation does not provide a privileged
cross origin request API. Its `fetch` is still a page request. The request is
subject to the origin of the page, the browser CORS implementation, and the
page Content Security Policy.

Setting a custom `Origin` header in the script was not an escape hatch. The
browser owns that header and page JavaScript is not allowed to replace it. A
script cannot claim to have the trusted `hister://` origin used by Hister
command line clients.

Hister normally authenticates API requests with an `X-Access-Token` header.
Adding that header to a cross origin request makes the request non simple, so
the browser sends an `OPTIONS` preflight before the real `POST`. The server
would have to approve the page origin and the custom header before the browser
would send the document.

We considered allowing `/api/add` when a valid access token was present. The
token does authenticate the caller, but it does not remove the browser CORS
rules. The browser must decide whether it is allowed to send the header before
the server gets a chance to validate its value.

There was another problem. Many pages define a policy similar to:

```text
Content-Security-Policy: default-src 'self'
```

If `connect-src` is not defined, `default-src` is used as its fallback. In that
case the browser blocks the request to the local Hister server before CORS is
even relevant.

The exact page we wanted to archive controlled whether the archive request was
allowed. That is not a reliable integration.

## Why Curl Still Was Not Enough

The native userscript approach could call `curl`, set the expected origin,
include the access token, and send page data without browser CORS enforcement.
It still had no automatic page load or content change trigger.

A Greasemonkey script had the event and DOM access we needed, but browser
JavaScript could not simply launch an arbitrary external process. Running curl
independently would require it to download the page again.

That defeats the central requirement. An independent curl request does not
have access to the live DOM inside the tab. It cannot see content added by
JavaScript, the current authenticated view, or the exact state the user read.

Curl solved transport when invoked manually through a native userscript. It did
not solve automatic triggering. Used independently, it solved the easy half of
the problem by deleting the important half.

## Attempt Three: `no-cors`

The next idea was `fetch` with `mode: "no-cors"`.

The name is easy to misread. It does not mean "ignore all cross origin
protections." It creates a restricted request whose response is opaque to the
calling script.

In this mode the script cannot attach an arbitrary `X-Access-Token` header.
The request must stay within the definition of a simple request. The response
status and body are not available to the script. The request also remains
subject to the page Content Security Policy, so a restrictive `connect-src`
still blocks it.

There is an important security consequence too: `no-cors` is available to
ordinary website JavaScript. It is not a special qutebrowser capability. Any
website can attempt a simple request to a service on localhost.

That does not automatically make such a service vulnerable. An explicit bearer
token in the request can still authenticate the action. It does mean that
`no-cors` itself provides no security boundary. Hister could never interpret
"the browser sent this request" as proof that the request came from its own
integration.

We could put the access token in a form encoded request body, but that led
directly to the next attempt.

Another variation was to put the document in a URL and let `/api/add` modify
state through `GET`. That would have made the endpoint callable through even
more browser primitives, exposed document data in URLs and logs, and violated
the expected semantics of `GET`. It would have discarded an important API
boundary just to accommodate a restricted transport.

## Attempt Four: Submit a Form

Classic HTML forms can submit cross origin `POST` requests without CORS
preflight. A form could include the rendered page fields and an access token:

```html
<form method="post" action="http://127.0.0.1:4433/api/add">
  <input name="url" />
  <input name="title" />
  <input name="text" />
  <input name="html" />
  <input name="favicon" />
  <input name="access_token" />
</form>
```

The server could accept the token from the request body, validate it, and treat
the explicit secret as protection against CSRF. Unlike a session cookie, the
browser does not attach the token automatically.

This approach got much further. It also accumulated warning signs.

First, a form submission is governed by the page `form-action` policy. A common
policy is:

```text
Content-Security-Policy: form-action 'self'
```

This permits forms to submit only to the origin of the page being viewed. A
form running on an arbitrary website therefore cannot submit to the local
Hister origin. Some pages apply the even stricter policy:

```text
Content-Security-Policy: form-action 'none'
```

That blocks every form submission. Unlike a response handling bug or an API
format mismatch, the script cannot work around either policy. The page tells
the browser not to submit the form to Hister, and the browser correctly
enforces that decision. This was the decisive reason we abandoned the form
approach.

Second, the token had to pass through a mechanism associated with the page
DOM. We experimented with detached forms and closed shadow roots to reduce
exposure. Even if a particular construction could be made safe, it depended on
subtle assumptions about isolated JavaScript worlds, shared DOM objects, page
mutation observation, and browser implementation details. This was too fragile
for an access token with permission to modify a Hister instance.

Third, normal form submission has browser navigation behavior. If Hister
returned `406` because a skip rule rejected the page, or `500` because indexing
failed, qutebrowser could navigate to that response.

To hide this behavior, the server learned to identify Greasemonkey submissions
and replace every result with an empty `204` response. That kept the page in
place, but it also hid useful status information. A successful index, a policy
rejection, and an internal error all looked identical to the script.

The server changes kept growing:

1. Accept an access token in a form field for one endpoint.
2. Let that token bypass the normal CSRF path.
3. Decode extra document fields from form data.
4. Recognize a special Greasemonkey client marker.
5. Capture the real handler response.
6. Replace that response with `204`.

At one point the response wrapper even caused a warning about a superfluous
`WriteHeader` call. The warning was fixable, but it highlighted the larger
problem. A browser integration was forcing special authentication and response
semantics deep into the server.

None of these pieces was absurd on its own. Together they added considerable
complexity without producing a solution that worked everywhere, since page CSP
rules could still block submission entirely. That was a sign that the
integration boundary was wrong.

## Changing the Boundary

The breakthrough was to stop making the inspected page responsible for
delivery.

Qutebrowser is built on Qt WebEngine, which exposes the Chrome DevTools
Protocol when remote debugging is enabled. The protocol can report tabs and
navigations, evaluate JavaScript in a page, and create an isolated JavaScript
world.

Start qutebrowser with a loopback debugging endpoint:

```bash
QTWEBENGINE_REMOTE_DEBUGGING=127.0.0.1:9222 qutebrowser
```

All existing qutebrowser processes must be closed first. Qutebrowser may reuse
an already running process, and that process will not inherit a debugging
setting added to a later command. This was the cause of an initially confusing
connection refused error during development.

Then run:

```bash
hister companion qutebrowser
```

The companion connects to the local DevTools endpoint, discovers page targets,
and attaches to them. It evaluates the extraction code inside an isolated
world, where it can read the rendered DOM without sharing variables with page
scripts.

The page does not send anything to Hister. The page only supplies content to
the companion through DevTools.

The Go process then submits the document using the normal Hister client. It can
set the access token header, receive the real HTTP status, and use the existing
command line and configuration options. CORS and page CSP do not apply because
the request does not originate from the page.

Most importantly, the companion still uses the existing rendered DOM. It does
not crawl or download the page again.

## Detecting Page Changes

Initial page loading is only part of the problem. Single page applications can
replace most of their content without a traditional navigation.

The companion installs a small `MutationObserver` in the isolated world. The
observer does not send page content. It only calls a DevTools binding to notify
the Go process that something changed. The companion then performs a fresh
extraction.

New pages are submitted after one second so the initial capture feels quick.
Later DOM changes use a ten second quiet period. A thirty second maximum wait
ensures that a page with continuous updates is eventually captured.

The companion fingerprints the page URL and HTML before submission. If the
content has not changed, it does not send another document. Only one extraction
runs for a page at a time, and an update that becomes ready during extraction
runs immediately afterward.

This is a much smaller state machine than trying to infer success through
opaque browser responses.

## Authentication and Failure Handling

The access token now stays entirely inside the Hister process. It is never
inserted into the inspected page or its DOM.

The companion receives real errors from the API. Temporary failures are
retried. A `406` skip rule rejection and a `422` sensitive content rejection
are treated as completed decisions rather than transient failures. DevTools
connections are reestablished when qutebrowser restarts or the debugging
connection disappears.

This also allowed us to delete all special Greasemonkey behavior from the
server. Form tokens no longer bypass authentication or CSRF checks, and
`/api/add` no longer needs to disguise errors as successful empty responses.

The final integration uses the same security model as every other Hister
command line client.

## The New Security Tradeoff

The DevTools solution is cleaner, but it is not free of security concerns.

A browser debugging endpoint has extensive access to open tabs. Anyone who can
connect to it may be able to inspect page content and control the browser. The
companion therefore accepts only `localhost` or a loopback IP address as its
DevTools host. It also replaces wildcard hosts advertised by Chromium with the
exact loopback host selected by the user.

The port must never be exposed to a network.

Favicon fetching needed similar care. A page controls its favicon URL, so
letting the Go process fetch any arbitrary URL would create an unnecessary
request primitive outside the browser. The companion only downloads favicons
from the same origin as the page, keeps redirects on that origin, and limits
the response size. Some externally hosted favicons will be skipped. That is a
deliberate tradeoff.

Moving code outside the browser page removed the CORS and CSP problems. It did
not remove the need to define a careful trust boundary.

## The Result

Qutebrowser support ended up larger than the first Greasemonkey script, but the
architecture is simpler.

The browser renders the page. DevTools exposes the rendered DOM to a local
companion. The companion sends the document through the normal authenticated
Hister API.

Each component now does the job it is suited for, and none of them has to
pretend that browser security controls do not exist.

You can find the setup instructions in the
[browser integration documentation](/docs/browser-extension).
