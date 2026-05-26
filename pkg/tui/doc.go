// Package tui provides terminal UI components using charmbracelet/bubbletea
// and lipgloss for styled output. Currently provides a colored status table
// that replaces the plain-text summary from provisioning state.
//
// Usage:
//   tui.PrintStatus(state)      // auto-detects TTY, falls back to plain text
//   fmt.Print(tui.RenderStatus(state)) // always renders styled output
package tui