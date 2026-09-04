---
date: '2026-08-04T12:00:00+02:00'
draft: false
title: 'New Import Command Structure and Sources'
description: 'Hister now supports a unified import command for files, browser history, Linkding, Linkwarden, Karakeep, Shaarli, and wallabag.'
---

Hister import commands are now grouped under `hister import`. Each import type
has a source subcommand:

```bash
hister import file export.json
hister import browser
hister import linkding https://linkding.example.com
hister import linkwarden https://linkwarden.example.com
hister import karakeep https://karakeep.example.com
hister import shaarli https://shaarli.example.com
hister import wallabag https://wallabag.example.com
```

All imports submit documents to the configured Hister server.

Use the import help command to list available sources and general usage. Each
source subcommand also provides its own usage and additional flags:

```bash
hister import --help
hister import linkding --help
```

## Command Changes

When upgrading from Hister v0.16.0, update existing import commands:

```bash
# Previous
hister import export.json
hister import-browser firefox

# Current
hister import file export.json
hister import browser firefox
```

## File Imports

`hister import file` supports Hister JSON exports, 7z archives containing a JSON
export, and saved HTML files. It accepts multiple files and directories.

Directory imports process supported files in filename order and do not include
nested directories.

## Browser Imports

`hister import browser` accepts a browser name, a browser database path, or both.
With no arguments, Hister detects available browser databases automatically.

Interrupted browser imports can be resumed. Their status and failed URLs are
available through the `hister crawl` commands.

## Service Imports

Hister can import saved content from the following services:

1. Linkding
2. Linkwarden
3. Karakeep
4. Shaarli
5. wallabag

Service imports retain available titles, content, dates, tags, state, favicons,
and source data. Missing page content and favicons are retrieved when possible.

Repeated service imports request newer records when supported by the source.
Deleting a record from the source does not delete the corresponding Hister
document.

## Labels and Options

The default labels are `import` for files, `browser` for browser history, and the
service name for service imports. Use `--label` to set a different label:

```bash
hister import linkwarden https://linkwarden.example.com --label reading
```

Service imports support date filtering, existing URL checks, batch size
selection, crawler options, and user selection. For example:

```bash
hister import karakeep https://karakeep.example.com \
  --start-date 2025-01-01 \
  --skip-existing \
  --batch-size 25
```

The `--user-id` and `--global` options require an administrator in multiuser
mode.

## Credentials

Service credentials can be provided through source specific environment
variables or the `--api-token` option. The global `--token` option authenticates
with the destination Hister server and is separate from the source credential.

See the [import documentation](/docs/import) for credential names, supported
options, and source specific data mappings.

Suggestions for additional import sources are welcome. Open an issue on
[GitHub](https://github.com/asciimoo/hister/issues) with the service name and its
available export format or API documentation.
