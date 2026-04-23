# Harness Memory

Agent experience storage for learning and adaptation.

## Structure

- `episodic/` - Event-specific learnings from individual sessions
- `procedural/` - Standard workflows and procedures that have been validated

## Usage

Files in this directory are automatically read by agents to build context across sessions.

## Format

Use markdown format with clear headers:

```markdown
# Topic

## Context
Brief description of what was learned.

## Problem
What issue was being solved.

## Solution
The approach that worked.

## When to Use
Conditions where this pattern applies.
```