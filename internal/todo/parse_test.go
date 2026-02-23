package todo

import (
	"testing"
)

func TestParse_Normal(t *testing.T) {
	content := `# Retime 2 commits
#
abc1234  2026-02-23 10:00:00  Fix navbar
def5678  2026-02-23 11:00:00  Create models
#
# Syntax reference...
`
	entries, err := Parse(content, false)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	if entries[0].Hash != "abc1234" {
		t.Errorf("entry[0].Hash = %q, want abc1234", entries[0].Hash)
	}
	if entries[0].RawTS != "2026-02-23 10:00:00" {
		t.Errorf("entry[0].RawTS = %q", entries[0].RawTS)
	}
	if entries[0].Subject != "Fix navbar" {
		t.Errorf("entry[0].Subject = %q", entries[0].Subject)
	}

	if entries[1].Hash != "def5678" {
		t.Errorf("entry[1].Hash = %q", entries[1].Hash)
	}
	if entries[1].Subject != "Create models" {
		t.Errorf("entry[1].Subject = %q", entries[1].Subject)
	}
}

func TestParse_WithShift(t *testing.T) {
	content := `abc1234  2026-02-23 10:00:00 +2h  Fix navbar`

	entries, err := Parse(content, false)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	if entries[0].RawTS != "2026-02-23 10:00:00 +2h" {
		t.Errorf("RawTS = %q, want timestamp with shift", entries[0].RawTS)
	}
}

func TestParse_PREV(t *testing.T) {
	content := `abc1234  PREV +45m  Fix navbar`

	entries, err := Parse(content, false)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if entries[0].RawTS != "PREV +45m" {
		t.Errorf("RawTS = %q, want 'PREV +45m'", entries[0].RawTS)
	}
}

func TestParse_NOW(t *testing.T) {
	content := `abc1234  NOW  Fix navbar`

	entries, err := Parse(content, false)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if entries[0].RawTS != "NOW" {
		t.Errorf("RawTS = %q, want NOW", entries[0].RawTS)
	}
}

func TestParse_SplitDates(t *testing.T) {
	content := `abc1234  2026-02-23 10:00:00  2026-02-23 10:30:00  Fix navbar`

	entries, err := Parse(content, true)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if entries[0].RawTS != "2026-02-23 10:00:00" {
		t.Errorf("RawTS = %q", entries[0].RawTS)
	}
	if entries[0].RawTS2 != "2026-02-23 10:30:00" {
		t.Errorf("RawTS2 = %q", entries[0].RawTS2)
	}
	if entries[0].Subject != "Fix navbar" {
		t.Errorf("Subject = %q", entries[0].Subject)
	}
}

func TestParse_SkipsComments(t *testing.T) {
	content := `# Header
# Comment
abc1234  2026-02-23 10:00:00  Fix navbar
# Another comment
`
	entries, err := Parse(content, false)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
}

func TestSplitColumns(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"abc1234  2026-02-23 10:00:00  Fix navbar", 3},
		{"abc1234  PREV +45m  Fix navbar", 3},
		{"abc1234  NOW  Fix navbar", 3},
		{"abc1234  2026-02-23 10:00:00  2026-02-23 10:30:00  Fix navbar", 4},
	}

	for _, tt := range tests {
		cols := splitColumns(tt.input)
		if len(cols) != tt.want {
			t.Errorf("splitColumns(%q) = %d columns %v, want %d", tt.input, len(cols), cols, tt.want)
		}
	}
}
