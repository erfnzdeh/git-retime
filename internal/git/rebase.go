package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ExecuteRebase runs a headless git rebase -i, injecting the compiled todo
// via GIT_SEQUENCE_EDITOR. The todoPath is the path to the compiled rebase
// todo file that will replace the one git generates.
func ExecuteRebase(todoPath, base string, needsRoot bool) error {
	args := []string{"rebase", "-i", "--rebase-merges"}
	if needsRoot {
		args = append(args, "--root")
	} else {
		args = append(args, base)
	}

	seqEditor := fmt.Sprintf("cp %q", todoPath)

	cmd := exec.Command("git", args...)
	cmd.Env = append(os.Environ(), "GIT_SEQUENCE_EDITOR="+seqEditor)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		// Attempt auto-abort.
		abortErr := abortRebase()
		if abortErr != nil {
			return fmt.Errorf("rebase failed: %w\nadditionally, rebase --abort failed: %s", err, abortErr)
		}
		return fmt.Errorf("rebase failed (auto-aborted): %w\nhint: your repository has been restored to its original state", err)
	}
	return nil
}

func abortRebase() error {
	out, err := exec.Command("git", "rebase", "--abort").CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// HasMergeCommits checks whether the given range contains merge commits.
func HasMergeCommits(revision string) (bool, error) {
	out, err := exec.Command("git", "log", "--merges", "--oneline", revision+"..HEAD").CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("checking for merges: %s", err)
	}
	return strings.TrimSpace(string(out)) != "", nil
}
