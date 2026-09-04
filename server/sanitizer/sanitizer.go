package sanitizer

import (
	stdhtml "html"
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

var (
	htmlSanitizerPolicy        *bluemonday.Policy
	htmlSanitizerPolicyTrusted *bluemonday.Policy
	textSanitizerPolicy        = bluemonday.StrictPolicy()
)

// trustedLayoutStyles are positioning/layout CSS properties that are denied by
// the default policy because attacker-controlled content could use them to
// overlap the host page (e.g. a fake URL bar positioned over the real one).
// They are permitted only by SanitizeTrustedHTML, for extractors whose source
// is editorially moderated (e.g. Wikipedia).
var trustedLayoutStyles = []string{
	"display", "position", "float",
	"top", "left", "right", "bottom",
}

// SVG elements allowed in sanitized output.
// foreignObject, use, and symbol are excluded because they can embed or
// reference third-party content (HTML subtrees, external SVG via xlink:href).
var svgElements = []string{
	"svg", "g", "path", "rect", "circle", "ellipse", "line",
	"polyline", "polygon", "text", "tspan", "defs", "clipPath",
	"mask", "linearGradient", "radialGradient",
	"stop", "marker", "pattern",
}

// Reusable validation patterns for SVG attribute values.
var (
	reNum       = regexp.MustCompile(`^-?[\d.]+(%|em|ex|px|pt|cm|mm|in)?$`)
	reNumNone   = regexp.MustCompile(`^(-?[\d.]+(%|em|ex|px|pt|cm|mm|in)?|none)$`)
	reNumPair   = regexp.MustCompile(`^-?[\d.]+ -?[\d.]+$`)
	reViewBox   = regexp.MustCompile(`^[\d.\- ]+$`)
	reTransform = regexp.MustCompile(`^[a-zA-Z(),.\d\s\-]+$`)
	reFill      = regexp.MustCompile(`^(none|currentColor|transparent|#[0-9a-fA-F]{3,8}|[a-zA-Z]+|var\(\s*--[a-zA-Z0-9-]+\s*(,\s*#?[a-zA-Z0-9]+\s*)?\))$`)
	reURLRef    = regexp.MustCompile(`^(none|url\(#[a-zA-Z0-9_-]+\))$`)
)

// svgAttrRule maps a set of attribute names to a validation regex and target elements.
type svgAttrRule struct {
	attrs    []string
	pattern  *regexp.Regexp
	elements []string // nil means svgElements
}

// svgAttrRules defines all SVG attribute allowlists declaratively.
// Rules with nil elements default to all SVG elements.
var svgAttrRules = []svgAttrRule{
	// Geometry.
	{[]string{"x", "y", "x1", "y1", "x2", "y2", "cx", "cy", "r", "rx", "ry", "width", "height", "dx", "dy"}, reNum, nil},
	{[]string{"viewBox"}, reViewBox, []string{"svg", "symbol", "pattern", "marker"}},
	{[]string{"preserveAspectRatio"}, regexp.MustCompile(`^(none|xMi[dn]YMi[dn]|xMa[dx]YMa[dx]|xMi[dn]YMa[dx]|xMa[dx]YMi[dn])(\s+(meet|slice))?$`), []string{"svg", "symbol", "pattern", "marker"}},

	// Path data.
	{[]string{"d"}, regexp.MustCompile(`^[MmZzLlHhVvCcSsQqTtAaEe\d\s,.\-+]+$`), []string{"path"}},
	{[]string{"points"}, regexp.MustCompile(`^[\d\s,.\-]+$`), []string{"polyline", "polygon"}},

	// Presentation (safe subset — no url() or expression()).
	{[]string{"fill", "stroke", "stop-color"}, reFill, nil},
	{[]string{"fill-opacity", "stroke-opacity", "opacity", "stop-opacity", "stroke-width"}, reNum, nil},
	{[]string{"stroke-dasharray"}, regexp.MustCompile(`^[\d., ]+|none$`), nil},
	{[]string{"stroke-linecap"}, regexp.MustCompile(`^(butt|round|square)$`), nil},
	{[]string{"stroke-linejoin"}, regexp.MustCompile(`^(miter|round|bevel)$`), nil},
	{[]string{"fill-rule", "clip-rule"}, regexp.MustCompile(`^(nonzero|evenodd)$`), nil},

	// Text.
	{[]string{"text-anchor"}, regexp.MustCompile(`^(start|middle|end)$`), []string{"text", "tspan"}},
	{[]string{"font-size", "font-family", "font-weight"}, nil, []string{"text", "tspan", "svg"}}, // uses bluemonday.Paragraph

	// Transform.
	{[]string{"transform"}, reTransform, nil},

	// Gradient / marker.
	{[]string{"offset"}, reNumNone, []string{"stop"}},
	{[]string{"gradientTransform"}, reTransform, []string{"linearGradient", "radialGradient"}},
	{[]string{"gradientUnits", "patternUnits"}, regexp.MustCompile(`^(userSpaceOnUse|objectBoundingBox)$`), []string{"linearGradient", "radialGradient", "pattern"}},
	{[]string{"markerWidth", "markerHeight", "refX", "refY"}, reNum, []string{"marker"}},
	{[]string{"orient"}, regexp.MustCompile(`^(auto|auto-start-reverse|[\d.]+)$`), []string{"marker"}},
	{[]string{"span"}, regexp.MustCompile(`^\d+$`), []string{"col", "colgroup"}},

	// Misc.
	{[]string{"pointer-events"}, regexp.MustCompile(`^(none|auto|visible|visiblePainted|visibleFill|visibleStroke|painted|fill|stroke|all)$`), nil},
	{[]string{"baseProfile"}, regexp.MustCompile(`^[a-zA-Z]+$`), []string{"svg"}},
	{[]string{"xmlns", "xmlns:xlink"}, regexp.MustCompile(`^https?://`), []string{"svg"}},
	{[]string{"version"}, reNumPair, []string{"svg"}},
	{[]string{"clip-path", "mask"}, reURLRef, nil},
}

func init() {
	htmlSanitizerPolicy = buildHTMLPolicy(false)
	htmlSanitizerPolicyTrusted = buildHTMLPolicy(true)
}

func buildHTMLPolicy(trusted bool) *bluemonday.Policy {
	p := bluemonday.NewPolicy()
	p.AllowElements(
		"a", "abbr", "b", "br", "canvas", "caption", "center", "cite",
		"code", "dd", "del", "details", "div", "dl", "dt", "em",
		"figcaption", "figure",
		"h1", "h2", "h3", "h4", "h5", "h6", "hr",
		"i", "img", "ins", "kbd", "label", "li",
		"math", "marquee", "media", "mediagroup", "noscript",
		"ol", "p", "pre", "source", "span", "strong",
		"sub", "summary", "sup",
		"table", "tbody", "td", "tfoot", "th", "thead", "title", "tr",
		"tt", "u", "ul", "video",
	)
	p.AllowElements(svgElements...)
	applySVGAttrRules(p)

	// Styles that can take a url() value (background, background-image,
	// clip-path, mask, list-style, cursor, ...) are intentionally excluded
	// — they can pull in third-party content.
	p.AllowStyles(
		// Text.
		"text-decoration", "color", "font-size", "font-weight", "font-style", "text-align",
		// Background (color only; no url()-bearing shorthands).
		"background-color",
		// Box model.
		"border", "border-top", "border-bottom", "border-left", "border-right",
		"border-collapse", "border-spacing", "border-color", "border-radius",
		"padding", "padding-top", "padding-bottom", "padding-left", "padding-right",
		"margin", "margin-top", "margin-bottom", "margin-left", "margin-right",
		// Sizing.
		"width", "max-width", "min-width", "height", "max-height",
		// Layout. display/position/float/top/left/right/bottom are intentionally
		// excluded: they let sanitized content overlap the host page and could be
		// used for clickjacking (e.g. a fake URL bar positioned over the real one).
		"clear", "z-index",
		"vertical-align", "line-height", "white-space",
		"overflow", "overflow-x", "overflow-y",
		// Flex.
		"flex-wrap", "align-items", "justify-content", "gap",
		// Table.
		"caption-side", "empty-cells",
		// Visual effects.
		"transform", "opacity",
		// Lists (list-style-type only; list-style / list-style-image can take url()).
		"list-style-type",
		// Print.
		"-webkit-print-color-adjust", "print-color-adjust",
	).Globally()

	p.AllowAttrs("href").OnElements("a")
	p.AllowAttrs("src", "srcset").OnElements("img", "source")
	p.AllowAttrs("alt", "title").Globally()
	// `id` must stay reachable for SVG internal refs (url(#id) in clip-path,
	// mask, gradients) but is scoped to SVG elements so it cannot collide
	// with host page IDs.
	p.AllowAttrs("id").Matching(regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)).OnElements(svgElements...)
	p.AllowAttrs("aria-hidden").Globally()
	p.RequireParseableURLs(true)
	p.AllowDataURIImages()
	p.AllowImages()
	p.AllowTables()
	p.RequireNoFollowOnLinks(false)
	p.AllowURLSchemes("mailto", "http", "https")
	if trusted {
		p.AllowStyles(trustedLayoutStyles...).Globally()
	}
	return p
}

