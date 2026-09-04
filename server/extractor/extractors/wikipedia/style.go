package wikipedia

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

// Inline style constants.
//
// All colours use semi-transparent rgba() so they adapt to both light and
// dark host backgrounds instead of forcing Wikipedia's light-mode palette.
const (
	// Semantic colour tokens (rgba, theme-neutral).
	colBorder    = "rgba(128,128,128,0.25)"
	colSurface   = "rgba(128,128,128,0.06)"
	colSurfaceHi = "rgba(128,128,128,0.12)"
	colAccent    = "rgba(100,150,220,0.18)"

	// Infobox: floated right panel.
	styleInfobox = "float: right; clear: right; width: 22em; max-width: 100%; " +
		"margin: 0 0 1em 1.5em; padding: 0; " +
		"border: 1px solid " + colBorder + "; border-collapse: collapse; " +
		"background-color: " + colSurface + "; font-size: 0.875em; line-height: 1.5"
	styleInfoboxCaption = "background-color: " + colAccent + "; padding: 0.5em; " +
		"text-align: center; font-weight: bold; font-size: 1.1em"
	styleInfoboxTH = "padding: 0.25em 0.5em; text-align: left; " +
		"vertical-align: top; font-weight: bold; width: 40%; " +
		"border-top: 1px solid " + colBorder
	styleInfoboxTD = "padding: 0.25em 0.5em; vertical-align: top; " +
		"border-top: 1px solid " + colBorder
	styleInfoboxImage = "padding: 0.4em; text-align: center; " +
		"border-top: 1px solid " + colBorder

	// Wikitable: bordered data table.
	styleWikitable = "border-collapse: collapse; border: 1px solid " + colBorder + "; " +
		"margin: 1em 0; background-color: " + colSurface + "; font-size: 0.875em"
	styleWikitableTH = "background-color: " + colSurfaceHi + "; border: 1px solid " + colBorder + "; " +
		"padding: 0.35em 0.65em; text-align: left; font-weight: bold"
	styleWikitableTD = "border: 1px solid " + colBorder + "; padding: 0.35em 0.65em; " +
		"vertical-align: top"
	styleWikitableCaption = "font-weight: bold; padding: 0.5em; " +
		"text-align: left; font-size: 1.05em"

	// Figures/thumbnails: floated with caption.
	styleThumbRight = "float: right; clear: right; margin: 0 0 0.8em 1.4em; " +
		"max-width: 220px; border: 1px solid " + colBorder + "; " +
		"background-color: " + colSurface + "; padding: 3px; overflow: hidden"
	styleThumbLeft = "float: left; clear: left; margin: 0 1.4em 0.8em 0; " +
		"max-width: 220px; border: 1px solid " + colBorder + "; " +
		"background-color: " + colSurface + "; padding: 3px; overflow: hidden"
	styleThumbCenter = "margin: 1em auto; display: table; max-width: 100%; " +
		"border: 1px solid " + colBorder + "; background-color: " + colSurface + "; padding: 3px"
	// Inline figure for clusters - no float, displayed inline so they wrap naturally.
	styleThumbInline = "display: inline-block; vertical-align: top; margin: 4px; " +
		"max-width: 180px; border: 1px solid " + colBorder + "; " +
		"background-color: " + colSurface + "; padding: 3px; overflow: hidden"
	styleFigcaption = "font-size: 0.85em; line-height: 1.4; padding: 4px"
	styleThumbImg   = "max-width: 100%; height: auto; display: block"

	// Hatnote: italic disambiguation notice.
	styleHatnote = "font-style: italic; padding-left: 1.6em; " +
		"margin-bottom: 0.5em; font-size: 0.9em"

	// Section headings.
	styleH2 = "border-bottom: 1px solid " + colBorder + "; padding-bottom: 0.25em; " +
		"margin-top: 1.5em"
	styleH3 = "margin-top: 1.2em"

	// Sidebar.
	styleSidebar = "float: right; clear: right; width: 18em; max-width: 100%; " +
		"margin: 0 0 1em 1.5em; padding: 0.5em; " +
		"border: 1px solid " + colBorder + "; background-color: " + colSurface + "; font-size: 0.85em"

	// References and notes section.
	styleReferencesWrap = "font-size: 0.8em; line-height: 1.6; " +
		"margin-top: 0.5em; padding-top: 0.5em"
	styleReferencesList = "margin: 0; padding-left: 2em; list-style-type: decimal"

	// Gallery: horizontal strip of thumbnails.
	styleGallery = "display: flex; flex-wrap: wrap; gap: 4px; margin: 1em 0; padding: 0; " +
		"list-style-type: none"
	styleGalleryBox   = "width: 155px; text-align: center"
	styleGalleryThumb = "width: 150px; height: 150px; display: flex; " +
		"align-items: center; justify-content: center; " +
		"border: 1px solid " + colBorder + "; " +
		"background-color: " + colSurface + "; padding: 3px; overflow: hidden"
	styleGalleryText = "font-size: 0.85em; line-height: 1.4; padding: 2px 4px; " +
		"max-width: 150px"

	// Pull-quote (cquote) decorative quote marks.
	styleQuoteMark = "vertical-align: top; border: none; font-size: 2em; " +
		"line-height: 0.6em; padding: 0.2em 0.3em; opacity: 0.3"
	styleQuoteBody  = "vertical-align: top; border: none; padding: 0.25em 0.5em"
	styleQuoteTable = "margin: 1em auto; border-collapse: collapse; border: none; width: auto"

	// Legend: colored key used on maps and charts.
	styleLegendColor = "display: inline-block; width: 1.2em; height: 1.2em; " +
		"margin-right: 0.4em; vertical-align: middle; border: 1px solid " + colBorder

	// CSS pie chart: we keep only the legend; the pie visual cannot render
	// without its CSS classes so it is removed.
	stylePieLegend = "margin: 0.5em 0; padding: 0"
)

