package model_test

import (
	"testing"

	"github.com/asciimoo/hister/server/model"
	"github.com/asciimoo/hister/server/testutil"
)

func TestLegacyIndexerMetadata(t *testing.T) {
	testutil.InitModel(t)

	metadata, err := model.GetLegacyIndexerMetadata()
	if err != nil {
		t.Fatal(err)
	}
	if metadata != nil {
		t.Fatalf("legacy metadata = %#v, want nil", metadata)
	}

	if err := model.DB.Exec(`
		CREATE TABLE indexer_versions (
			version INTEGER,
			analyzer_fingerprint TEXT
		)
	`).Error; err != nil {
		t.Fatal(err)
	}
	if err := model.DB.Exec(
		"INSERT INTO indexer_versions (version, analyzer_fingerprint) VALUES (?, ?)",
		8,
		"fingerprint",
	).Error; err != nil {
		t.Fatal(err)
	}

	metadata, err = model.GetLegacyIndexerMetadata()
	if err != nil {
		t.Fatal(err)
	}
	if metadata == nil {
		t.Fatal("legacy metadata is nil")
	}
	if metadata.Version != 8 || metadata.AnalyzerFingerprint != "fingerprint" {
		t.Fatalf("legacy metadata = %#v", metadata)
	}
}
