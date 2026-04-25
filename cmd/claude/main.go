package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/atom-yt/claude-code-go/internal/session"
	"github.com/atom-yt/claude-code-go/internal/tui"
)

const version = "v0.1.0"

var (
	flagModel     string
	flagAPIKey    string
	flagProvider  string
	flagBaseURL   string
	flagVerbose   bool
	flagAltScreen  bool // 使用备用屏幕模式（默认false以兼容hermes-agent/claude-code）
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "claude [prompt]",
	Short: "Atom — AI-powered coding assistant",
	Long:  "Atom is an AI-powered CLI coding assistant.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var initialPrompt string
		if len(args) > 0 {
			initialPrompt = strings.TrimSpace(args[0])
		}

		cfg := tui.Config{
			Model:    flagModel,
			APIKey:   flagAPIKey,
			Provider: flagProvider,
			BaseURL:  flagBaseURL,
			Verbose:  flagVerbose,
		}

		m := tui.NewModel(cfg, initialPrompt)
		var opts []tea.ProgramOption
		if flagAltScreen {
			opts = append(opts, tea.WithAltScreen())
		}
		p := tea.NewProgram(m, opts...)
		_, err := p.Run()
		return err
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("claude-code-go %s\n", version)
	},
}

var resumeCmd = &cobra.Command{
	Use:   "resume [session-id]",
	Short: "Resume a previous session",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var rec session.Record
		var err error
		if len(args) == 1 {
			rec, err = session.Load(args[0])
		} else {
			rec, err = session.Latest()
		}
		if err != nil {
			return fmt.Errorf("no session to resume: %w", err)
		}

		cfg := tui.Config{
			Model:    firstNonEmpty(flagModel, rec.Model),
			APIKey:   flagAPIKey,
			Provider: flagProvider,
			BaseURL:  flagBaseURL,
			Verbose:  flagVerbose,
		}

		m := tui.NewModelWithHistory(cfg, rec)
		var opts []tea.ProgramOption
		if flagAltScreen {
			opts = append(opts, tea.WithAltScreen())
		}
		p := tea.NewProgram(m, opts...)
		_, err = p.Run()
		return err
	},
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagModel, "model", "claude-sonnet-4-6", "Model to use")
	rootCmd.PersistentFlags().StringVar(&flagAPIKey, "api-key", "", "API key (overrides ANTHROPIC_API_KEY / OPENAI_API_KEY)")
	rootCmd.PersistentFlags().StringVar(&flagProvider, "provider", "", "Provider: anthropic (default), openai, kimi, deepseek, codex, or custom")
	rootCmd.PersistentFlags().StringVar(&flagBaseURL, "base-url", "", "Custom API base URL (e.g. https://api.moonshot.cn)")
	rootCmd.PersistentFlags().BoolVar(&flagVerbose, "verbose", false, "Enable verbose/debug output")
	rootCmd.PersistentFlags().BoolVar(&flagAltScreen, "alt-screen", false, "Enable alternate screen mode (disables native copy/paste)")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(resumeCmd)
}
