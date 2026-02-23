package todo

import (
	"fmt"
	"strings"

	"github.com/erfnzdeh/git-retime/internal/git"
)

// IsAbort returns true if the content signals an abort: either empty
// (no commit lines) or ABORT on the first non-comment line.
func IsAbort(content string) bool {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.ToUpper(line) == "ABORT" {
			return true
		}
		// Found a non-comment, non-abort line.
		return false
	}
	// No commit lines at all → abort.
	return true
}

// ValidateStructure checks that the parsed entries have the same hashes in the
// same order as the original commits. Returns an error if lines were deleted
// or reordered.
func ValidateStructure(entries []ParsedEntry, originals []git.CommitInfo) error {
	if len(entries) < len(originals) {
		missing := findMissing(entries, originals)
		return fmt.Errorf("commit(s) deleted from todo file: %s\ngit-retime only modifies timestamps — do not remove lines", strings.Join(missing, ", "))
	}

	if len(entries) > len(originals) {
		return fmt.Errorf("extra lines in todo file: expected %d commits, found %d", len(originals), len(entries))
	}

	for i, e := range entries {
		origShort := originals[i].ShortHash
		if e.Hash != origShort {
			return fmt.Errorf("commit order changed at line %d: expected %s, got %s\ngit-retime only modifies timestamps — do not reorder lines", i+1, origShort, e.Hash)
		}
	}

	return nil
}

func findMissing(entries []ParsedEntry, originals []git.CommitInfo) []string {
	present := make(map[string]bool, len(entries))
	for _, e := range entries {
		present[e.Hash] = true
	}

	var missing []string
	for _, o := range originals {
		if !present[o.ShortHash] {
			missing = append(missing, o.ShortHash)
		}
	}
	return missing
}