// simpleStyles maps a CSS selector to the inline style to apply. These are
// processed in order by styleContent before the more complex styling passes.
var simpleStyles = []struct{ sel, style string }{
	{".hatnote", styleHatnote},
	{"h2", styleH2},
	{"h3", styleH3},
	{"table.sidebar, .sidebar", styleSidebar},
	{".mw-references-wrap", styleReferencesWrap},
	{"ol.references", styleReferencesList},
}

// styleContent injects inline styles onto key Wikipedia elements so they
// render richly after the sanitizer strips class attributes.
func styleContent(s *goquery.Selection) {
	for _, r := range simpleStyles {
		s.Find(r.sel).Each(func(_ int, el *goquery.Selection) {
			setStyle(el, r.style)
		})
	}
	styleInfoboxes(s)
	styleWikitables(s)
	styleQuotes(s)
	styleFigures(s)
	styleGalleries(s)
	styleLegends(s)
	stylePieCharts(s)
	styleImages(s)
}

func styleInfoboxes(s *goquery.Selection) {
	s.Find("table.infobox").Each(func(_ int, tbl *goquery.Selection) {
		setStyle(tbl, styleInfobox)
		tbl.Find("tr").Each(func(_ int, tr *goquery.Selection) {
			th := tr.Find("th")
			td := tr.Find("td")
			// Header row spanning full width (title).
			if th.Length() > 0 && td.Length() == 0 {
				colspan, _ := th.Attr("colspan")
				if colspan != "" {
					setStyle(th, styleInfoboxCaption)
				} else {
					setStyle(th, styleInfoboxTH)
				}
			}
			// Image row (td with colspan and images inside).
			if td.Length() > 0 && th.Length() == 0 {
				colspan, _ := td.Attr("colspan")
				if colspan != "" && td.Find("img").Length() > 0 {
					setStyle(td, styleInfoboxImage)
				} else if colspan != "" {
					setStyle(td, styleInfoboxCaption)
				}
			}
			// Normal key-value rows.
			if th.Length() > 0 && td.Length() > 0 {
				th.Each(func(_ int, cell *goquery.Selection) {
					setStyle(cell, styleInfoboxTH)
				})
				td.Each(func(_ int, cell *goquery.Selection) {
					setStyle(cell, styleInfoboxTD)
				})
			}
		})
	})
}

var wikitableChildStyles = []struct{ sel, style string }{
	{"caption", styleWikitableCaption},
	{"th", styleWikitableTH},
	{"td", styleWikitableTD},
}

func styleWikitables(s *goquery.Selection) {
	s.Find("table.wikitable, table.sortable").Each(func(_ int, tbl *goquery.Selection) {
		if tbl.HasClass("infobox") {
			return
		}
		setStyle(tbl, styleWikitable)
		applyChildStyles(tbl, wikitableChildStyles)
		tbl.WrapHtml(`<div style="overflow-x: auto; max-width: 100%"></div>`)
	})
}

// figureChildStyles maps child selectors to their styles, shared by both
// modern <figure> and legacy div.thumb paths.
var figureChildStyles = []struct{ sel, style string }{
	{"figcaption, .thumbcaption", styleFigcaption},
	{"img", styleThumbImg},
}

// alignmentStyle returns the style for a container based on its class string.
// The mapping is checked in order; the first match wins, falling back to
// styleThumbRight.
func alignmentStyle(cls string, mappings []struct{ class, style string }) string {
	for _, m := range mappings {
		if strings.Contains(cls, m.class) {
			return m.style
		}
	}
	return styleThumbRight
}

