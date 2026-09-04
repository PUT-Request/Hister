package ytdlp

import (
	"strings"
	"testing"
)

func TestParseVTT(t *testing.T) {
	result := parseVTT(sampleVTT)

	if !strings.Contains(result, "elephants") {
		t.Errorf("parseVTT missing 'elephants', got: %q", result)
	}
	if !strings.Contains(result, "really long trunks") {
		t.Errorf("parseVTT missing 'really long trunks', got: %q", result)
	}
	if strings.Contains(result, "<c>") {
		t.Error("parseVTT should strip VTT tags")
	}
	if strings.Contains(result, "-->") {
		t.Error("parseVTT should strip timestamps")
	}
	if strings.Contains(result, "WEBVTT") {
		t.Error("parseVTT should strip WEBVTT header")
	}
}
