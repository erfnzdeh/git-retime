package timestamp

import (
	"testing"
	"time"
)

func TestResolveAll_Unchanged(t *testing.T) {
	orig := time.Date(2026, 2, 23, 10, 0, 0, 0, time.Local)
	commits := []Commit{
		{
			Hash:           "abc1234",
			OrigAuthorDate: orig,
			OrigCommitDate: orig,
			EditedRaw:      FormatLocal(orig),
			Subject:        "Test commit",
			NewSubject:     "Test commit",
		},
	}

	err := ResolveAll(commits, time.Now(), false)
	if err != nil {
		t.Fatalf("ResolveAll: %v", err)
	}

	if !commits[0].ResolvedAuthorDate.Equal(orig) {
		t.Errorf("expected unchanged timestamp %v, got %v", orig, commits[0].ResolvedAuthorDate)
	}
}

func TestResolveAll_Shift(t *testing.T) {
	orig := time.Date(2026, 2, 23, 10, 0, 0, 0, time.Local)
	commits := []Commit{
		{
			Hash:           "abc1234",
			OrigAuthorDate: orig,
			OrigCommitDate: orig,
			EditedRaw:      FormatLocal(orig) + " +2h",
			Subject:        "Test",
			NewSubject:     "Test",
		},
	}

	err := ResolveAll(commits, time.Now(), false)
	if err != nil {
		t.Fatalf("ResolveAll: %v", err)
	}

	expected := orig.Add(2 * time.Hour)
	if !commits[0].ResolvedAuthorDate.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, commits[0].ResolvedAuthorDate)
	}
}

func TestResolveAll_BareShift_FromPrevResolved(t *testing.T) {
	orig1 := time.Date(2026, 2, 23, 10, 0, 0, 0, time.Local)
	orig2 := time.Date(2026, 2, 23, 11, 0, 0, 0, time.Local)

	// First commit is shifted to 12:00; second uses bare +30m → should be 12:30.
	commits := []Commit{
		{
			Hash:           "abc1234",
			OrigAuthorDate: orig1,
			OrigCommitDate: orig1,
			EditedRaw:      FormatLocal(orig1) + " +2h", // resolves to 12:00
			Subject:        "First",
			NewSubject:     "First",
		},
		{
			Hash:           "def5678",
			OrigAuthorDate: orig2,
			OrigCommitDate: orig2,
			EditedRaw:      "+30m",
			Subject:        "Second",
			NewSubject:     "Second",
		},
	}

	err := ResolveAll(commits, time.Now(), false)
	if err != nil {
		t.Fatalf("ResolveAll: %v", err)
	}

	expected := orig1.Add(2*time.Hour + 30*time.Minute) // 12:30
	if !commits[1].ResolvedAuthorDate.Equal(expected) {
		t.Errorf("bare shift: expected %v, got %v", expected, commits[1].ResolvedAuthorDate)
	}
}

func TestResolveAll_BareShift_Chaining(t *testing.T) {
	orig := time.Date(2026, 2, 23, 10, 0, 0, 0, time.Local)

	// Three commits: first absolute at 10:00, then +1h, then +30m → 10:00, 11:00, 11:30
	commits := []Commit{
		{
			Hash:           "aaa",
			OrigAuthorDate: orig,
			OrigCommitDate: orig,
			EditedRaw:      FormatLocal(orig),
			Subject:        "A",
			NewSubject:     "A",
		},
		{
			Hash:           "bbb",
			OrigAuthorDate: orig,
			OrigCommitDate: orig,
			EditedRaw:      "+1h",
			Subject:        "B",
			NewSubject:     "B",
		},
		{
			Hash:           "ccc",
			OrigAuthorDate: orig,
			OrigCommitDate: orig,
			EditedRaw:      "+30m",
			Subject:        "C",
			NewSubject:     "C",
		},
	}

	err := ResolveAll(commits, time.Now(), false)
	if err != nil {
		t.Fatalf("ResolveAll: %v", err)
	}

	if !commits[1].ResolvedAuthorDate.Equal(orig.Add(time.Hour)) {
		t.Errorf("B: expected 11:00, got %v", commits[1].ResolvedAuthorDate)
	}
	if !commits[2].ResolvedAuthorDate.Equal(orig.Add(time.Hour + 30*time.Minute)) {
		t.Errorf("C: expected 11:30, got %v", commits[2].ResolvedAuthorDate)
	}
}

func TestResolveAll_BareShift_FirstCommitErrors(t *testing.T) {
	orig := time.Date(2026, 2, 23, 10, 0, 0, 0, time.Local)
	commits := []Commit{
		{
			Hash:           "abc1234",
			OrigAuthorDate: orig,
			OrigCommitDate: orig,
			EditedRaw:      "+2h",
			Subject:        "First",
			NewSubject:     "First",
		},
	}

	err := ResolveAll(commits, time.Now(), false)
	if err == nil {
		t.Fatal("expected error for bare shift on first commit")
	}
}

func TestResolveAll_NOW(t *testing.T) {
	orig := time.Date(2026, 2, 23, 10, 0, 0, 0, time.Local)
	now := time.Date(2026, 2, 23, 15, 0, 0, 0, time.Local)

	commits := []Commit{
		{
			Hash:           "abc1234",
			OrigAuthorDate: orig,
			OrigCommitDate: orig,
			EditedRaw:      "NOW",
			Subject:        "Test",
			NewSubject:     "Test",
		},
	}

	err := ResolveAll(commits, now, false)
	if err != nil {
		t.Fatalf("ResolveAll: %v", err)
	}

	if !commits[0].ResolvedAuthorDate.Equal(now) {
		t.Errorf("expected NOW=%v, got %v", now, commits[0].ResolvedAuthorDate)
	}
}


func TestResolveAll_AbsoluteEdit(t *testing.T) {
	loc := time.FixedZone("IST", 5*3600+30*60)
	orig := time.Date(2026, 2, 23, 10, 0, 0, 0, loc)

	// The displayed local time of orig and the new desired time.
	displayedOrig := FormatLocal(orig)
	_ = displayedOrig

	newLocal := time.Date(2026, 2, 23, 14, 0, 0, 0, time.Local)

	commits := []Commit{
		{
			Hash:           "abc1234",
			OrigAuthorDate: orig,
			OrigCommitDate: orig,
			EditedRaw:      FormatLocal(newLocal),
			Subject:        "Test",
			NewSubject:     "Test",
		},
	}

	err := ResolveAll(commits, time.Now(), false)
	if err != nil {
		t.Fatalf("ResolveAll: %v", err)
	}

	// The original TZ offset should be preserved.
	_, offset := commits[0].ResolvedAuthorDate.Zone()
	if offset != 5*3600+30*60 {
		t.Errorf("timezone offset changed: got %d", offset)
	}
}

func TestResolveAll_Empty(t *testing.T) {
	orig := time.Date(2026, 2, 23, 10, 0, 0, 0, time.Local)
	commits := []Commit{
		{
			Hash:           "abc1234",
			OrigAuthorDate: orig,
			OrigCommitDate: orig,
			EditedRaw:      "",
			Subject:        "Test",
			NewSubject:     "Test",
		},
	}

	err := ResolveAll(commits, time.Now(), false)
	if err != nil {
		t.Fatalf("ResolveAll: %v", err)
	}

	if !commits[0].ResolvedAuthorDate.Equal(orig) {
		t.Errorf("empty should keep original, got %v", commits[0].ResolvedAuthorDate)
	}
}