// Alignment class→style mappings for modern <figure> and legacy div.thumb.
var (
	figureAlignments = []struct{ class, style string }{
		{"mw-halign-left", styleThumbLeft},
		{"mw-halign-center", styleThumbCenter},
	}
	thumbAlignments = []struct{ class, style string }{
		{"tleft", styleThumbLeft},
		{"tcenter", styleThumbCenter},
	}
)

func styleFigures(s *goquery.Selection) {
	clustered := markFigureClusters(s)

	s.Find("figure").Each(func(_ int, fig *goquery.Selection) {
		if clustered[fig.Get(0)] {
			setStyle(fig, styleThumbInline)
		} else {
			cls, _ := fig.Attr("class")
			setStyle(fig, alignmentStyle(cls, figureAlignments))
		}
		applyChildStyles(fig, figureChildStyles)
	})

	// Legacy div.thumb thumbnails.
	s.Find("div.thumb").Each(func(_ int, div *goquery.Selection) {
		if div.ParentsFiltered("ul.gallery").Length() > 0 {
			return
		}
		cls, _ := div.Attr("class")
		setStyle(div, alignmentStyle(cls, thumbAlignments))
		applyChildStyles(div, figureChildStyles)
	})
}

// markFigureClusters finds figures that have another figure as an immediate
// sibling (no intervening block content). Returns a set of underlying
// html.Node pointers that should use inline rather than float layout.
func markFigureClusters(s *goquery.Selection) map[*html.Node]bool {
	clustered := make(map[*html.Node]bool)
	s.Find("figure").Each(func(_ int, fig *goquery.Selection) {
		next := fig.Next()
		if next.Length() > 0 && goquery.NodeName(next) == "figure" {
			clustered[fig.Get(0)] = true
			clustered[next.Get(0)] = true
		}
	})
	return clustered
}

// applyChildStyles applies a list of selector→style rules to children of parent.
func applyChildStyles(parent *goquery.Selection, rules []struct{ sel, style string }) {
	for _, r := range rules {
		parent.Find(r.sel).Each(func(_ int, el *goquery.Selection) {
			setStyle(el, r.style)
		})
	}
}

var galleryChildStyles = []struct{ sel, style string }{
	{"div.thumb", styleGalleryThumb},
	{".gallerytext", styleGalleryText},
	{"img", styleThumbImg},
}

func styleGalleries(s *goquery.Selection) {
	s.Find("ul.gallery").Each(func(_ int, ul *goquery.Selection) {
		setStyle(ul, styleGallery)
		ul.Find("li.gallerybox").Each(func(_ int, li *goquery.Selection) {
			setStyle(li, styleGalleryBox)
			applyChildStyles(li, galleryChildStyles)
		})
	})
}

// styleQuotes replaces hardcoded colours on cquote/pullquote tables with
// theme-neutral styles so the decorative quote marks work in dark mode.
func styleQuotes(s *goquery.Selection) {
	s.Find("table.cquote, table.pullquote").Each(func(_ int, tbl *goquery.Selection) {
		setStyle(tbl, styleQuoteTable)
		tbl.Find("td").Each(func(_ int, td *goquery.Selection) {
			text := strings.TrimSpace(td.Text())
			// The decorative quote-mark cells contain just " or ".
			if text == "\u201c" || text == "\u201d" || text == `"` {
				setStyle(td, styleQuoteMark)
			} else if td.Find("cite").Length() == 0 {
				setStyle(td, styleQuoteBody)
			}
		})
	})
}

func styleLegends(s *goquery.Selection) {
	s.Find(".legend-color").Each(func(_ int, span *goquery.Selection) {
		existing, _ := span.Attr("style")
		setStyle(span, styleLegendColor+"; "+existing)
	})
}

// Pie chart elements to remove (the pie visual can't render without its CSS classes).
var pieChartNoise = []string{".smooth-pie", ".smooth-pie-border"}

func stylePieCharts(s *goquery.Selection) {
	removeSelectors(s, pieChartNoise)
	s.Find(".smooth-pie-legend").Each(func(_ int, ol *goquery.Selection) {
		setStyle(ol, stylePieLegend)
	})
}

// styleImages constrains standalone images so they don't blow out the preview.
func styleImages(s *goquery.Selection) {
	s.Find("img").Each(func(_ int, img *goquery.Selection) {
		if img.ParentsFiltered("figure, .thumb").Length() > 0 {
			return
		}
		if _, ok := img.Attr("width"); ok {
			setStyle(img, "max-width: 100%; height: auto")
		}
	})
}

func setStyle(s *goquery.Selection, style string) {
	s.SetAttr("style", style)
}
