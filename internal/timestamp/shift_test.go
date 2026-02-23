package timestamp

import (
	"testing"
	"time"
)

func TestParseShift(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
		wantErr  bool
	}{
		{"+2h", 2 * time.Hour, false},
		{"-30m", -30 * time.Minute, false},
		{"+1d", 24 * time.Hour, false},
		{"+1w", 7 * 24 * time.Hour, false},
		{"+45s", 45 * time.Second, false},
		{"+1d2h30m", 24*time.Hour + 2*time.Hour + 30*time.Minute, false},
		{"-1w2d3h4m5s", -(7*24*time.Hour + 2*24*time.Hour + 3*time.Hour + 4*time.Minute + 5*time.Second), false},
		{"+0h", 0, false},

		// Errors.
		{"", 0, true},
		{"2h", 0, true},     // missing sign
		{"+", 0, true},      // no value
		{"+2x", 0, true},    // unknown unit
		{"+h", 0, true},     // missing number
		{"+-2h", 0, true},   // double sign
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseShift(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for %q, got %v", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.input, err)
			}
			if got != tt.expected {
				t.Errorf("ParseShift(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestContainsShift(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"+2h", true},
		{"-30m", true},
		{"+1d2h30m", true},
		{"PREV", false},
		{"NOW", false},
		{"+", false},
		{"", false},
		{"2h", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ContainsShift(tt.input); got != tt.want {
				t.Errorf("ContainsShift(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
