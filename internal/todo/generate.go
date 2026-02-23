package todo

import (
	"fmt"
	"strings"

	"github.com/erfnzdeh/git-retime/internal/git"
	"github.com/erfnzdeh/git-retime/internal/timestamp"
)

// Generate produces the .git-retime-todo file content from a list of commits.
// Commits must be in oldest-first order.
func Generate(commits []git.CommitInfo, base string, splitDates bool) string {
	var b strings.Builder

	if base != "" {
		fmt.Fprintf(&b, "# Retime %d commit(s) onto %s\n", len(commits), base[:minInt(7, len(base))])
	} else {
		fmt.Fprintf(&b, "# Retime %d commit(s) (root)\n", len(commits))
	}
	b.WriteString("#\n")

	for _, c := range commits {
		ts := timestamp.FormatLocal(c.AuthorDate)
		if splitDates {
			ts2 := timestamp.FormatLocal(c.CommitDate)
			fmt.Fprintf(&b, "%s  %s  %s  %s\n", c.ShortHash, ts, ts2, c.Subject)
		} else {
			fmt.Fprintf(&b, "%s  %s  %s\n", c.ShortHash, ts, c.Subject)
		}
	}

	b.WriteString("#\n")
	b.WriteString(HelpBlock(splitDates))

	return b.String()
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
