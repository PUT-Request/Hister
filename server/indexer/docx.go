// SPDX-License-Identifier: AGPL-3.0-or-later

package indexer

import (
	"errors"
	"fmt"
	"runtime/debug"
	"strings"

	docx "github.com/mmonterroca/docxgo/v2"
	"github.com/mmonterroca/docxgo/v2/domain"
	"github.com/rs/zerolog/log"

	"github.com/asciimoo/hister/server/document"
)

type docxFileType struct{}

func (docxFileType) Match(path string) bool {
	return hasExtension(path, ".docx")
}

func (docxFileType) Prepare(d *document.Document, docxData []byte) error {
	text, title, err := extractDocxText(docxData)
	if err != nil {
		return fmt.Errorf("docx text extraction: %w", err)
	}
	if strings.TrimSpace(text) == "" {
		return errors.New("docx contains no extractable text")
	}
	d.Text = text
	if d.Title == "" {
		d.Title = title
	}
	d.AddMetadata("type", "docx")
	return nil
}

func (i *Indexer) AddDocx(d *document.Document, docxData []byte) error {
	if err := (docxFileType{}).Prepare(d, docxData); err != nil {
		return err
	}
	return i.Add(d)
}

func extractDocxText(docxData []byte) (text string, title string, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Debug().Msgf("docx parser panic: %v\n%s", r, debug.Stack())
			err = fmt.Errorf("docx parser panic: %v", r)
		}
	}()

	doc, err := docx.OpenDocumentFromBytes(docxData)
	if err != nil {
		return "", "", fmt.Errorf("open docx: %w", err)
	}

	if meta := doc.Metadata(); meta != nil {
		title = strings.TrimSpace(meta.Title)
	}

	paragraphs := doc.Paragraphs()
	textParts := make([]string, 0, len(paragraphs))
	for _, p := range paragraphs {
		paragraphText := strings.TrimSpace(p.Text())
		if paragraphText == "" {
			continue
		}
		if title == "" && isDocxTitleParagraph(p) {
			title = paragraphText
		}
		textParts = append(textParts, paragraphText)
	}
	if title == "" && len(textParts) > 0 {
		title = textParts[0]
	}
	return strings.Join(textParts, "\n\n"), title, nil
}

func isDocxTitleParagraph(p domain.Paragraph) bool {
	style := p.Style()
	if style == nil {
		return false
	}
	styleID := strings.ToLower(strings.ReplaceAll(style.ID(), " ", ""))
	styleName := strings.ToLower(strings.ReplaceAll(style.Name(), " ", ""))
	for _, titleStyle := range []string{
		strings.ToLower(domain.StyleIDTitle),
		strings.ToLower(domain.StyleIDHeading1),
		"title",
		"heading1",
	} {
		if styleID == titleStyle || styleName == titleStyle {
			return true
		}
	}
	return false
}
