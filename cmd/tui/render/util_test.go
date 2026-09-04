package render

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestTruncateLineUsesTerminalCellWidth(t *testing.T) {
	got := truncateLine("ab界z", 4)
	if got != "ab…" {
		t.Fatalf("truncateLine() = %q, want %q", got, "ab…")
	}
	if width := ansi.StringWidth(got); width != 3 {
		t.Fatalf("truncated width = %d, want 3", width)
	}
}

func TestSanitizeTerminalTextRemovesControlSequences(t *testing.T) {
	got := sanitizeTerminalText("safe\x1b[31mred\x1b[0m\x1b]52;c;secret\a\tend")
	if got != "safered    end" {
		t.Fatalf("sanitized text = %q", got)
	}
	if strings.ContainsAny(got, "\x1b\a") {
		t.Fatalf("sanitized text retained a terminal control: %q", got)
	}
}
