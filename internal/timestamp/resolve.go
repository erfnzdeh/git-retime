package timestamp

import (
	"fmt"
	"strings"
	"time"
)

// Commit holds the original and resolved timestamps for one commit.
type Commit struct {
	Hash           string
	OrigAuthorDate time.Time
	OrigCommitDate time.Time
	Subject        string
	Body           string

	// EditedRaw is the raw string from the timestamp column after editing.
	EditedRaw string
	// EditedRaw2 is the second timestamp column (--split-dates mode only).
	EditedRaw2 string

	ResolvedAuthorDate time.Time
	ResolvedCommitDate time.Time

	// NewSubject is the commit subject after editing (may be unchanged).
	NewSubject string
}

// ResolveAll resolves the EditedRaw fields of each commit into final
// timestamps. The commits slice must be in oldest-first order.
//
// The now parameter is captured once and used for all NOW references.
func ResolveAll(commits []Commit, now time.Time, splitDates bool) error {
	for i := range commits {
		c := &commits[i]

		resolved, err := resolveOne(c.EditedRaw, c.OrigAuthorDate, getPrevOriginal(commits, i), now)
		if err != nil {
			return fmt.Errorf("commit %s: author date: %w", c.Hash, err)
		}
		c.ResolvedAuthorDate = resolved

		if splitDates {
			resolved2, err := resolveOne(c.EditedRaw2, c.OrigCommitDate, getPrevOriginal(commits, i), now)
			if err != nil {
				return fmt.Errorf("commit %s: committer date: %w", c.Hash, err)
			}
			c.ResolvedCommitDate = resolved2
		} else {
			c.ResolvedCommitDate = c.ResolvedAuthorDate
		}
	}
	return nil
}

func getPrevOriginal(commits []Commit, i int) *time.Time {
	if i == 0 {
		return nil
	}
	t := commits[i-1].OrigAuthorDate
	return &t
}

func resolveOne(raw string, original time.Time, prevOriginal *time.Time, now time.Time) (time.Time, error) {
	raw = strings.TrimSpace(raw)

	if raw == "" {
		return original, nil
	}

	if strings.ToUpper(raw) == "NOW" {
		return now, nil
	}

	if strings.HasPrefix(strings.ToUpper(raw), "PREV") {
		return resolvePrev(raw, prevOriginal)
	}

	return resolveAbsoluteOrShift(raw, original)
}

func resolvePrev(raw string, prevOriginal *time.Time) (time.Time, error) {
	if prevOriginal == nil {
		return time.Time{}, fmt.Errorf("PREV used on first commit (no previous commit)")
	}

	rest := strings.TrimSpace(raw[4:])
	if rest == "" {
		return *prevOriginal, nil
	}

	if !ContainsShift(rest) {
		return time.Time{}, fmt.Errorf("invalid expression after PREV: %q", rest)
	}

	shift, err := ParseShift(rest)
	if err != nil {
		return time.Time{}, fmt.Errorf("PREV shift: %w", err)
	}

	return prevOriginal.Add(shift), nil
}

func resolveAbsoluteOrShift(raw string, original time.Time) (time.Time, error) {
	displayedOriginal := FormatLocal(original)

	// Check if the timestamp contains RR tokens — resolve them first.
	if ContainsRR(raw) {
		resolved, err := resolveWithRR(raw)
		if err != nil {
			return time.Time{}, err
		}
		raw = resolved
	}

	// Try to extract a trailing shift expression.
	tsStr, shiftExpr := splitTrailingShift(raw)
	tsStr = strings.TrimSpace(tsStr)

	if tsStr == "" && shiftExpr != "" {
		// Bare shift like "+2h" — apply to original.
		shift, err := ParseShift(shiftExpr)
		if err != nil {
			return time.Time{}, err
		}
		return original.Add(shift), nil
	}

	parsedLocal, err := ParseLocal(tsStr)
	if err != nil {
		return time.Time{}, err
	}

	if shiftExpr != "" {
		shift, err := ParseShift(shiftExpr)
		if err != nil {
			return time.Time{}, err
		}
		parsedLocal = parsedLocal.Add(shift)
	}

	// Compute the delta between the displayed original and the new displayed
	// time, then apply that delta to the real original (preserving TZ offset).
	origLocal, _ := ParseLocal(displayedOriginal)
	delta := ComputeDelta(origLocal, parsedLocal)
	return ApplyDelta(original, delta), nil
}

func resolveWithRR(raw string) (string, error) {
	// Split into date and time portions. The raw may have a trailing shift.
	tsStr, shiftExpr := splitTrailingShift(raw)
	tsStr = strings.TrimSpace(tsStr)

	parts := strings.SplitN(tsStr, " ", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("expected 'YYYY-MM-DD HH:MM:SS' with RR, got %q", tsStr)
	}

	datePart := parts[0]
	timePart := parts[1]

	if ContainsRR(datePart) {
		return "", fmt.Errorf("RR is not supported in date fields: %q", datePart)
	}

	resolvedTime, err := ResolveRR(timePart)
	if err != nil {
		return "", err
	}

	result := datePart + " " + resolvedTime
	if shiftExpr != "" {
		result += " " + shiftExpr
	}
	return result, nil
}

// splitTrailingShift separates a trailing +/- shift expression from the
// timestamp string. Example: "2026-02-23 10:00:00 +2h" -> ("2026-02-23 10:00:00", "+2h")
func splitTrailingShift(s string) (string, string) {
	s = strings.TrimSpace(s)
	// Walk backwards to find the last token.
	idx := strings.LastIndex(s, " ")
	if idx < 0 {
		if ContainsShift(s) {
			return "", s
		}
		return s, ""
	}

	lastToken := s[idx+1:]
	if ContainsShift(lastToken) {
		return s[:idx], lastToken
	}
	return s, ""
}
