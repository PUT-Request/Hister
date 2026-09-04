package wikipedia

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// writeInfoboxText extracts the infobox table and writes its key-value pairs
// as searchable text.
func writeInfoboxText(b *strings.Builder, content *goquery.Selection) {
	content.Find("table.infobox").Each(func(_ int, tbl *goquery.Selection) {
		tbl.Find("tr").Each(func(_ int, tr *goquery.Selection) {
			th := strings.TrimSpace(tr.Find("th").Text())
			td := strings.TrimSpace(tr.Find("td").Text())
			if line := joinKeyValue(th, td); line != "" {
				b.WriteString(line)
				b.WriteByte('\n')
			}
		})
		b.WriteByte('\n')
	})
}

// joinKeyValue formats an infobox row as "key: value", "key", or "value"
// depending on which parts are non-empty.
func joinKeyValue(key, value string) string {
	switch {
	case key != "" && value != "":
		return key + ": " + value
	case key != "":
		return key
	default:
		return value
	}
}

// headingTags is the set of heading elements that get extra whitespace.
var headingTags = map[string]bool{
	"h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
}

// writeArticleText walks the cleaned content and writes headings, paragraphs,
// lists, and table data as structured plain text.
func writeArticleText(b *strings.Builder, content *goquery.Selection) {
	content.Find("h2, h3, h4, h5, h6, p, li, dt, dd, table.wikitable, table.sortable").Each(func(_ int, s *goquery.Selection) {
		tag := goquery.NodeName(s)
		if tag == "table" {
			writeTableText(b, s)
			return
		}
		// Skip list items inside wikitables (already handled by writeTableText).
		if !headingTags[tag] && s.ParentsFiltered("table.wikitable, table.sortable").Length() > 0 {
			return
		}
		text := strings.TrimSpace(s.Text())
		if text == "" {
			return
		}
		if headingTags[tag] {
			b.WriteByte('\n')
		}
		b.WriteString(text)
		b.WriteByte('\n')
	})
}

// writeTableText renders a wikitable as tab-separated text with a header row.
func writeTableText(b *strings.Builder, tbl *goquery.Selection) {
	if caption := strings.TrimSpace(tbl.Find("caption").Text()); caption != "" {
		b.WriteString(caption)
		b.WriteByte('\n')
	}
	tbl.Find("tr").Each(func(_ int, tr *goquery.Selection) {
		var cells []string
		tr.Find("th, td").Each(func(_ int, cell *goquery.Selection) {
			cells = append(cells, strings.TrimSpace(cell.Text()))
		})
		if len(cells) > 0 {
			b.WriteString(strings.Join(cells, "\t"))
			b.WriteByte('\n')
		}
	})
	b.WriteByte('\n')
}
