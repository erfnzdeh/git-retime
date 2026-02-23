# git-retime

The missing interactive time editor for Git. Edit commit timestamps using a clean, data-centric interface.

`git-retime` opens your editor with a structured todo file where you can shift, randomize, or rewrite commit timestamps using a concise syntax, then applies the changes via a headless `git rebase -i`.

## Install

**macOS / Linux:**

```bash
curl -fsSL https://raw.githubusercontent.com/erfnzdeh/git-retime/main/install/install.sh | bash
```

**Windows (PowerShell):**

```powershell
iwr -useb https://raw.githubusercontent.com/erfnzdeh/git-retime/main/install/install.ps1 | iex
```

**From source:**

```bash
go install github.com/erfnzdeh/git-retime@latest
```

## Usage

```bash
git retime HEAD~5           # Open editor for the last 5 commits
git retime abc1234          # Retime from abc1234 to HEAD
git retime HEAD~3 --shift +2h          # Shift last 3 commits by 2 hours
git retime HEAD~5 --randomize 09:00-17:00  # Randomize times within working hours
```

Running `git retime HEAD~5` opens your editor with a file like this:

```
# Retime 5 commits onto abc1234
#
a1b2c3d  2026-02-23 10:00:00  Fix navbar
f012345  2026-02-23 10:30:00  Create user models
c5d6e7f  2026-02-23 11:00:00  Add API endpoints
8901abc  2026-02-23 11:45:00  Write tests
e5f6a7b  2026-02-23 12:15:00  Update README
```

Edit the middle column, save, and the timestamps are rewritten.

## Syntax Reference

| Syntax | Example | Meaning |
|--------|---------|---------|
| Absolute | `2026-02-23 14:00:00` | Set an exact time |
| Shift | `2026-02-23 10:00:00 +2h` | Shift from the written time |
| Bare shift | `+2h` or `-30m` | Shift from the original time |
| Compound | `+1d2h30m` | 1 day, 2 hours, 30 minutes |
| `PREV` | `PREV` | Same as previous commit's original time |
| `PREV` + offset | `PREV +45m` | Previous original time + 45 minutes |
| `NOW` | `NOW` | Current wall-clock time |
| `RR` | `RR:RR:00` | Random hour (0-23), random minute (0-59) |
| `RR(min,max)` | `RR(09,17):RR:00` | Random hour in 9-17, random minute |

**Units:** `w` (weeks), `d` (days), `h` (hours), `m` (minutes), `s` (seconds)

**RR** only works on time fields (HH:MM:SS), not date fields.

## Editing Commit Messages

The last column is the commit message subject. Editing it will rewrite the commit message (the body is preserved).

## Time Paradox Detection

If an edited timestamp creates a child commit older than its parent, `git-retime` warns you and asks whether to proceed:

```
warning: time paradox detected
  f012345 (2026-02-23 09:00:00) is older than a1b2c3d (2026-02-23 10:00:00)
Proceed anyway? [y/N]
```

## Flags

| Flag | Description |
|------|-------------|
| `--shift +2h` | Non-interactive: shift all commits by an offset |
| `--randomize 09:00-17:00` | Non-interactive: randomize time-of-day within a range |
| `--split-dates` | Edit author and committer dates independently (two timestamp columns) |
| `-i` | Accepted for compatibility (interactive is the default) |

## Aborting

To abort a retime session, either:
- Delete all lines in the editor
- Write `ABORT` on the first line

## How It Works

1. Fetches commits via `git log`
2. Generates a `.git-retime-todo` file with timestamps in your local timezone
3. Opens your `$GIT_EDITOR`
4. Parses edits, computes deltas, and applies them to the original timestamps (preserving the original timezone offset)
5. Compiles a `git rebase -i` todo with `pick` + `exec` lines that amend each commit's dates
6. Executes a headless rebase via `GIT_SEQUENCE_EDITOR`

If the rebase fails, `git-retime` automatically runs `git rebase --abort` to restore your repository.

## Constraints

- You cannot delete or reorder lines (only timestamps and messages are editable)
- `RR` is only supported in time fields, not date fields
- `PREV` refers to the **original** timestamp of the previous commit, not its edited value
- Merge topology is preserved via `--rebase-merges`
- Root commits are supported (uses `--root` internally)

## License

MIT
