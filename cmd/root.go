package cmd

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/erfnzdeh/git-retime/internal/compile"
	"github.com/erfnzdeh/git-retime/internal/git"
	"github.com/erfnzdeh/git-retime/internal/timestamp"
	"github.com/erfnzdeh/git-retime/internal/todo"
)

type options struct {
	shift                 string
	randomize             string
	randomizeAllowParadox bool
	splitDates            bool
	interactive           bool // no-op, accepted for UX compatibility
}

func Run(args []string) error {
	// Reorder args so flags come before the positional revision argument,
	// allowing users to write "git retime HEAD~3 --shift +2h" naturally.
	flagArgs, positional := reorderArgs(args)

	fs := flag.NewFlagSet("git-retime", flag.ContinueOnError)
	var opts options

	fs.StringVar(&opts.shift, "shift", "", "shift all commits by offset (e.g. +2h, -1d30m)")
	fs.StringVar(&opts.randomize, "randomize", "", "randomize time-of-day within range (e.g. 09:00-17:00)")
	fs.BoolVar(&opts.randomizeAllowParadox, "randomize-allow-paradox", false, "allow non-monotonic times when randomizing (by default times are sorted within each day)")
	fs.BoolVar(&opts.splitDates, "split-dates", false, "edit author and committer dates independently")
	fs.BoolVar(&opts.interactive, "i", false, "interactive mode (default, accepted for compatibility)")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: git retime [options] <revision>\n\n")
		fmt.Fprintf(os.Stderr, "Interactively edit commit timestamps.\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  git retime HEAD~5              Open editor for the last 5 commits\n")
		fmt.Fprintf(os.Stderr, "  git retime abc1234             Retime from abc1234 to HEAD\n")
		fmt.Fprintf(os.Stderr, "  git retime HEAD~3 --shift +2h  Shift last 3 commits by 2 hours\n")
		fmt.Fprintf(os.Stderr, "  git retime HEAD~5 --randomize 09:00-17:00\n")
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(flagArgs); err != nil {
		return err
	}

	// Combine any remaining args from flag parsing with our positional args.
	positional = append(positional, fs.Args()...)

	if len(positional) < 1 {
		fs.Usage()
		return errors.New("missing revision argument")
	}

	revision := positional[0]

	commits, base, needsRoot, err := git.FetchCommits(revision)
	if err != nil {
		return err
	}

	if len(commits) == 0 {
		return errors.New("no commits in the specified range")
	}

	now := time.Now()

	if opts.shift != "" {
		return runShift(commits, base, needsRoot, opts.shift, opts.splitDates, now)
	}
	if opts.randomize != "" {
		return runRandomize(commits, base, needsRoot, opts.randomize, opts.splitDates, opts.randomizeAllowParadox, now)
	}

	return runInteractive(commits, base, needsRoot, opts.splitDates, now)
}

func runInteractive(commits []git.CommitInfo, base string, needsRoot, splitDates bool, now time.Time) error {
	editor, err := git.GetEditor()
	if err != nil {
		return err
	}

	todoContent := todo.Generate(commits, base, splitDates)

	todoPath := filepath.Join(gitDir(), "git-retime-todo")

	for {
		if err := os.WriteFile(todoPath, []byte(todoContent), 0644); err != nil {
			return fmt.Errorf("writing todo file: %w", err)
		}
		defer os.Remove(todoPath)

		if err := git.OpenEditor(editor, todoPath); err != nil {
			return fmt.Errorf("editor failed: %w", err)
		}

		edited, err := os.ReadFile(todoPath)
		if err != nil {
			return fmt.Errorf("reading edited todo: %w", err)
		}

		content := string(edited)

		if todo.IsAbort(content) {
			fmt.Fprintln(os.Stderr, "retime aborted")
			return nil
		}

		entries, err := todo.Parse(content, splitDates)
		if err != nil {
			return err
		}

		if err := todo.ValidateStructure(entries, commits); err != nil {
			return err
		}

		tsCommits, err := todo.ToCommits(entries, commits, splitDates)
		if err != nil {
			return err
		}

		if err := timestamp.ResolveAll(tsCommits, now, splitDates); err != nil {
			return err
		}

		paradoxes := checkParadoxes(tsCommits)
		if len(paradoxes) > 0 {
			fmt.Fprintln(os.Stderr, "warning: time paradox detected")
			for _, p := range paradoxes {
				fmt.Fprintln(os.Stderr, "  "+p)
			}

			proceed, err := promptYesNo("Proceed anyway?")
			if err != nil {
				return err
			}
			if !proceed {
				// Re-generate the todo content to let the user fix it.
				// Keep their edits by reusing the file content.
				todoContent = content
				continue
			}
		}

		return executeRebase(tsCommits, base, needsRoot)
	}
}

func runShift(commits []git.CommitInfo, base string, needsRoot bool, shiftExpr string, splitDates bool, now time.Time) error {
	shift, err := timestamp.ParseShift(shiftExpr)
	if err != nil {
		return fmt.Errorf("invalid --shift value: %w", err)
	}

	tsCommits := make([]timestamp.Commit, len(commits))
	for i, c := range commits {
		tsCommits[i] = timestamp.Commit{
			Hash:               c.Hash,
			OrigAuthorDate:     c.AuthorDate,
			OrigCommitDate:     c.CommitDate,
			Subject:            c.Subject,
			Body:               c.Body,
			NewSubject:         c.Subject,
			ResolvedAuthorDate: c.AuthorDate.Add(shift),
			ResolvedCommitDate: c.CommitDate.Add(shift),
		}
	}

	return executeRebase(tsCommits, base, needsRoot)
}

