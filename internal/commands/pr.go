package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// prCmd implements the /pr command.
type prCmd struct{}

func (c *prCmd) Name() string        { return "pr" }
func (c *prCmd) Aliases() []string   { return nil }
func (c *prCmd) Description() string { return "Create a pull request with AI-generated description" }

func (c *prCmd) Execute(ctx context.Context, args []string, cmdCtx *Context) (string, error) {
	// Check if we're in a git repository
	if !isGitRepo() {
		return "Error: Not in a git repository", nil
	}

	// Check if gh CLI is installed
	if !hasGhCLI() {
		return "Error: GitHub CLI (gh) is not installed. Install it from https://cli.github.com/", nil
	}

	// Get current branch
	branch, err := getCurrentBranch()
	if err != nil {
		return fmt.Sprintf("Error getting current branch: %v", err), nil
	}

	// Check if branch has changes compared to base
	baseBranch := "main"
	// Check if origin/main exists
	if _, err := runGit("rev-parse", "--verify", "origin/main"); err == nil {
		baseBranch = "origin/main"
	} else if _, err := runGit("rev-parse", "--verify", "origin/master"); err == nil {
		baseBranch = "origin/master"
	}

	// Generate PR content
	title, body, err := c.generatePRContent(ctx, branch, baseBranch)
	if err != nil {
		return fmt.Sprintf("Error generating PR content: %v", err), nil
	}

	// Create PR
	return c.createPR(title, body)
}

// generatePRContent generates the PR title and body
func (c *prCmd) generatePRContent(ctx context.Context, branch, baseBranch string) (string, string, error) {
	// Get commits between base and current branch
	commits, err := runGit("log", fmt.Sprintf("%s..HEAD", baseBranch), "--pretty=format:%s")
	if err != nil {
		return "", "", err
	}

	// Get changed files
	files, err := runGit("diff", "--name-only", fmt.Sprintf("%s...HEAD", baseBranch))
	if err != nil {
		return "", "", err
	}

	// Get full diff (limited)
	diff, err := runGit("diff", fmt.Sprintf("%s...HEAD", baseBranch))
	if err != nil {
		return "", "", err
	}

	// Parse commits to determine PR type
	commitLines := strings.Split(commits, "\n")
	var features, fixes, docs, tests, others []string

	for _, commit := range commitLines {
		commit = strings.TrimSpace(commit)
		if commit == "" {
			continue
		}

		switch {
		case strings.HasPrefix(commit, "feat:"):
			features = append(features, commit[6:])
		case strings.HasPrefix(commit, "fix:"):
			fixes = append(fixes, commit[5:])
		case strings.HasPrefix(commit, "docs:"):
			docs = append(docs, commit[6:])
		case strings.HasPrefix(commit, "test:"):
			tests = append(tests, commit[6:])
		default:
			others = append(others, commit)
		}
	}

	// Generate title
	title := c.generateTitle(branch, features, fixes, docs, tests, others)

	// Generate body
	body := c.generateBody(branch, features, fixes, docs, tests, others, files, diff)

	return title, body, nil
}

// generateTitle generates a PR title
func (c *prCmd) generateTitle(branch string, features, fixes, docs, tests, others []string) string {
	switch {
	case len(features) > 0:
		if len(features) == 1 {
			return fmt.Sprintf("feat: %s", features[0])
		}
		return fmt.Sprintf("feat: %d new features", len(features))
	case len(fixes) > 0:
		if len(fixes) == 1 {
			return fmt.Sprintf("fix: %s", fixes[0])
		}
		return fmt.Sprintf("fix: %d bug fixes", len(fixes))
	case len(docs) > 0:
		return fmt.Sprintf("docs: %s", branch)
	case len(tests) > 0:
		return fmt.Sprintf("test: %s", branch)
	default:
		return fmt.Sprintf("chore: %s", branch)
	}
}

// generateBody generates a PR body
func (c *prCmd) generateBody(branch string, features, fixes, docs, tests, others []string, files, diff string) string {
	var body strings.Builder

	body.WriteString("## Summary\n\n")

	// Build summary bullets
	var bullets []string
	if len(features) > 0 {
		bullets = append(bullets, fmt.Sprintf("- %d new feature(s)", len(features)))
		for _, f := range features {
			bullets = append(bullets, fmt.Sprintf("  - %s", f))
		}
	}
	if len(fixes) > 0 {
		bullets = append(bullets, fmt.Sprintf("- %d bug fix(es)", len(fixes)))
		for _, f := range fixes {
			bullets = append(bullets, fmt.Sprintf("  - %s", f))
		}
	}
	if len(docs) > 0 {
		bullets = append(bullets, fmt.Sprintf("- %d documentation update(s)", len(docs)))
	}
	if len(tests) > 0 {
		bullets = append(bullets, fmt.Sprintf("- %d test addition(s)/update(s)", len(tests)))
	}
	if len(others) > 0 {
		bullets = append(bullets, fmt.Sprintf("- %d other change(s)", len(others)))
	}

	if len(bullets) > 0 {
		for _, b := range bullets {
			body.WriteString(b + "\n")
		}
	} else {
		body.WriteString("- Various improvements and bug fixes\n")
	}

	// Test plan section
	body.WriteString("\n## Test plan\n\n")
	body.WriteString("- [ ] Code compiles without errors\n")
	body.WriteString("- [ ] Unit tests pass: `make test`\n")
	body.WriteString("- [ ] Manual testing completed\n")

	// Changed files section
	if strings.TrimSpace(files) != "" {
		fileList := strings.Split(files, "\n")
		if len(fileList) > 0 && len(fileList) < 20 {
			body.WriteString("\n## Changed files\n\n")
			for _, f := range fileList {
				if strings.TrimSpace(f) != "" {
					body.WriteString("- " + f + "\n")
				}
			}
		}
	}

	// Footer
	body.WriteString("\n---\n")
	body.WriteString("🤖 Generated with [Claude Code](https://claude.com/claude-code)")

	return body.String()
}

// createPR creates the pull request using gh CLI
func (c *prCmd) createPR(title, body string) (string, error) {
	// Create a temp file for the body
	tmpFile, err := os.CreateTemp("", "pr-body-*.txt")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(body); err != nil {
		tmpFile.Close()
		return "", fmt.Errorf("failed to write body: %w", err)
	}
	tmpFile.Close()

	// Run gh pr create
	cmd := exec.Command("gh", "pr", "create", "--title", title, "--body-file", tmpFile.Name())
	cmd.Dir = getGitDir()
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Sprintf("Failed to create PR: %v\n%s", err, string(output)), nil
	}

	return fmt.Sprintf("Pull request created:\n%s", string(output)), nil
}

// getCurrentBranch returns the current git branch
func getCurrentBranch() (string, error) {
	return runGit("rev-parse", "--abbrev-ref", "HEAD")
}

// getRemote returns the default remote name
func getRemote() (string, error) {
	output, err := runGit("remote", "show")
	if err != nil {
		return "", err
	}
	remotes := strings.Split(output, "\n")
	if len(remotes) > 0 && strings.TrimSpace(remotes[0]) != "" {
		return strings.TrimSpace(remotes[0]), nil
	}
	return "", fmt.Errorf("no remote found")
}

// hasGhCLI checks if GitHub CLI is installed
func hasGhCLI() bool {
	_, err := exec.LookPath("gh")
	return err == nil
}
