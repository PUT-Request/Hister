package org

import (
	"strings"

	"github.com/asciimoo/hister/server/extractor/sdk"
	"github.com/asciimoo/hister/server/sanitizer"
)

// OrgModeExtractor serves previews for locally indexed Org files.
// During indexing, Indexer.AddOrg renders the source to HTML and stores
// it in doc.HTML, so Preview only needs to sanitize that HTML.
type OrgModeExtractor struct {
	sdk.ConfigSupport
}

func (e *OrgModeExtractor) Name() string { return "OrgMode" }

func (e *OrgModeExtractor) Description() string {
	return "Renders locally indexed Org files (.org) as HTML for preview."
}

func (e *OrgModeExtractor) Capabilities() sdk.Capabilities {
	return sdk.Capabilities{Preview: true}
}

// Match returns true for file:// URLs with an .org extension.
func (e *OrgModeExtractor) Match(d *sdk.Document) bool {
	if !strings.HasPrefix(d.URL, "file://") {
		return false
	}
	lower := strings.ToLower(d.URL)
	return strings.HasSuffix(lower, ".org")
}

// Extract is a no-op: indexing is handled by Indexer.AddOrg.
func (e *OrgModeExtractor) Extract(_ *sdk.Document) sdk.ExtractResult {
	return sdk.ExtractFallback(nil)
}

// Preview sanitizes the rendered HTML stored in doc.HTML.
func (e *OrgModeExtractor) Preview(d *sdk.Document) sdk.PreviewResult {
	if d.HTML == "" {
		return sdk.PreviewFallback(nil)
	}
	return sdk.Previewed(sdk.PreviewResponse{Content: sanitizer.SanitizeHTML(d.HTML)})
}
