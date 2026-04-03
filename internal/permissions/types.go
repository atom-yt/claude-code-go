// Package permissions implements the permission checker and rule matching.
package permissions

// Mode controls the default permission behavior.
type Mode string

const (
	ModeDefault  Mode = "default"   // ask on first use per tool
	ModeTrustAll Mode = "trust-all" // auto-approve everything
	ModeManual   Mode = "manual"    // always ask
)

// Rule describes a match condition for a tool call.
// Fields are ANDed together; omitted fields match anything.
type Rule struct {
	// Tool name, e.g. "Bash", "Write". Empty matches all tools.
	Tool string `json:"tool"`
	// Path glob applied to the file_path or path input argument.
	Path string `json:"path"`
	// Command prefix applied to the command input argument (Bash tool).
	Command string `json:"command"`
}

// Decision is the result of a permission check.
type Decision struct {
	Allowed bool
	Reason  string // populated when Allowed is false
}

// AskRequest carries context for an interactive permission prompt.
type AskRequest struct {
	ToolName  string
	Input     map[string]any
	RuleMatch string // which rule triggered the ask
}