func runRandomize(commits []git.CommitInfo, base string, needsRoot bool, rangeExpr string, splitDates, allowParadox bool, now time.Time) error {
	parts := strings.SplitN(rangeExpr, "-", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid --randomize range: expected HH:MM-HH:MM, got %q", rangeExpr)
	}

	startTime, err := parseTimeOfDay(parts[0])
	if err != nil {
		return fmt.Errorf("invalid randomize start: %w", err)
	}
	endTime, err := parseTimeOfDay(parts[1])
	if err != nil {
		return fmt.Errorf("invalid randomize end: %w", err)
	}

	if endTime <= startTime {
		return fmt.Errorf("randomize end time must be after start time")
	}

	times := make([]time.Time, len(commits))
	for i, c := range commits {
		times[i] = randomizeTime(c.AuthorDate, startTime, endTime)
	}

	if !allowParadox {
		sortTimesWithinDays(commits, times)
	}

	tsCommits := make([]timestamp.Commit, len(commits))
	for i, c := range commits {
		tsCommits[i] = timestamp.Commit{
			Hash:               c.Hash,
			OrigAuthorDate:     c.AuthorDate,
			OrigCommitDate:     c.CommitDate,
			Subject:            c.Subject,
			Body:               c.Body,
			NewSubject:         c.Subject,
			ResolvedAuthorDate: times[i],
			ResolvedCommitDate: times[i],
		}
	}

	return executeRebase(tsCommits, base, needsRoot)
}

// sortTimesWithinDays sorts the randomized times in-place, but only within
// commits that share the same calendar date. Commits on different dates are
// left independent — a later commit on an earlier date is fine by design.
func sortTimesWithinDays(commits []git.CommitInfo, times []time.Time) {
	type dateKey struct {
		y int
		m time.Month
		d int
	}

	// Collect indices per date, preserving commit order within each group.
	groups := make(map[dateKey][]int)
	for i, c := range commits {
		y, m, d := c.AuthorDate.Date()
		k := dateKey{y, m, d}
		groups[k] = append(groups[k], i)
	}

	for _, indices := range groups {
		// Extract times for this date group.
		groupTimes := make([]time.Time, len(indices))
		for j, idx := range indices {
			groupTimes[j] = times[idx]
		}
		// Sort ascending.
		sort.Slice(groupTimes, func(a, b int) bool {
			return groupTimes[a].Before(groupTimes[b])
		})
		// Write back in the same positional order (indices are already ascending).
		for j, idx := range indices {
			times[idx] = groupTimes[j]
		}
	}
}

func executeRebase(tsCommits []timestamp.Commit, base string, needsRoot bool) error {
	compiled := compile.Compile(tsCommits)

	tmpFile, err := os.CreateTemp("", "git-retime-rebase-*.todo")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(compiled); err != nil {
		tmpFile.Close()
		return fmt.Errorf("writing compiled todo: %w", err)
	}
	tmpFile.Close()

	return git.ExecuteRebase(tmpFile.Name(), base, needsRoot)
}

func checkParadoxes(commits []timestamp.Commit) []string {
	var warnings []string
	for i := 1; i < len(commits); i++ {
		prev := commits[i-1]
		curr := commits[i]
		if curr.ResolvedAuthorDate.Before(prev.ResolvedAuthorDate) {
			warnings = append(warnings, fmt.Sprintf(
				"%s (%s) is older than %s (%s)",
				curr.Hash[:minInt(7, len(curr.Hash))],
				timestamp.FormatLocal(curr.ResolvedAuthorDate),
				prev.Hash[:minInt(7, len(prev.Hash))],
				timestamp.FormatLocal(prev.ResolvedAuthorDate),
			))
		}
	}
	return warnings
}

func promptYesNo(question string) (bool, error) {
	fmt.Fprintf(os.Stderr, "%s [y/N] ", question)
	reader := bufio.NewReader(os.Stdin)
	answer, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "y" || answer == "yes", nil
}

func gitDir() string {
	// Rely on GIT_DIR or default to .git in current directory.
	if dir := os.Getenv("GIT_DIR"); dir != "" {
		return dir
	}
	return ".git"
}

// parseTimeOfDay parses "HH:MM" into total seconds since midnight.
func parseTimeOfDay(s string) (int, error) {
	s = strings.TrimSpace(s)
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return 0, fmt.Errorf("expected HH:MM format, got %q", s)
	}
	var h, m int
	if _, err := fmt.Sscanf(parts[0], "%d", &h); err != nil {
		return 0, err
	}
	if _, err := fmt.Sscanf(parts[1], "%d", &m); err != nil {
		return 0, err
	}
	if h < 0 || h > 23 || m < 0 || m > 59 {
		return 0, fmt.Errorf("time out of range: %s", s)
	}
	return h*3600 + m*60, nil
}

func randomizeTime(original time.Time, startSec, endSec int) time.Time {
	y, mo, d := original.Date()
	loc := original.Location()

	randomSec := startSec + rand.IntN(endSec-startSec)
	h := randomSec / 3600
	m := (randomSec % 3600) / 60
	sec := randomSec % 60

	return time.Date(y, mo, d, h, m, sec, 0, loc)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// reorderArgs separates flag arguments from positional arguments so that
// flags can appear anywhere in the command line. Flags that take values
// (--shift, --randomize) consume the next argument as their value.
func reorderArgs(args []string) (flagArgs, positional []string) {
	valueFlagSet := map[string]bool{
		"--shift": true, "-shift": true,
		"--randomize": true, "-randomize": true,
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			flagArgs = append(flagArgs, arg)
			if valueFlagSet[arg] && i+1 < len(args) {
				i++
				flagArgs = append(flagArgs, args[i])
			}
		} else {
			positional = append(positional, arg)
		}
	}
	return
}
