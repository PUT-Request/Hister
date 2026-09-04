package ytdlp

import "testing"

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		seconds float64
		want    string
	}{
		{0, "0:00"},
		{19, "0:19"},
		{65, "1:05"},
		{3661, "1:01:01"},
		{7200, "2:00:00"},
	}
	for _, tt := range tests {
		if got := formatDuration(tt.seconds); got != tt.want {
			t.Errorf("formatDuration(%v) = %q, want %q", tt.seconds, got, tt.want)
		}
	}
}

func TestFormatDate(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"20050424", "2005-04-24"},
		{"20230415", "2023-04-15"},
		{"short", "short"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := formatDate(tt.input); got != tt.want {
			t.Errorf("formatDate(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
