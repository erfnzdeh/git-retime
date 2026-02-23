package timestamp

import (
	"fmt"
	"math/rand/v2"
	"regexp"
	"strconv"
	"strings"
)

var rrPattern = regexp.MustCompile(`RR(?:\((\d+),(\d+)\))?`)

// field position context for bare RR defaults
type fieldKind int

const (
	fieldHour fieldKind = iota
	fieldMinute
	fieldSecond
)

func defaultRange(f fieldKind) (int, int) {
	switch f {
	case fieldHour:
		return 0, 23
	case fieldMinute:
		return 0, 59
	case fieldSecond:
		return 0, 59
	}
	return 0, 59
}

// ResolveRR replaces all RR(...) and bare RR tokens in a time-of-day string
// (HH:MM:SS). The input must be exactly the time portion; the date portion
// must be handled separately.
//
// Examples:
//
//	"RR:RR:00"          -> "14:37:00"
//	"RR(08,17):RR:00"   -> "12:45:00"
func ResolveRR(timeStr string) (string, error) {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 3 {
		return "", fmt.Errorf("expected HH:MM:SS format, got %q", timeStr)
	}

	kinds := []fieldKind{fieldHour, fieldMinute, fieldSecond}
	for i, part := range parts {
		resolved, err := resolveField(part, kinds[i])
		if err != nil {
			return "", err
		}
		parts[i] = resolved
	}
	return strings.Join(parts, ":"), nil
}

func resolveField(field string, kind fieldKind) (string, error) {
	loc := rrPattern.FindStringIndex(field)
	if loc == nil {
		return field, nil
	}

	matches := rrPattern.FindStringSubmatch(field)
	var lo, hi int
	if matches[1] == "" {
		lo, hi = defaultRange(kind)
	} else {
		var err error
		lo, err = strconv.Atoi(matches[1])
		if err != nil {
			return "", fmt.Errorf("invalid RR min: %w", err)
		}
		hi, err = strconv.Atoi(matches[2])
		if err != nil {
			return "", fmt.Errorf("invalid RR max: %w", err)
		}
	}

	if lo > hi {
		return "", fmt.Errorf("RR min (%d) > max (%d)", lo, hi)
	}

	val := lo + rand.IntN(hi-lo+1)
	replacement := fmt.Sprintf("%02d", val)
	return field[:loc[0]] + replacement + field[loc[1]:], nil
}

// ContainsRR returns true if the string contains an RR token.
func ContainsRR(s string) bool {
	return rrPattern.MatchString(s)
}
