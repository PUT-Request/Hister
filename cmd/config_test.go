package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteDefaultConfigFileCreatesParentDirectories(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "nested", "config", "config.yml")
	content := []byte("app:\n  search_url: https://example.com\n")

	if err := writeDefaultConfigFile(filename, content); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(content) {
		t.Fatalf("config content = %q, want %q", got, content)
	}
}
