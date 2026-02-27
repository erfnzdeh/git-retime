package todo

import "strings"

// HelpBlock returns the commented-out syntax cheat sheet appended to the
// bottom of the .git-retime-todo file.
func HelpBlock(splitDates bool) string {
	var b strings.Builder

	b.WriteString("# --- Syntax Reference ---\n")
	b.WriteString("#\n")

	if splitDates {
		b.WriteString("# Format: <hash>  <author-date>  <committer-date>  <message>\n")
	} else {
		b.WriteString("# Format: <hash>  <timestamp>  <message>\n")
	}

	b.WriteString("#\n")
	b.WriteString("# Timestamps are displayed in your local timezone.\n")
	b.WriteString("# Edit the timestamp column to change commit dates.\n")
	b.WriteString("# The commit message (last column) is also editable.\n")
	b.WriteString("#\n")
	b.WriteString("# Commands:\n")
	b.WriteString("#   (leave unchanged)          Keep the original timestamp\n")
	b.WriteString("#   2026-02-23 14:00:00        Set an absolute time\n")
	b.WriteString("#   2026-02-23 10:00:00 +2h    Shift from the written time\n")
	b.WriteString("#   +2h, -30m, +1d2h30m        Shift from the previous commit's new time\n")
	b.WriteString("#   NOW                        Current time (identical for all NOW commits)\n")
	b.WriteString("#   RR or RR(08,17)            Randomize a time field (HH:MM:SS only)\n")
	b.WriteString("#     e.g. 2026-02-23 RR(09,17):RR:00\n")
	b.WriteString("#\n")
	b.WriteString("# Units: w=weeks, d=days, h=hours, m=minutes, s=seconds\n")
	b.WriteString("# Compound shifts: +1d2h30m (1 day, 2 hours, 30 minutes)\n")
	b.WriteString("#\n")
	b.WriteString("# To abort: delete all lines or write ABORT on the first line.\n")
	b.WriteString("# Do not delete or reorder lines.\n")

	return b.String()
}
