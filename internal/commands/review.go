package commands

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// reviewCmd implements the /review command.
type reviewCmd struct{}

func (c *reviewCmd) Name() string      { return "review" }
func (c *reviewCmd) Aliases() []string { return nil }
func (c *reviewCmd) Description() string {
	return "Review code changes for quality, bugs, and best practices"
}

func (c *reviewCmd) Execute(ctx context.Context, args []string, cmdCtx *Context) (string, error) {
	// Check if we're in a git repository
	if !isGitRepo() {
		return "Error: Not in a git repository", nil
	}

	// Determine what to review
	target := "staged" // default to staged changes
	if len(args) > 0 {
		target = args[0]
	}

	switch target {
	case "staged", "":
		return c.reviewStaged(ctx)
	case "unstaged":
		return c.reviewUnstaged(ctx)
	case "last":
		return c.reviewLastCommit(ctx)
	case "all":
		return c.reviewAllChanges(ctx)
	default:
		// Treat as file/path
		return c.reviewPath(ctx, target)
	}
}

// reviewStaged reviews staged changes
func (c *reviewCmd) reviewStaged(ctx context.Context) (string, error) {
	// Get staged diff
	diff, err := runGit("diff", "--staged")
	if err != nil {
		return fmt.Sprintf("Error getting staged diff: %v", err), nil
	}

	if strings.TrimSpace(diff) == "" {
		return "No staged changes to review", nil
	}

	// Get list of changed files
	files, err := runGit("diff", "--staged", "--name-only")
	if err != nil {
		return fmt.Sprintf("Error getting file list: %v", err), nil
	}

	return c.formatReviewResult("Staged Changes", files, diff), nil
}

// reviewUnstaged reviews unstaged changes
func (c *reviewCmd) reviewUnstaged(ctx context.Context) (string, error) {
	diff, err := runGit("diff")
	if err != nil {
		return fmt.Sprintf("Error getting unstaged diff: %v", err), nil
	}

	if strings.TrimSpace(diff) == "" {
		return "No unstaged changes to review", nil
	}

	files, err := runGit("diff", "--name-only")
	if err != nil {
		return fmt.Sprintf("Error getting file list: %v", err), nil
	}

	return c.formatReviewResult("Unstaged Changes", files, diff), nil
}

// reviewLastCommit reviews the last commit
func (c *reviewCmd) reviewLastCommit(ctx context.Context) (string, error) {
	diff, err := runGit("show", "--stat", "HEAD")
	if err != nil {
		return fmt.Sprintf("Error getting last commit: %v", err), nil
	}

	// Get the actual diff
	fullDiff, err := runGit("show", "HEAD")
	if err != nil {
		return fmt.Sprintf("Error getting commit diff: %v", err), nil
	}

	return c.formatReviewResult("Last Commit", diff, fullDiff), nil
}

// reviewAllChanges reviews all uncommitted changes (staged + unstaged)
func (c *reviewCmd) reviewAllChanges(ctx context.Context) (string, error) {
	status, err := runGit("status", "--short")
	if err != nil {
		return fmt.Sprintf("Error getting status: %v", err), nil
	}

	if strings.TrimSpace(status) == "" {
		return "No changes to review", nil
	}

	// Get combined diff
	stagedDiff, _ := runGit("diff", "--staged")
	unstagedDiff, _ := runGit("diff")

	combinedDiff := stagedDiff + unstagedDiff

	return c.formatReviewResult("All Changes", status, combinedDiff), nil
}

// reviewPath reviews a specific file or directory
func (c *reviewCmd) reviewPath(ctx context.Context, path string) (string, error) {
	// Check if path exists
	if _, err := os.Stat(path); err != nil {
		return fmt.Sprintf("Path not found: %s", path), nil
	}

	// Try to get git diff for the path
	diff, err := runGit("diff", "HEAD", "--", path)
	if err != nil {
		// If not tracked, show the file content for review
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Sprintf("Error reading file: %v", err), nil
		}
		return fmt.Sprintf("Review for %s (untracked):\n\n%s", path, string(content)), nil
	}

	if strings.TrimSpace(diff) == "" {
		// No changes, show current content
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Sprintf("Error reading file: %v", err), nil
		}
		return fmt.Sprintf("No changes in %s. Current content:\n\n%s", path, truncateContent(string(content))), nil
	}

	return c.formatReviewResult(fmt.Sprintf("Changes in %s", path), path, diff), nil
}

