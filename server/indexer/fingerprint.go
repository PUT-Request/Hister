package indexer

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

type analyzerFingerprintConfig struct {
	Version         int  `json:"version"`
	DetectLanguages bool `json:"detect_languages"`
	KeepStopwords   bool `json:"keep_stopwords"`
}

// AnalyzerFingerprint identifies configuration that changes index analysis.
func AnalyzerFingerprint(detectLanguages, keepStopwords bool) string {
	config := analyzerFingerprintConfig{
		Version:         1,
		DetectLanguages: detectLanguages,
		KeepStopwords:   detectLanguages && keepStopwords,
	}
	data, err := json.Marshal(config)
	if err != nil {
		panic(err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
