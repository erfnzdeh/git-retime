package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// GetEditor resolves the user's preferred editor using git's own resolution
// logic: GIT_SEQUENCE_EDITOR -> GIT_EDITOR -> core.editor -> VISUAL -> EDITOR -> vi
func GetEditor() (string, error) {
	out, err := exec.Command("git", "var", "GIT_EDITOR").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("cannot determine editor: %s\n%s", err, string(out))
	}
	editor := strings.TrimSpace(string(out))
	if editor == "" {
		return "vi", nil
	}
	return editor, nil
}

// OpenEditor launches the editor with the given file path, connecting it
// to the user's terminal (stdin/stdout/stderr).
func OpenEditor(editor, filePath string) error {
	parts := strings.Fields(editor)
	if len(parts) == 0 {
		return fmt.Errorf("empty editor command")
	}

	args := append(parts[1:], filePath)
	cmd := exec.Command(parts[0], args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
