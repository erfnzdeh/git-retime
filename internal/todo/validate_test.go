package todo

import (
	"testing"

	"github.com/erfnzdeh/git-retime/internal/git"
)

func TestIsAbort_EmptyFile(t *testing.T) {
	if !IsAbort("") {
		t.Error("empty file should be abort")
	}
}

func TestIsAbort_OnlyComments(t *testing.T) {
	if !IsAbort("# just comments\n# nothing else\n") {
		t.Error("only comments should be abort")
	}
}

func TestIsAbort_AbortKeyword(t *testing.T) {
	if !IsAbort("ABORT\n") {
		t.Error("ABORT keyword should be abort")
	}
	if !IsAbort("  abort  \n") {
		t.Error("case-insensitive ABORT should work")
	}
}

func TestIsAbort_NotAbort(t *testing.T) {
	if IsAbort("abc1234  2026-02-23 10:00:00  Fix navbar\n") {
		t.Error("valid content should not be abort")
	}
}

func TestValidateStructure_Valid(t *testing.T) {
	entries := []ParsedEntry{
		{Hash: "abc1234", RawTS: "2026-02-23 10:00:00", Subject: "First"},
		{Hash: "def5678", RawTS: "2026-02-23 11:00:00", Subject: "Second"},
	}
	originals := []git.CommitInfo{
		{ShortHash: "abc1234", Subject: "First"},
		{ShortHash: "def5678", Subject: "Second"},
	}

	if err := ValidateStructure(entries, originals); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateStructure_Deleted(t *testing.T) {
	entries := []ParsedEntry{
		{Hash: "abc1234"},
	}
	originals := []git.CommitInfo{
		{ShortHash: "abc1234"},
		{ShortHash: "def5678"},
	}

	err := ValidateStructure(entries, originals)
	if err == nil {
		t.Error("expected error for deleted commit")
	}
}

func TestValidateStructure_Reordered(t *testing.T) {
	entries := []ParsedEntry{
		{Hash: "def5678"},
		{Hash: "abc1234"},
	}
	originals := []git.CommitInfo{
		{ShortHash: "abc1234"},
		{ShortHash: "def5678"},
	}

	err := ValidateStructure(entries, originals)
	if err == nil {
		t.Error("expected error for reordered commits")
	}
}

func TestValidateStructure_Extra(t *testing.T) {
	entries := []ParsedEntry{
		{Hash: "abc1234"},
		{Hash: "def5678"},
		{Hash: "ghi9012"},
	}
	originals := []git.CommitInfo{
		{ShortHash: "abc1234"},
		{ShortHash: "def5678"},
	}

	err := ValidateStructure(entries, originals)
	if err == nil {
		t.Error("expected error for extra lines")
	}
}
