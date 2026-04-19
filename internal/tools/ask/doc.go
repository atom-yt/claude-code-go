// Package ask provides the AskUserQuestion tool for interactive user decision support.
//
// The AskUserQuestion tool allows agents to:
//
// 1. Ask questions during execution to gather user input
// 2. Present multiple choice options with descriptions
// 3. Support preview content for visual comparison
// 4. Allow multi-select questions
// 5. Accept custom "Other" text input
//
// Question structure:
//
// - question: The question text (required)
// - header: Short label displayed as a chip/tag (optional, max 12 chars)
// - options: List of 2-4 options with label, description, and optional preview
// - multiSelect: Whether multiple options can be selected (optional, defaults to false)
//
// The tool is useful for:
//
// - Gathering user preferences and requirements
// - Clarifying ambiguous instructions
// - Getting decisions on implementation choices
// - Offering choices between different approaches
//
// Note: This tool requires interactive user input and should be used strategically
// when user decision is truly needed for the task.
package ask
