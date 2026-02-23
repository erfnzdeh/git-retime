package todo

import (
	"fmt"
	"strings"

	"github.com/erfnzdeh/git-retime/internal/git"
	"github.com/erfnzdeh/git-retime/internal/timestamp"
)

// ParsedEntry is a single commit line from the edited todo file.
type ParsedEntry struct {
	Hash      string
	RawTS     string
	RawTS2    string // only in split-dates mode
	Subject   string
}

// Parse reads the edited .git-retime-todo content and returns parsed entries.
// It filters out comment lines and blank lines.
func Parse(content string, splitDates bool) ([]ParsedEntry, error) {
	lines := strings.Split(content, "\n")
	var entries []ParsedEntry

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		entry, err := parseLine(line, splitDates)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func parseLine(line string, splitDates bool) (ParsedEntry, error) {
	// The format is: <hash>  <timestamp>  <message>
	// In split-dates mode: <hash>  <author-ts>  <committer-ts>  <message>
	//
	// Columns are separated by two or more spaces. The last column (message)
	// is everything after the final double-space delimiter.

	parts := splitColumns(line)

	if splitDates {
		if len(parts) < 4 {
			return ParsedEntry{}, fmt.Errorf("expected 4 columns (split-dates mode), got %d in: %q", len(parts), line)
		}
		return ParsedEntry{
			Hash:    parts[0],
			RawTS:   parts[1],
			RawTS2:  parts[2],
			Subject: strings.Join(parts[3:], "  "),
		}, nil
	}

	if len(parts) < 3 {
		return ParsedEntry{}, fmt.Errorf("expected 3 columns, got %d in: %q", len(parts), line)
	}

	return ParsedEntry{
		Hash:    parts[0],
		RawTS:   parts[1],
		Subject: strings.Join(parts[2:], "  "),
	}, nil
}

// splitColumns splits a line by runs of 2+ spaces.
func splitColumns(line string) []string {
	var parts []string
	var current strings.Builder
	spaceCount := 0

	for _, r := range line {
		if r == ' ' {
			spaceCount++
			if spaceCount < 2 {
				current.WriteRune(r)
			}
			continue
		}

		if spaceCount >= 2 {
			// Flush the accumulated part (trim trailing single space).
			parts = append(parts, strings.TrimRight(current.String(), " "))
			current.Reset()
		} else if spaceCount == 1 {
			// Single space is part of the column content (e.g., inside a timestamp).
		}
		spaceCount = 0
		current.WriteRune(r)
	}

	if current.Len() > 0 {
		parts = append(parts, strings.TrimRight(current.String(), " "))
	}

	return parts
}

// ToCommits converts parsed entries into timestamp.Commit structs, merging
// with the original commit info for delta calculation.
func ToCommits(entries []ParsedEntry, originals []git.CommitInfo, splitDates bool) ([]timestamp.Commit, error) {
	if len(entries) != len(originals) {
		return nil, fmt.Errorf("entry count (%d) does not match original commit count (%d)", len(entries), len(originals))
	}

	commits := make([]timestamp.Commit, len(entries))
	for i, e := range entries {
		orig := originals[i]
		commits[i] = timestamp.Commit{
			Hash:           orig.Hash,
			OrigAuthorDate: orig.AuthorDate,
			OrigCommitDate: orig.CommitDate,
			Subject:        orig.Subject,
			Body:           orig.Body,
			EditedRaw:      e.RawTS,
			NewSubject:     e.Subject,
		}
		if splitDates {
			commits[i].EditedRaw2 = e.RawTS2
		}
	}

	return commits, nil
}
