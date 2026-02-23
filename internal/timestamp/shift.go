package timestamp

import (
	"fmt"
	"strings"
	"time"
	"unicode"
)

// ParseShift parses a compound shift expression like "+1d2h30m" or "-45s"
// into a time.Duration. Supported units: w (weeks), d (days), h (hours),
// m (minutes), s (seconds).
func ParseShift(expr string) (time.Duration, error) {
	if len(expr) == 0 {
		return 0, fmt.Errorf("empty shift expression")
	}

	sign := time.Duration(1)
	s := expr
	switch s[0] {
	case '+':
		s = s[1:]
	case '-':
		sign = -1
		s = s[1:]
	default:
		return 0, fmt.Errorf("shift must start with + or -: %q", expr)
	}

	if len(s) == 0 {
		return 0, fmt.Errorf("shift has no value: %q", expr)
	}

	var total time.Duration
	for len(s) > 0 {
		// Consume digits.
		i := 0
		for i < len(s) && unicode.IsDigit(rune(s[i])) {
			i++
		}
		if i == 0 {
			return 0, fmt.Errorf("expected number in shift %q at %q", expr, s)
		}
		var n int
		for _, c := range s[:i] {
			n = n*10 + int(c-'0')
		}

		if i >= len(s) {
			return 0, fmt.Errorf("missing unit in shift %q", expr)
		}
		unit := s[i]
		s = s[i+1:]

		switch unit {
		case 'w':
			total += time.Duration(n) * 7 * 24 * time.Hour
		case 'd':
			total += time.Duration(n) * 24 * time.Hour
		case 'h':
			total += time.Duration(n) * time.Hour
		case 'm':
			total += time.Duration(n) * time.Minute
		case 's':
			total += time.Duration(n) * time.Second
		default:
			return 0, fmt.Errorf("unknown unit %q in shift %q", string(unit), expr)
		}
	}

	return sign * total, nil
}

// ContainsShift returns true if the token looks like a shift expression
// (starts with + or - followed by a digit).
func ContainsShift(s string) bool {
	s = strings.TrimSpace(s)
	if len(s) < 2 {
		return false
	}
	return (s[0] == '+' || s[0] == '-') && unicode.IsDigit(rune(s[1]))
}
