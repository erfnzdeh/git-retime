package compile

import (
	"strings"
	"testing"
	"time"

	ts "github.com/erfnzdeh/git-retime/internal/timestamp"
)

func TestCompile_Basic(t *testing.T) {
	commits := []ts.Commit{
		{
			Hash:               "abc1234abcd",
			Subject:            "Fix navbar",
			NewSubject:         "Fix navbar",
			ResolvedAuthorDate: time.Date(2026, 2, 23, 10, 0, 0, 0, time.UTC),
			ResolvedCommitDate: time.Date(2026, 2, 23, 10, 0, 0, 0, time.UTC),
		},
	}

	result := Compile(commits)

	if !strings.Contains(result, "pick abc1234 Fix navbar") {
		t.Errorf("expected pick line, got:\n%s", result)
	}
	if !strings.Contains(result, "exec") {
		t.Errorf("expected exec line, got:\n%s", result)
	}
	if !strings.Contains(result, "--no-edit") {
		t.Errorf("expected --no-edit for unchanged message, got:\n%s", result)
	}
	if !strings.Contains(result, "2026-02-23T10:00:00Z") {
		t.Errorf("expected RFC3339 date, got:\n%s", result)
	}
}

func TestCompile_ChangedMessage(t *testing.T) {
	commits := []ts.Commit{
		{
			Hash:               "abc1234abcd",
			Subject:            "Fix navbar",
			NewSubject:         "Fix top navigation bar",
			Body:               "Some body text",
			ResolvedAuthorDate: time.Date(2026, 2, 23, 10, 0, 0, 0, time.UTC),
			ResolvedCommitDate: time.Date(2026, 2, 23, 10, 0, 0, 0, time.UTC),
		},
	}

	result := Compile(commits)

	if strings.Contains(result, "--no-edit") {
		t.Errorf("should NOT have --no-edit for changed message, got:\n%s", result)
	}
	if !strings.Contains(result, "-m") {
		t.Errorf("expected -m for changed message, got:\n%s", result)
	}
}

func TestCompile_MultipleCommits(t *testing.T) {
	commits := []ts.Commit{
		{
			Hash:               "abc1234abcd",
			Subject:            "First",
			NewSubject:         "First",
			ResolvedAuthorDate: time.Date(2026, 2, 23, 10, 0, 0, 0, time.UTC),
			ResolvedCommitDate: time.Date(2026, 2, 23, 10, 0, 0, 0, time.UTC),
		},
		{
			Hash:               "def5678efgh",
			Subject:            "Second",
			NewSubject:         "Second",
			ResolvedAuthorDate: time.Date(2026, 2, 23, 11, 0, 0, 0, time.UTC),
			ResolvedCommitDate: time.Date(2026, 2, 23, 11, 0, 0, 0, time.UTC),
		},
	}

	result := Compile(commits)

	lines := strings.Split(strings.TrimSpace(result), "\n")
	pickCount := 0
	execCount := 0
	for _, line := range lines {
		if strings.HasPrefix(line, "pick") {
			pickCount++
		}
		if strings.HasPrefix(line, "exec") {
			execCount++
		}
	}

	if pickCount != 2 {
		t.Errorf("expected 2 pick lines, got %d", pickCount)
	}
	if execCount != 2 {
		t.Errorf("expected 2 exec lines, got %d", execCount)
	}
}
