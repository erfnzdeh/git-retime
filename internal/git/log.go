package git

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type CommitInfo struct {
	Hash       string
	ShortHash  string
	AuthorDate time.Time
	CommitDate time.Time
	Subject    string
	Body       string
}

const fieldSep = "\x1f"
const recordSep = "\x1e"

// ResolveRevision uses git rev-parse to resolve an arbitrary revision
// expression to a full commit hash.
func ResolveRevision(rev string) (string, error) {
	out, err := exec.Command("git", "rev-parse", rev).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("cannot resolve revision %q: %s\n%s", rev, err, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

// FetchCommits resolves a revision and fetches the commits to retime.
//
// Semantics match git rebase -i: the revision is the base (exclusive).
// Commits after it up to HEAD are included. If the revision is the root
// commit (no parent), it is also included and needsRoot is set so the
// rebase uses --root.
func FetchCommits(revision string) (commits []CommitInfo, base string, needsRoot bool, err error) {
	resolved, err := ResolveRevision(revision)
	if err != nil {
		return nil, "", false, err
	}

	// Fetch commits after revision up to HEAD.
	afterCommits, err := fetchLog(resolved + "..HEAD")
	if err != nil {
		return nil, "", false, err
	}

	// Check whether the resolved revision itself is the root commit.
	_, parentErr := exec.Command("git", "rev-parse", "--verify", "--quiet", resolved+"^").CombinedOutput()
	if parentErr != nil {
		// Revision is the root commit — include it and use --root for rebase.
		rootCommits, err := fetchLog("-1 " + resolved)
		if err != nil {
			return nil, "", false, fmt.Errorf("fetching root commit: %w", err)
		}
		commits = append(rootCommits, afterCommits...)
		return commits, "", true, nil
	}

	return afterCommits, resolved, false, nil
}

func fetchLog(rangeExpr string) ([]CommitInfo, error) {
	format := strings.Join([]string{"%H", "%h", "%aI", "%cI", "%s", "%b"}, fieldSep) + recordSep

	args := []string{"log", "--format=" + format, "--reverse"}
	args = append(args, strings.Fields(rangeExpr)...)

	out, err := exec.Command("git", args...).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git log failed: %s\n%s", err, strings.TrimSpace(string(out)))
	}

	return parseLogOutput(string(out))
}

func parseLogOutput(output string) ([]CommitInfo, error) {
	records := strings.Split(output, recordSep)
	var commits []CommitInfo

	for _, rec := range records {
		rec = strings.TrimSpace(rec)
		if rec == "" {
			continue
		}

		fields := strings.SplitN(rec, fieldSep, 6)
		if len(fields) < 5 {
			return nil, fmt.Errorf("unexpected git log output: %q", rec)
		}

		authorDate, err := time.Parse(time.RFC3339, strings.TrimSpace(fields[2]))
		if err != nil {
			return nil, fmt.Errorf("parsing author date %q: %w", fields[2], err)
		}
		commitDate, err := time.Parse(time.RFC3339, strings.TrimSpace(fields[3]))
		if err != nil {
			return nil, fmt.Errorf("parsing commit date %q: %w", fields[3], err)
		}

		body := ""
		if len(fields) == 6 {
			body = strings.TrimSpace(fields[5])
		}

		commits = append(commits, CommitInfo{
			Hash:       strings.TrimSpace(fields[0]),
			ShortHash:  strings.TrimSpace(fields[1]),
			AuthorDate: authorDate,
			CommitDate: commitDate,
			Subject:    strings.TrimSpace(fields[4]),
			Body:       body,
		})
	}

	return commits, nil
}
