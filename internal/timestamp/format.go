package timestamp

import (
	"fmt"
	"time"
)

const DisplayLayout = "2006-01-02 15:04:05"

// FormatLocal renders t in the user's local timezone without showing the offset.
func FormatLocal(t time.Time) string {
	return t.In(time.Local).Format(DisplayLayout)
}

// ParseLocal parses a timestamp string in DisplayLayout and interprets it in
// the user's local timezone. This is used when reading the todo file.
func ParseLocal(s string) (time.Time, error) {
	t, err := time.ParseInLocation(DisplayLayout, s, time.Local)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timestamp %q: %w", s, err)
	}
	return t, nil
}

// ComputeDelta returns the duration between the displayed original and the
// displayed new time. Both must already be in local timezone representation.
func ComputeDelta(displayedOriginal, displayedNew time.Time) time.Duration {
	return displayedNew.Sub(displayedOriginal)
}

// ApplyDelta adds a duration to the original timestamp (preserving its
// timezone offset) and returns the result.
func ApplyDelta(original time.Time, delta time.Duration) time.Time {
	return original.Add(delta)
}

// FormatGit returns the timestamp in RFC 3339 format suitable for
// GIT_AUTHOR_DATE / GIT_COMMITTER_DATE environment variables.
func FormatGit(t time.Time) string {
	return t.Format(time.RFC3339)
}
