package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCWD(t *testing.T) {
	cwd, err := getCWD()
	assert.NoError(t, err, "getCWD should not return an error")
	assert.NotEmpty(t, cwd, "getCWD should return a non-empty string")
}

func TestGlamourRenderer_New(t *testing.T) {
	g := &glamourRenderer{}
	assert.Nil(t, g.renderer, "New renderer should have nil renderer")
	assert.Equal(t, 0, g.width, "New renderer should have width 0")
}

func TestGlamourRenderer_Get_CachesRenderer(t *testing.T) {
	g := &glamourRenderer{}
	r1 := g.get(80)
	r2 := g.get(80)

	assert.NotNil(t, r1, "get should return a renderer")
	assert.Same(t, r1, r2, "Should return the same renderer for same width")
}

func TestGlamourRenderer_Get_RebuildsOnWidthChange(t *testing.T) {
	g := &glamourRenderer{}
	r1 := g.get(80)
	r2 := g.get(100)

	assert.NotNil(t, r1, "get should return a renderer")
	assert.NotNil(t, r2, "get should return a renderer")
	assert.NotSame(t, r1, r2, "Should return different renderers for different widths")
}

func TestGlamourRenderer_Get_AfterFirstCall(t *testing.T) {
	g := &glamourRenderer{}
	r1 := g.get(80)

	assert.NotNil(t, r1, "First call should build and return a renderer")
	assert.NotNil(t, g.renderer, "Renderer should be cached after first call")
	assert.Equal(t, 80, g.width, "Width should be cached")
}