// applySVGAttrRules registers all SVG attribute allowlists from the rules table.
func applySVGAttrRules(p *bluemonday.Policy) {
	for _, r := range svgAttrRules {
		elems := r.elements
		if elems == nil {
			elems = svgElements
		}
		if r.pattern == nil {
			// nil pattern means use bluemonday.Paragraph (permissive text match).
			p.AllowAttrs(r.attrs...).Matching(bluemonday.Paragraph).OnElements(elems...)
		} else {
			p.AllowAttrs(r.attrs...).Matching(r.pattern).OnElements(elems...)
		}
	}
}

func SanitizeHTML(h string) string {
	return htmlSanitizerPolicy.Sanitize(h)
}

// SanitizeTrustedHTML is a variant of SanitizeHTML for extractors whose input
// comes from editorially moderated sources (e.g. Wikipedia). It additionally
// permits layout CSS properties (display/position/float/top/left/right/bottom)
// that the default policy strips to prevent clickjacking overlays. Do not use
// for arbitrary third-party HTML.
func SanitizeTrustedHTML(h string) string {
	return htmlSanitizerPolicyTrusted.Sanitize(h)
}

// SanitizeText strips every HTML tag, decodes entities and trims
// surrounding whitespace, producing safe plain text suitable for
// storing in metadata or displaying verbatim.
func SanitizeText(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	s = textSanitizerPolicy.Sanitize(s)
	s = stdhtml.UnescapeString(s)
	return strings.TrimSpace(s)
}
