package indexer

import (
	"fmt"

	"github.com/blevesearch/bleve/v2/analysis"
	"github.com/blevesearch/bleve/v2/analysis/token/stop"
	"github.com/blevesearch/bleve/v2/registry"
)

const keepStopwordsAnalyzerType = "hister_keep_stopwords"

func init() {
	if err := registry.RegisterAnalyzer(keepStopwordsAnalyzerType, newKeepStopwordsAnalyzer); err != nil {
		panic(err)
	}
}

func newKeepStopwordsAnalyzer(config map[string]any, cache *registry.Cache) (analysis.Analyzer, error) {
	language, ok := config["language"].(string)
	if !ok || language == "" {
		return nil, fmt.Errorf("keep stopwords analyzer requires a language")
	}
	base, err := cache.AnalyzerNamed(language)
	if err != nil {
		return nil, err
	}
	baseAnalyzer, ok := base.(*analysis.DefaultAnalyzer)
	if !ok {
		return nil, fmt.Errorf("language analyzer %q has unsupported type %T", language, base)
	}

	filters := make([]analysis.TokenFilter, 0, len(baseAnalyzer.TokenFilters))
	for _, filter := range baseAnalyzer.TokenFilters {
		if _, isStopFilter := filter.(*stop.StopTokensFilter); isStopFilter {
			continue
		}
		filters = append(filters, filter)
	}
	return &analysis.DefaultAnalyzer{
		CharFilters:  baseAnalyzer.CharFilters,
		Tokenizer:    baseAnalyzer.Tokenizer,
		TokenFilters: filters,
	}, nil
}
