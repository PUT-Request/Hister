package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/asciimoo/hister/cmd/tui"
	"github.com/asciimoo/hister/server/document"
	"github.com/asciimoo/hister/server/indexer"

	"github.com/spf13/cobra"
)

func searchDocToMap(d *document.Document) map[string]any {
	return map[string]any{
		"id":          d.ID(),
		"url":         d.URL,
		"title":       d.Title,
		"domain":      d.Domain,
		"score":       d.Score,
		"added":       d.Added,
		"updated":     d.Updated,
		"language":    d.Language,
		"type":        d.Type,
		"text":        d.Text,
		"favicon":     d.Favicon,
		"favicon_key": d.FaviconKey,
		"user_id":     d.UserID,
		"html":        d.HTML,
	}
}

// searchFilterMap returns only the requested keys; returns the full map when fields is empty.
func searchFilterMap(m map[string]any, fields []string) map[string]any {
	if len(fields) == 0 {
		return m
	}
	out := make(map[string]any, len(fields))
	for _, f := range fields {
		out[f] = m[f]
	}
	return out
}

var searchCmd = &cobra.Command{
	Use:   "search [search terms]",
	Short: "Search indexed documents",
	Long:  "Search indexed documents using the running server.\nRun without search terms to open the terminal interface, or provide terms to print results to standard output.",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			if err := tui.SearchTUI(cfg); err != nil {
				exit(1, err.Error())
			}
			return
		}
		qs := strings.Join(args, " ")
		format, _ := cmd.Flags().GetString("format")
		limit, _ := cmd.Flags().GetInt("limit")
		sortMode, _ := cmd.Flags().GetString("sort")
		switch sortMode {
		case "relevance":
			sortMode = ""
		case "date", "domain", "visits":
		default:
			exit(1, "Unknown sort order: "+sortMode+" (valid values: relevance, date, domain, visits)")
		}

		// Parse and validate --fields.
		var fields []string
		includeHTML := false
		if fieldsRaw, _ := cmd.Flags().GetString("fields"); fieldsRaw != "" {
			validFields := map[string]bool{
				"id": true, "url": true, "title": true, "domain": true, "score": true,
				"added": true, "updated": true, "language": true, "type": true, "text": true,
				"favicon": true, "favicon_key": true, "user_id": true, "html": true,
			}
			for f := range strings.SplitSeq(fieldsRaw, ",") {
				f = strings.TrimSpace(f)
				if f == "" {
					continue
				}
				if !validFields[f] {
					exit(1, "Unknown field: "+f+" (valid fields: id, url, title, domain, score, added, updated, language, type, text, favicon, favicon_key, user_id, html)")
				}
				fields = append(fields, f)
				if f == "html" {
					includeHTML = true
				}
			}
		}

		// CSV column order: use --fields if given, else a sensible default.
		csvFields := fields
		if format == "csv" && len(csvFields) == 0 {
			csvFields = []string{"title", "url", "domain", "score", "added", "updated", "language", "text"}
		}

		// printDoc emits a single document in the requested format.
		var csvWriter *csv.Writer
		printDoc := func(d *document.Document) {
			m := searchFilterMap(searchDocToMap(d), fields)
			switch format {
			case "json":
				b, err := json.Marshal(m)
				if err != nil {
					exit(1, "Failed to encode JSON: "+err.Error())
				}
				fmt.Printf("%s,\n", b)
			case "csv":
				row := make([]string, 0, len(csvFields))
				for _, f := range csvFields {
					row = append(row, fmt.Sprintf("%v", m[f]))
				}
				if err := csvWriter.Write(row); err != nil {
					exit(1, "Failed to write CSV row: "+err.Error())
				}
			default:
				if len(fields) == 0 {
					fmt.Printf("%s\n%s\n\n", d.Title, d.URL)
				} else {
					parts := make([]string, 0, len(fields))
					for _, f := range fields {
						parts = append(parts, fmt.Sprintf("%v", m[f]))
					}
					fmt.Println(strings.Join(parts, "\n"))
					if len(fields) > 1 {
						fmt.Println()
					}
				}
			}
		}

		// Format-specific initialisation.
		switch format {
		case "json":
			fmt.Println("[")
		case "csv":
			csvWriter = csv.NewWriter(os.Stdout)
			if err := csvWriter.Write(csvFields); err != nil {
				exit(1, "Failed to write CSV header: "+err.Error())
			}
		}

		// Page through all results, streaming output directly.
		c := newClient()
		var (
			pageKey string
			total   int
			done    bool
		)
		for !done {
			res, err := c.Search(&indexer.Query{Text: qs, IncludeHTML: includeHTML, PageKey: pageKey, Sort: sortMode})
			if err != nil {
				exit(1, "Search failed: "+err.Error())
			}
			for _, d := range res.Documents {
				printDoc(d)
				total++
				if limit > 0 && total >= limit {
					done = true
					break
				}
			}
			if res.PageKey == "" || len(res.Documents) == 0 {
				done = true
			}
			pageKey = res.PageKey
		}

		// Format-specific teardown.
		switch format {
		case "json":
			fmt.Println("]")
		case "csv":
			csvWriter.Flush()
			if err := csvWriter.Error(); err != nil {
				exit(1, "Failed to write CSV: "+err.Error())
			}
		}
	},
}
