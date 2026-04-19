package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// commitCmd implements the /commit command.
type commitCmd struct{}

func (c *commitCmd) Name() string        { return "commit" }
func (c *commitCmd) Aliases() []string   { return []string{"c"} }
func (c *commitCmd) Description() string { return "Create a git commit with AI-generated message" }

func (c *commitCmd) Execute(ctx context.Context, args []string, cmdCtx *Context) (string, error) {
	// Check if we're in a git repository
	if !isGitRepo() {
		return "Error: Not in a git repository", nil
	}

	// Get git status
	status, err := runGit("status", "--short")
	if err != nil {
		return fmt.Sprintf("Error getting git status: %v", err), nil
	}

	if strings.TrimSpace(status) == "" {
		return "No changes to commit", nil
	}

	// Get diff to understand changes
	diff, err := runGit("diff", "--staged")
	if err == nil && strings.TrimSpace(diff) != "" {
		// Staged changes exist
		return c.createCommitFromStaged(ctx, status, diff)
	}

	// No staged changes, show unstaged
	diff, err = runGit("diff")
	if err != nil {
		return fmt.Sprintf("Error getting diff: %v", err), nil
	}

	// Show status and ask for confirmation (simplified for now)
	return fmt.Sprintf("Changes detected:\n%s\n\nStaging changes and creating commit...", status), c.stageAndCommit(ctx)
}

// createCommitFromStaged creates a commit from already-staged changes
func (c *commitCmd) createCommitFromStaged(ctx context.Context, status, diff string) (string, error) {
	// Parse changes and draft commit message
	commitMsg, err := c.generateCommitMessage(ctx, status, diff)
	if err != nil {
		return fmt.Sprintf("Error generating commit message: %v", err), nil
	}

	// Create commit with the message
	return c.doCommit(commitMsg)
}

// stageAndCommit stages all changes and creates a commit
func (c *commitCmd) stageAndCommit(ctx context.Context) error {
	// Stage all changes
	if _, err := runGit("add", "-A"); err != nil {
		return fmt.Errorf("failed to stage changes: %w", err)
	}

	// Get status and diff
	status, _ := runGit("status", "--short")
	diff, _ := runGit("diff", "--cached")

	// Generate commit message
	commitMsg, err := c.generateCommitMessage(ctx, status, diff)
	if err != nil {
		return err
	}

	// Create commit
	_, err = c.doCommit(commitMsg)
	return err
}

// generateCommitMessage generates a commit message based on changes
func (c *commitCmd) generateCommitMessage(ctx context.Context, status, diff string) (string, error) {
	// Analyze the changes to determine commit type
	commitType := "chore"
	var changes []string

	lines := strings.Split(status, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		// Parse git status output: M file, A file, D file, etc.
		if len(line) < 3 {
			continue
		}

		statusCode := line[:2]
		filePath := strings.TrimSpace(line[2:])

		// Determine commit type based on changes
		switch {
		case strings.Contains(filePath, "test"):
			commitType = "test"
		case strings.Contains(filePath, "doc"):
			commitType = "docs"
		case strings.Contains(filePath, "cmd/") || strings.Contains(filePath, "main.go"):
			commitType = "build"
		case statusCode[0] == 'M' || statusCode[1] == 'M':
			if commitType == "chore" {
				commitType = "fix"
			}
		case statusCode[0] == 'A' || statusCode[1] == 'A':
			if commitType == "chore" {
				commitType = "feat"
			}
		}

		changes = append(changes, filePath)
	}

	// Build commit message
	var msgBuilder strings.Builder
	msgBuilder.WriteString(fmt.Sprintf("%s: ", commitType))

	// Generate a brief description
	if len(changes) == 1 {
		msgBuilder.WriteString(fmt.Sprintf("Update %s", changes[0]))
	} else if len(changes) <= 3 {
		msgBuilder.WriteString(fmt.Sprintf("Update %d files", len(changes)))
	} else {
		msgBuilder.WriteString(fmt.Sprintf("Multiple updates (%d files)", len(changes)))
	}

	// Add co-author footer
	msgBuilder.WriteString("\n\n")
	msgBuilder.WriteString("Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>")

	return msgBuilder.String(), nil
}

// doCommit creates the git commit
func (c *commitCmd) doCommit(message string) (string, error) {
	// Use git commit with the message
	output, err := runGit("commit", "-m", message)
	if err != nil {
		// Check if it was a pre-commit hook failure
		if strings.Contains(output, "pre-commit hook") {
			return fmt.Sprintf("Commit failed due to pre-commit hook:\n%s\n\nPlease fix the issues and try again.", output), nil
		}
		return fmt.Sprintf("Commit failed: %v\n%s", err, output), nil
	}

	// Show the commit
	commitOutput, _ := runGit("log", "-1", "--pretty=format:%h - %s")
	return fmt.Sprintf("Commit created:\n%s", commitOutput), nil
}

// isGitRepo checks if the current directory is a git repository
func isGitRepo() bool {
	_, err := runGit("rev-parse", "--is-inside-work-tree")
	return err == nil
}

// runGit executes a git command and returns the output
func runGit(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = getGitDir()
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// getGitDir returns the directory to run git commands in
// Returns the current working directory or .claude/worktrees/ if inside one
func getGitDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}

	// Check if we're in a worktree
	if strings.Contains(wd, "/.claude/worktrees/") {
		return wd
	}

	return wd
}
