package timestamp

import (
	"testing"
	"time"
)

func TestFormatAndParseLocal(t *testing.T) {
	original := time.Date(2026, 2, 23, 10, 30, 0, 0, time.UTC)
	formatted := FormatLocal(original)

	parsed, err := ParseLocal(formatted)
	if err != nil {
		t.Fatalf("ParseLocal(%q): %v", formatted, err)
	}

	// After round-tripping through local TZ, the wall clock should match.
	if parsed.Format(DisplayLayout) != formatted {
		t.Errorf("round-trip mismatch: formatted=%q, reparsed=%q", formatted, parsed.Format(DisplayLayout))
	}
}

func TestComputeDelta(t *testing.T) {
	a := time.Date(2026, 2, 23, 10, 0, 0, 0, time.Local)
	b := time.Date(2026, 2, 23, 12, 30, 0, 0, time.Local)

	delta := ComputeDelta(a, b)
	if delta != 2*time.Hour+30*time.Minute {
		t.Errorf("delta = %v, want 2h30m", delta)
	}
}

func TestApplyDelta(t *testing.T) {
	loc := time.FixedZone("IST", 5*3600+30*60)
	original := time.Date(2026, 2, 23, 10, 0, 0, 0, loc)

	result := ApplyDelta(original, 2*time.Hour)
	if result.Hour() != 12 {
		t.Errorf("expected hour 12, got %d", result.Hour())
	}
	// Timezone should be preserved.
	_, offset := result.Zone()
	if offset != 5*3600+30*60 {
		t.Errorf("timezone offset changed: got %d", offset)
	}
}

func TestFormatGit(t *testing.T) {
	ts := time.Date(2026, 2, 23, 10, 0, 0, 0, time.UTC)
	got := FormatGit(ts)
	want := "2026-02-23T10:00:00Z"
	if got != want {
		t.Errorf("FormatGit = %q, want %q", got, want)
	}
}

func TestParseLocalInvalid(t *testing.T) {
	_, err := ParseLocal("not-a-timestamp")
	if err == nil {
		t.Error("expected error for invalid timestamp")
	}
}
