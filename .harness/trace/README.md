# Harness Trace

Execution traces and failure records for quality improvement.

## Structure

- `failures/` - Records of failed operations and how they were resolved

## Purpose

Track recurring issues to identify patterns and improve processes.

## Format

```markdown
# Failure: {description}

## Date
{ISO date}

## Type
- layer_violation
- quality_rule
- test_failure
- validation_error

## Severity
- critical
- warning
- info

## Context
- Task: what was being done
- File: file that triggered failure
- Rule: which rule violated

## Error Output
```
error message
```

## Resolution
How it was fixed, or "unresolved" if pending.
```