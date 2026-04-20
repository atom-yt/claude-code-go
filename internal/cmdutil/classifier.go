// Package cmdutil provides command classification and security utilities.
package cmdutil

import (
	"fmt"
	"regexp"
	"strings"
)

// CommandCategory describes the type of command.
type CommandCategory int

const (
	CategorySafe CommandCategory = iota
	CategoryDestructive
	CategoryFileSystem
	CategoryNetwork
	CategoryPackage
)

func (c CommandCategory) String() string {
	switch c {
	case CategoryDestructive:
		return "destructive"
	case CategoryFileSystem:
		return "filesystem"
	case CategoryNetwork:
		return "network"
	case CategoryPackage:
		return "package"
	default:
		return "safe"
	}
}

// DestructiveCommand is a dangerous command pattern.
type DestructiveCommand struct {
	Pattern    string
	Reason     string
	Category   CommandCategory
	AutoReject bool
}

// Destructive patterns to detect.
// These are commands that can cause data loss or system damage.
var destructivePatterns = []DestructiveCommand{
	// File deletion
	{Pattern: `^rm\s+-rf\s+/`, Reason: "Recursive deletion from root", Category: CategoryDestructive, AutoReject: true},
	{Pattern: `^rm\s+/\*`, Reason: "Deleting all files in root", Category: CategoryDestructive, AutoReject: true},
	{Pattern: `^rm\s+-rf\s+\.\.?\.`, Reason: "Recursive deletion to parent", Category: CategoryDestructive, AutoReject: true},
	{Pattern: `^rm\s+-f\s+/\*`, Reason: "Force delete all in root", Category: CategoryDestructive, AutoReject: true},

	// Disk operations
	{Pattern: `^mkfs\.\w+\s+`, Reason: "Formatting filesystem", Category: CategoryDestructive, AutoReject: true},
	{Pattern: `^dd\s+`, Reason: "Low-level disk write", Category: CategoryDestructive, AutoReject: true},
	{Pattern: `^wipefs\s+`, Reason: "Wiping filesystem signatures", Category: CategoryDestructive, AutoReject: true},

	// System operations
	{Pattern: `^shutdown\b`, Reason: "System shutdown", Category: CategoryDestructive, AutoReject: true},
	{Pattern: `^reboot\b`, Reason: "System reboot", Category: CategoryDestructive, AutoReject: true},
	{Pattern: `^halt\b`, Reason: "System halt", Category: CategoryDestructive, AutoReject: true},
	{Pattern: `^poweroff\b`, Reason: "System poweroff", Category: CategoryDestructive, AutoReject: true},

	// Package removal
	{Pattern: `^apt-get\s+(autoremove|purge)`, Reason: "Removing packages", Category: CategoryPackage, AutoReject: false},
	{Pattern: `^apt(-get)?\s+remove\b`, Reason: "Removing packages", Category: CategoryPackage, AutoReject: false},
	{Pattern: `^(yum|dnf)\s+remove\b`, Reason: "Removing packages", Category: CategoryPackage, AutoReject: false},
	{Pattern: `^pip\s+uninstall\b`, Reason: "Removing Python packages", Category: CategoryPackage, AutoReject: false},
	{Pattern: `^npm\s+uninstall\s+-g\b`, Reason: "Removing global npm packages", Category: CategoryPackage, AutoReject: false},

	// File system
	{Pattern: `^chmod\s+-R\s+[0-9]+\s+/`, Reason: "Recursive permission change", Category: CategoryFileSystem, AutoReject: false},
	{Pattern: `^chown\s+-R\s+.*\s+/`, Reason: "Recursive ownership change", Category: CategoryFileSystem, AutoReject: false},

	// Network
	{Pattern: `^iptables\s+`, Reason: "Firewall modification", Category: CategoryNetwork, AutoReject: false},
	{Pattern: `^ufw\s+(disable|reset)`, Reason: "Firewall disable/reset", Category: CategoryNetwork, AutoReject: false},
	{Pattern: `^nft\s+`, Reason: "Firewall modification", Category: CategoryNetwork, AutoReject: false},

	// Git operations
	{Pattern: `^git\s+branch\s+-D\b`, Reason: "Force deleting git branch", Category: CategoryFileSystem, AutoReject: false},
	{Pattern: `^git\s+push\b.*--force\b`, Reason: "Force pushing to git", Category: CategoryFileSystem, AutoReject: false},
}

var compiledPatterns []struct {
	regex     *regexp.Regexp
	pattern   DestructiveCommand
}

func init() {
	compiledPatterns = make([]struct {
		regex     *regexp.Regexp
		pattern   DestructiveCommand
	}, len(destructivePatterns))
	for i, p := range destructivePatterns {
		compiledPatterns[i].regex = regexp.MustCompile(p.Pattern)
		compiledPatterns[i].pattern = p
	}
}

// Classify analyzes a shell command and returns its category.
// Returns the category, reason, and whether the command should be auto-rejected.
func Classify(command string) (CommandCategory, string, bool) {
	// Normalize command for matching
	normalized := strings.TrimSpace(command)
	if normalized == "" {
		return CategorySafe, "", false
	}

	// Check each destructive pattern
	for _, cp := range compiledPatterns {
		if cp.regex.MatchString(normalized) {
			return cp.pattern.Category, cp.pattern.Reason, cp.pattern.AutoReject
		}
	}

	return CategorySafe, "", false
}

// IsDestructive returns true if the command is classified as destructive.
func IsDestructive(command string) bool {
	cat, _, _ := Classify(command)
	return cat == CategoryDestructive
}

// GetWarning returns a warning message for a potentially dangerous command.
func GetWarning(command string) string {
	cat, reason, _ := Classify(command)
	if cat == CategorySafe {
		return ""
	}

	return fmt.Sprintf("⚠️  Warning: This command is classified as %s: %s", cat, reason)
}