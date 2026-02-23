package compile

import (
	"fmt"
	"strings"

	ts "github.com/erfnzdeh/git-retime/internal/timestamp"
)

// Compile translates resolved commits into a git rebase-todo file.
// Each commit becomes a "pick" line followed by an "exec" line that
// amends the commit's timestamps (and optionally the message).
func Compile(commits []ts.Commit) string {
	var b strings.Builder

	for _, c := range commits {
		shortHash := c.Hash
		if len(shortHash) > 7 {
			shortHash = shortHash[:7]
		}

		subject := c.NewSubject
		if subject == "" {
			subject = c.Subject
		}

		fmt.Fprintf(&b, "pick %s %s\n", shortHash, subject)
		b.WriteString(buildExec(c))
		b.WriteByte('\n')
	}

	return b.String()
}

func buildExec(c ts.Commit) string {
	authorDate := ts.FormatGit(c.ResolvedAuthorDate)
	commitDate := ts.FormatGit(c.ResolvedCommitDate)

	subjectChanged := c.NewSubject != "" && c.NewSubject != c.Subject

	// GIT_COMMITTER_DATE env var sets the committer date.
	// --date flag sets the author date (GIT_AUTHOR_DATE env var doesn't
	// override during amend).
	var parts []string
	parts = append(parts, "exec")
	parts = append(parts, fmt.Sprintf("GIT_COMMITTER_DATE=%q", commitDate))

	if subjectChanged {
		parts = append(parts, "git commit --amend --allow-empty")
		parts = append(parts, fmt.Sprintf("--date=%q", authorDate))
		parts = append(parts, buildMessageArgs(c)...)
	} else {
		parts = append(parts, "git commit --amend --no-edit --allow-empty")
		parts = append(parts, fmt.Sprintf("--date=%q", authorDate))
	}

	return strings.Join(parts, " ")
}

func buildMessageArgs(c ts.Commit) []string {
	// Reconstruct the full message: new subject + original body.
	fullMsg := c.NewSubject
	if c.Body != "" {
		fullMsg += "\n\n" + c.Body
	}
	return []string{"-m", shellQuote(fullMsg)}
}

func shellQuote(s string) string {
	// Use $'...' quoting for strings with newlines, single quotes, etc.
	replacer := strings.NewReplacer(
		`\`, `\\`,
		`'`, `\'`,
		"\n", `\n`,
		"\t", `\t`,
	)
	return "$'" + replacer.Replace(s) + "'"
}
