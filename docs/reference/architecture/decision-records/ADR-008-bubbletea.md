# ADR-008: Use Bubbletea for TUI instead of plain text

**Status:** Accepted

**Date:** 2025-06-01

## Context

Need colored, interactive terminal output for status monitoring.

## Decision

Use charmbracelet/bubbletea + lipgloss. Rich terminal UX, cross-platform, composable components. Fallback to plain text when not in a TTY.

## Consequences

- Positive: Rich, interactive terminal experience.
- Positive: Cross-platform composable components.
- Negative: Adds 4 dependencies to go.mod (~500KB).
- Neutral: --plain flag available for CI/pipe usage.
