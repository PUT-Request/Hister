package cmd

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func newIndexTestCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "index [URL...]"}
	cmd.Flags().String("job-id", "", "")
	cmd.Flags().String("input", "", "")
	cmd.Flags().String("url-list", "", "")
	return cmd
}

func TestValidateIndexArgs(t *testing.T) {
	tests := []struct {
		name        string
		jobID       string
		input       string
		legacyInput string
		args        []string
		wantErr     bool
		wantErrText string
	}{
		{name: "URL", args: []string{"https://example.com"}},
		{name: "job ID without URL", jobID: "docs-crawl"},
		{name: "input without URL", input: "urls.txt"},
		{name: "legacy input without URL", legacyInput: "urls.txt"},
		{name: "neither job ID, input, nor URL", wantErr: true},
		{name: "job ID and input", jobID: "docs-crawl", input: "urls.txt", wantErr: true, wantErrText: "--job-id and --input"},
		{name: "input and legacy input", input: "urls.txt", legacyInput: "old.txt", wantErr: true, wantErrText: "--input and --url-list"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newIndexTestCommand()
			if err := cmd.Flags().Set("job-id", tt.jobID); err != nil {
				t.Fatalf("set job-id flag: %v", err)
			}
			if err := cmd.Flags().Set("input", tt.input); err != nil {
				t.Fatalf("set input flag: %v", err)
			}
			if err := cmd.Flags().Set("url-list", tt.legacyInput); err != nil {
				t.Fatalf("set url-list flag: %v", err)
			}

			err := validateIndexArgs(cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateIndexArgs() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.wantErrText != "" && !strings.Contains(err.Error(), tt.wantErrText) {
				t.Fatalf("validateIndexArgs() error = %q, want text %q", err, tt.wantErrText)
			}
		})
	}
}

func TestResolveIndexURLsPrefersInputFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "urls.txt")
	if err := os.WriteFile(path, []byte(" https://example.com \r\n\nhttps://example.org\n"), 0o600); err != nil {
		t.Fatalf("write input: %v", err)
	}
	cmd := newIndexTestCommand()
	if err := cmd.Flags().Set("input", path); err != nil {
		t.Fatalf("set input flag: %v", err)
	}
	want := []string{"https://example.com", "https://example.org"}

	got, err := resolveIndexURLs(cmd, []string{"https://ignored.example"})
	if err != nil {
		t.Fatalf("resolveIndexURLs() error: %v", err)
	}
	if !slices.Equal(got, want) {
		t.Fatalf("resolveIndexURLs() = %q, want %q", got, want)
	}
}

func TestResolveIndexURLsReadsStandardInput(t *testing.T) {
	cmd := newIndexTestCommand()
	if err := cmd.Flags().Set("input", "-"); err != nil {
		t.Fatalf("set input flag: %v", err)
	}
	cmd.SetIn(strings.NewReader("https://example.com\n\n https://example.org \n"))

	got, err := resolveIndexURLs(cmd, nil)
	if err != nil {
		t.Fatalf("resolveIndexURLs() error: %v", err)
	}
	want := []string{"https://example.com", "https://example.org"}
	if !slices.Equal(got, want) {
		t.Fatalf("resolveIndexURLs() = %q, want %q", got, want)
	}
}

func TestResolveIndexURLsSupportsLegacyFlag(t *testing.T) {
	path := filepath.Join(t.TempDir(), "urls.txt")
	if err := os.WriteFile(path, []byte("https://example.com\n"), 0o600); err != nil {
		t.Fatalf("write input: %v", err)
	}
	cmd := newIndexTestCommand()
	if err := cmd.Flags().Set("url-list", path); err != nil {
		t.Fatalf("set url-list flag: %v", err)
	}

	got, err := resolveIndexURLs(cmd, nil)
	if err != nil {
		t.Fatalf("resolveIndexURLs() error: %v", err)
	}
	if want := []string{"https://example.com"}; !slices.Equal(got, want) {
		t.Fatalf("resolveIndexURLs() = %q, want %q", got, want)
	}
}

func TestResolveIndexURLsRejectsEmptyInput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "urls.txt")
	if err := os.WriteFile(path, []byte("\n \t\n"), 0o600); err != nil {
		t.Fatalf("write input: %v", err)
	}
	cmd := newIndexTestCommand()
	if err := cmd.Flags().Set("input", path); err != nil {
		t.Fatalf("set input flag: %v", err)
	}

	if _, err := resolveIndexURLs(cmd, nil); err == nil {
		t.Fatal("resolveIndexURLs() expected an error for empty input")
	}
}

func TestResolveIndexURLsRejectsEmptyStandardInput(t *testing.T) {
	cmd := newIndexTestCommand()
	if err := cmd.Flags().Set("input", "-"); err != nil {
		t.Fatalf("set input flag: %v", err)
	}
	cmd.SetIn(strings.NewReader("\n \t\n"))

	if _, err := resolveIndexURLs(cmd, nil); err == nil || err.Error() != "standard input contains no URLs" {
		t.Fatalf("resolveIndexURLs() error = %v, want empty standard input error", err)
	}
}

func TestIndexInputJobName(t *testing.T) {
	if got, want := indexInputJobName(filepath.Join("tmp", "documentation-urls.txt")), "documentation-urls.txt"; got != want {
		t.Fatalf("indexInputJobName(file) = %q, want %q", got, want)
	}
	if got, want := indexInputJobName("-"), "stdin"; got != want {
		t.Fatalf("indexInputJobName(stdin) = %q, want %q", got, want)
	}
}

func TestIndexInputFlags(t *testing.T) {
	if flag := indexCmd.Flags().Lookup("input"); flag == nil || flag.Hidden {
		t.Fatal("--input should be visible")
	}
	legacy := indexCmd.Flags().Lookup("url-list")
	if legacy == nil || !legacy.Hidden || legacy.Deprecated == "" {
		t.Fatal("--url-list should remain as a hidden deprecated alias")
	}
}
