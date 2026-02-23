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

func TestResolveAll_PREV(t *testing.T) {
	orig1 := time.Date(2026, 2, 23, 10, 0, 0, 0, time.Local)
	orig2 := time.Date(2026, 2, 23, 11, 0, 0, 0, time.Local)

	commits := []Commit{
		{
			Hash:           "abc1234",
			OrigAuthorDate: orig1,
			OrigCommitDate: orig1,
			EditedRaw:      FormatLocal(orig1),
			Subject:        "First",
			NewSubject:     "First",
		},
		{
			Hash:           "def5678",
			OrigAuthorDate: orig2,
			OrigCommitDate: orig2,
			EditedRaw:      "PREV +30m",
			Subject:        "Second",
			NewSubject:     "Second",
		},
	}

	err := ResolveAll(commits, time.Now(), false)
	if err != nil {
		t.Fatalf("ResolveAll: %v", err)
	}

	// PREV refers to the ORIGINAL timestamp of the line above.
	expected := orig1.Add(30 * time.Minute)
	if !commits[1].ResolvedAuthorDate.Equal(expected) {
		t.Errorf("expected PREV+30m=%v, got %v", expected, commits[1].ResolvedAuthorDate)
	}
}

func TestResolveAll_PREV_Bare(t *testing.T) {
	orig1 := time.Date(2026, 2, 23, 10, 0, 0, 0, time.Local)
	orig2 := time.Date(2026, 2, 23, 11, 0, 0, 0, time.Local)

	commits := []Commit{
		{
			Hash:           "abc1234",
			OrigAuthorDate: orig1,
			OrigCommitDate: orig1,
			EditedRaw:      FormatLocal(orig1),
			Subject:        "First",
			NewSubject:     "First",
		},
		{
			Hash:           "def5678",
			OrigAuthorDate: orig2,
			OrigCommitDate: orig2,
			EditedRaw:      "PREV",
			Subject:        "Second",
			NewSubject:     "Second",
		},
	}

	err := ResolveAll(commits, time.Now(), false)
	if err != nil {
		t.Fatalf("ResolveAll: %v", err)
	}

	if !commits[1].ResolvedAuthorDate.Equal(orig1) {
		t.Errorf("bare PREV: expected %v, got %v", orig1, commits[1].ResolvedAuthorDate)
	}
}

func TestResolveAll_PREV_FirstCommit(t *testing.T) {
	orig := time.Date(2026, 2, 23, 10, 0, 0, 0, time.Local)
	commits := []Commit{
		{
			Hash:           "abc1234",
			OrigAuthorDate: orig,
			OrigCommitDate: orig,
			EditedRaw:      "PREV",
			Subject:        "First",
			NewSubject:     "First",
		},
	}

	err := ResolveAll(commits, time.Now(), false)
	if err == nil {
		t.Fatal("expected error for PREV on first commit")
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
