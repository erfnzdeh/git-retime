package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestIntegration_Shift builds git-retime, creates a temp repo with real
// commits, runs --shift, and verifies the timestamps changed.
func TestIntegration_Shift(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	binary := buildBinary(t)
	repoDir := createTempRepo(t, 5)

	origDates := getAuthorDates(t, repoDir)
	if len(origDates) != 5 {
		t.Fatalf("expected 5 commits, got %d", len(origDates))
	}

	// Retime the last 3 commits (HEAD~3 is the base, not included).
	runRetime(t, binary, repoDir, "HEAD~3", "--shift", "+2h")

	newDates := getAuthorDates(t, repoDir)
	if len(newDates) != 5 {
		t.Fatalf("expected 5 commits after retime, got %d", len(newDates))
	}

	// First 2 commits should be unchanged.
	for i := 0; i < 2; i++ {
		if origDates[i] != newDates[i] {
			t.Errorf("commit %d should be unchanged: orig=%s, new=%s", i, origDates[i], newDates[i])
		}
	}

	// Last 3 commits should be shifted by +2h.
	for i := 2; i < 5; i++ {
		origT, _ := time.Parse(time.RFC3339, origDates[i])
		newT, _ := time.Parse(time.RFC3339, newDates[i])
		diff := newT.Sub(origT)
		if diff != 2*time.Hour {
			t.Errorf("commit %d: expected +2h shift, got %v (orig=%s, new=%s)", i, diff, origDates[i], newDates[i])
		}
	}
}

// TestIntegration_Randomize tests the --randomize flag.
func TestIntegration_Randomize(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	binary := buildBinary(t)
	repoDir := createTempRepo(t, 4)

	runRetime(t, binary, repoDir, "HEAD~3", "--randomize", "09:00-17:00")

	newDates := getAuthorDates(t, repoDir)
	// Last 3 commits should have randomized hours.
	for i := 1; i < 4; i++ {
		ts, _ := time.Parse(time.RFC3339, newDates[i])
		h := ts.Hour()
		if h < 9 || h > 16 {
			t.Errorf("commit %d: hour %d not in [9,16] range (date=%s)", i, h, newDates[i])
		}
	}
}

// TestIntegration_RootCommit tests retiming that includes the root commit.
func TestIntegration_RootCommit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	binary := buildBinary(t)
	repoDir := createTempRepo(t, 3)

	origDates := getAuthorDates(t, repoDir)

	// Get the root commit hash.
	rootHash := strings.TrimSpace(runGit(t, repoDir, "rev-list", "--max-parents=0", "HEAD"))

	// Shift from root (root is included because it has no parent).
	runRetime(t, binary, repoDir, rootHash, "--shift", "+1h")

	newDates := getAuthorDates(t, repoDir)
	if len(newDates) != 3 {
		t.Fatalf("expected 3 commits after root retime, got %d", len(newDates))
	}

	// All commits should be shifted by 1 hour.
	for i := range origDates {
		origT, _ := time.Parse(time.RFC3339, origDates[i])
		newT, _ := time.Parse(time.RFC3339, newDates[i])
		diff := newT.Sub(origT)
		if diff != 1*time.Hour {
			t.Errorf("commit %d: expected +1h shift, got %v", i, diff)
		}
	}
}

// TestIntegration_MessagePreserved verifies commit messages survive retiming.
func TestIntegration_MessagePreserved(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	binary := buildBinary(t)
	repoDir := createTempRepo(t, 3)

	origMessages := getSubjects(t, repoDir)
	runRetime(t, binary, repoDir, "HEAD~2", "--shift", "+1h")
	newMessages := getSubjects(t, repoDir)

	for i := range origMessages {
		if origMessages[i] != newMessages[i] {
			t.Errorf("commit %d message changed: %q -> %q", i, origMessages[i], newMessages[i])
		}
	}
}

func buildBinary(t *testing.T) string {
	t.Helper()
	tmpBin := filepath.Join(t.TempDir(), "git-retime")
	cmd := exec.Command("go", "build", "-o", tmpBin, ".")
	cmd.Dir = getProjectRoot(t)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build failed: %v\n%s", err, string(out))
	}
	return tmpBin
}

func createTempRepo(t *testing.T, numCommits int) string {
	t.Helper()
	dir := t.TempDir()

	gitEnv := append(os.Environ(),
		"GIT_AUTHOR_NAME=Test",
		"GIT_AUTHOR_EMAIL=test@test.com",
		"GIT_COMMITTER_NAME=Test",
		"GIT_COMMITTER_EMAIL=test@test.com",
	)

	gitCmd := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = gitEnv
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, string(out))
		}
	}

	gitCmd("init")
	gitCmd("config", "user.email", "test@test.com")
	gitCmd("config", "user.name", "Test")

	baseTime := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	for i := 0; i < numCommits; i++ {
		commitTime := baseTime.Add(time.Duration(i) * time.Hour)
		dateStr := commitTime.Format(time.RFC3339)

		filename := filepath.Join(dir, strings.Repeat("f", i+1)+".txt")
		os.WriteFile(filename, []byte("content"), 0644)

		gitCmd("add", "-A")

		cmd := exec.Command("git", "commit", "-m", "Commit "+string(rune('A'+i)))
		cmd.Dir = dir
		cmd.Env = append(gitEnv,
			"GIT_AUTHOR_DATE="+dateStr,
			"GIT_COMMITTER_DATE="+dateStr,
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("commit %d failed: %v\n%s", i, err, string(out))
		}
	}

	return dir
}

func runRetime(t *testing.T, binary, repoDir string, args ...string) {
	t.Helper()
	cmd := exec.Command(binary, args...)
	cmd.Dir = repoDir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=Test",
		"GIT_AUTHOR_EMAIL=test@test.com",
		"GIT_COMMITTER_NAME=Test",
		"GIT_COMMITTER_EMAIL=test@test.com",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git-retime %v failed: %v\noutput: %s", args, err, string(out))
	}
}

func runGit(t *testing.T, repoDir string, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", repoDir}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(out))
	}
	return string(out)
}

func getAuthorDates(t *testing.T, repoDir string) []string {
	t.Helper()
	out := runGit(t, repoDir, "log", "--reverse", "--format=%aI")
	return nonEmpty(strings.Split(strings.TrimSpace(out), "\n"))
}

func getSubjects(t *testing.T, repoDir string) []string {
	t.Helper()
	out := runGit(t, repoDir, "log", "--reverse", "--format=%s")
	return nonEmpty(strings.Split(strings.TrimSpace(out), "\n"))
}

func nonEmpty(ss []string) []string {
	var out []string
	for _, s := range ss {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func getProjectRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find project root (go.mod)")
		}
		dir = parent
	}
}
