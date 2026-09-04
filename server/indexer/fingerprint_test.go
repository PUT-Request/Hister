package indexer

import "testing"

func TestAnalyzerFingerprint(t *testing.T) {
	base := AnalyzerFingerprint(true, false)
	if base == AnalyzerFingerprint(false, false) {
		t.Fatal("language detection must affect the analyzer fingerprint")
	}
	if base == AnalyzerFingerprint(true, true) {
		t.Fatal("keeping stopwords must affect language analyzer fingerprints")
	}
	if AnalyzerFingerprint(false, false) != AnalyzerFingerprint(false, true) {
		t.Fatal("keeping stopwords is irrelevant when language detection is disabled")
	}
	if base != AnalyzerFingerprint(true, false) {
		t.Fatal("analyzer fingerprint must be deterministic")
	}
}
