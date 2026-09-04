// Package urlutil provides shared URL helpers for extractors.
package urlutil

import (
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ResolveURL resolves ref against base. Returns ref unchanged if it is already
// absolute, a fragment, or a data URI. Protocol-relative URLs (//host/...)
// are resolved using the base scheme.
func ResolveURL(base *url.URL, ref string) string {
	if ref == "" || strings.HasPrefix(ref, "#") || strings.HasPrefix(ref, "data:") {
		return ref
	}
	// Handle protocol-relative URLs (//upload.wikimedia.org/...)
	if strings.HasPrefix(ref, "//") {
		return base.Scheme + ":" + ref
	}
	u, err := url.Parse(ref)
	if err != nil || u.IsAbs() {
		return ref
	}
	return base.ResolveReference(u).String()
}

// RewriteURLs rewrites relative href, src, and srcset attributes to absolute
// URLs using base. No-op if base is nil.
func RewriteURLs(s *goquery.Selection, base *url.URL) {
	if base == nil {
		return
	}
	for _, attr := range []string{"href", "src", "srcset"} {
		s.Find("[" + attr + "]").Each(func(_ int, el *goquery.Selection) {
			if v, ok := el.Attr(attr); ok {
				if attr == "srcset" {
					v = ResolveSrcset(base, v)
				} else {
					v = ResolveURL(base, v)
				}
				el.SetAttr(attr, v)
			}
		})
	}
}

// ResolveSrcset rewrites each URL in a srcset attribute value against base.
func ResolveSrcset(base *url.URL, srcset string) string {
	var parts []string
	for entry := range strings.SplitSeq(srcset, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		fields := strings.Fields(entry)
		if len(fields) >= 1 {
			fields[0] = ResolveURL(base, fields[0])
		}
		parts = append(parts, strings.Join(fields, " "))
	}
	return strings.Join(parts, ", ")
}
