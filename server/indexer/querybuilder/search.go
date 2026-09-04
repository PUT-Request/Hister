package querybuilder

import (
	"strings"

	"github.com/asciimoo/hister/server/indexer/searchschema"
)

// SearchExpression separates search directives from text matched by Bleve.
type SearchExpression struct {
	Text    string
	Sort    string
	HasSort bool
}

// ParseSearch separates supported search directives while preserving the
// remaining tokens for normal query building. Invalid directives remain
// searchable text.
func ParseSearch(input string) SearchExpression {
	expression := SearchExpression{Text: input}
	tokens, err := Tokenize(input)
	if err != nil {
		return expression
	}

	runes := []rune(input)
	parts := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if sortMode, ok := sortDirective(token); ok {
			expression.Sort = sortMode
			expression.HasSort = true
			continue
		}
		parts = append(parts, string(runes[token.start:token.end]))
	}
	if !expression.HasSort {
		return expression
	}

	expression.Text = strings.Join(parts, " ")
	if strings.TrimSpace(expression.Text) == "" {
		expression.Text = "*"
	}
	return expression
}

func sortDirective(token Token) (string, bool) {
	if token.Type != TokenWord {
		return "", false
	}
	value, ok := strings.CutPrefix(token.Value, "sort:")
	if !ok {
		return "", false
	}
	definition, ok := searchschema.LookupSort(value)
	if !ok {
		return "", false
	}
	if definition.Default {
		return "", true
	}
	return definition.Value, true
}
