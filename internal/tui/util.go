package tui

import (
	"os"

	"github.com/charmbracelet/glamour"
)

// getCWD returns the current working directory, or empty string on error.
func getCWD() (string, error) {
	return os.Getwd()
}

// glamourRenderer holds a cached glamour renderer for a given width.
type glamourRenderer struct {
	width    int
	renderer *glamour.TermRenderer
}

// get returns the cached renderer if the width matches, otherwise builds a new one.
func (g *glamourRenderer) get(width int) *glamour.TermRenderer {
	if g.renderer != nil && g.width == width {
		return g.renderer
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return nil
	}
	g.width = width
	g.renderer = r
	return r
}