// formatReviewResult formats the review output
func (c *reviewCmd) formatReviewResult(title, files, diff string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("## Code Review: %s\n\n", title))

	// Files section
	sb.WriteString("### Files Changed\n")
	sb.WriteString("```\n")
	sb.WriteString(files)
	sb.WriteString("\n```\n\n")

	// Diff section (truncated if too large)
	sb.WriteString("### Changes\n")
	sb.WriteString("```diff\n")
	sb.WriteString(truncateDiff(diff))
	sb.WriteString("\n```\n\n")

	// Review checklist
	sb.WriteString("### Review Checklist\n")
	sb.WriteString("- [ ] Code compiles without errors\n")
	sb.WriteString("- [ ] No obvious bugs or logic errors\n")
	sb.WriteString("- [ ] Code follows project conventions\n")
	sb.WriteString("- [ ] Error handling is appropriate\n")
	sb.WriteString("- [ ] No security vulnerabilities\n")
	sb.WriteString("- [ ] Tests are added/updated if needed\n")
	sb.WriteString("- [ ] Documentation is updated if needed\n")

	return sb.String()
}

// truncateDiff truncates diff if too large
func truncateDiff(diff string) string {
	const maxLen = 10000
	if len(diff) > maxLen {
		return diff[:maxLen] + "\n\n... (diff truncated, too large)"
	}
	return diff
}

// truncateContent truncates content if too large
func truncateContent(content string) string {
	const maxLen = 5000
	if len(content) > maxLen {
		return content[:maxLen] + "\n\n... (content truncated)"
	}
	return content
}

// Security review command
type securityReviewCmd struct{}

func (c *securityReviewCmd) Name() string        { return "security-review" }
func (c *securityReviewCmd) Aliases() []string   { return []string{"sec-review"} }
func (c *securityReviewCmd) Description() string { return "Review code for security vulnerabilities" }

func (c *securityReviewCmd) Execute(ctx context.Context, args []string, cmdCtx *Context) (string, error) {
	// Check if we're in a git repository
	if !isGitRepo() {
		return "Error: Not in a git repository", nil
	}

	// Get all changes
	status, err := runGit("status", "--short")
	if err != nil {
		return fmt.Sprintf("Error getting status: %v", err), nil
	}

	if strings.TrimSpace(status) == "" {
		return "No changes to review", nil
	}

	// Get diff
	stagedDiff, _ := runGit("diff", "--staged")
	unstagedDiff, _ := runGit("diff")
	combinedDiff := stagedDiff + unstagedDiff

	return c.formatSecurityReview(status, combinedDiff), nil
}

func (c *securityReviewCmd) formatSecurityReview(files, diff string) string {
	var sb strings.Builder

	sb.WriteString("## Security Review\n\n")

	// Files section
	sb.WriteString("### Files to Review\n")
	sb.WriteString("```\n")
	sb.WriteString(files)
	sb.WriteString("\n```\n\n")

	// Security checklist
	sb.WriteString("### Security Checklist\n\n")

	checks := []struct {
		category string
		items    []string
	}{
		{
			category: "Input Validation",
			items: []string{
				"- [ ] All user inputs are validated and sanitized",
				"- [ ] No SQL injection vulnerabilities",
				"- [ ] No command injection vulnerabilities",
				"- [ ] No XSS vulnerabilities",
			},
		},
		{
			category: "Authentication & Authorization",
			items: []string{
				"- [ ] Proper authentication checks",
				"- [ ] Authorization is enforced",
				"- [ ] No hardcoded credentials",
			},
		},
		{
			category: "Data Protection",
			items: []string{
				"- [ ] Sensitive data is encrypted",
				"- [ ] No sensitive data in logs",
				"- [ ] Proper key management",
			},
		},
		{
			category: "Error Handling",
			items: []string{
				"- [ ] Errors don't leak sensitive info",
				"- [ ] Proper error handling throughout",
			},
		},
		{
			category: "Dependencies",
			items: []string{
				"- [ ] No known vulnerable dependencies",
				"- [ ] Dependencies are up to date",
			},
		},
	}

	for _, check := range checks {
		sb.WriteString(fmt.Sprintf("#### %s\n", check.category))
		for _, item := range check.items {
			sb.WriteString(item + "\n")
		}
		sb.WriteString("\n")
	}

	// Common vulnerability patterns to check
	sb.WriteString("### Patterns to Check\n")
	sb.WriteString("```\n")
	sb.WriteString("Common patterns that may indicate vulnerabilities:\n")
	sb.WriteString("- eval(), exec(), system() calls with user input\n")
	sb.WriteString("- String concatenation in SQL queries\n")
	sb.WriteString("- Unvalidated file paths\n")
	sb.WriteString("- Hardcoded secrets/API keys\n")
	sb.WriteString("- Unencrypted sensitive data\n")
	sb.WriteString("- Missing authentication checks\n")
	sb.WriteString("```\n")

	return sb.String()
}
