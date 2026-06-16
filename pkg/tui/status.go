// Package tui provides terminal UI components for the Ubiquity CLI
// using charmbracelet/bubbletea and lipgloss for rich terminal output.
package tui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ubiquitycluster/ubiquity/pkg/provision"
)

type statusModel struct {
	state *provision.State
}

func (m statusModel) Init() tea.Cmd {
	return nil
}

func (m statusModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit
	}
	return m, nil
}

func (m statusModel) View() string {
	return RenderStatus(m.state)
}

var (
	// Colors
	colorGreen  = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	colorYellow = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	colorRed    = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	colorBlue   = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
	colorGray   = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	// Header style
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("7")).
			Background(lipgloss.Color("4")).
			Padding(0, 1).
			MarginBottom(1)

	// Phase styles
	phaseNameStyle = lipgloss.NewStyle().Width(15).Bold(true)
	statusStyle    = lipgloss.NewStyle().Width(12)
	durationStyle  = lipgloss.NewStyle().Width(10).Foreground(lipgloss.Color("8"))
	errorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Italic(true)
)

// StatusColor returns a styled status string with appropriate color.
func StatusColor(status provision.PhaseState) string {
	switch status {
	case provision.PhaseSuccess:
		return colorGreen.Render(string(status))
	case provision.PhaseRunning:
		return colorYellow.Render(string(status))
	case provision.PhaseFailed:
		return colorRed.Render(string(status))
	case provision.PhaseSkipped:
		return colorGray.Render(string(status))
	case provision.PhasePending:
		return colorGray.Render(string(status))
	default:
		return string(status)
	}
}

// RenderStatus generates a styled table of provisioning phases.
func RenderStatus(state *provision.State) string {
	if state == nil {
		return "No provisioning state found.\nRun 'ubiquity init' then 'ubiquity up' to start deployment."
	}

	var b strings.Builder

	b.WriteString(headerStyle.Render(fmt.Sprintf(" Ubiquity Cluster Status — %s ", state.Environment)))
	b.WriteString("\n\n")

	// Table header
	b.WriteString(phaseNameStyle.Render("Phase"))
	b.WriteString(" ")
	b.WriteString(statusStyle.Render("Status"))
	b.WriteString(" ")
	b.WriteString(durationStyle.Render("Duration"))
	b.WriteString("\n")

	// Separator
	sep := strings.Repeat("─", 40)
	b.WriteString(colorGray.Render(sep))
	b.WriteString("\n")

	// Phase rows
	for _, p := range state.Phases {
		duration := p.Duration
		if duration == "" && p.Status == provision.PhasePending {
			duration = "—"
		}
		if duration == "" && p.Status == provision.PhaseRunning {
			duration = "running…"
		}

		b.WriteString(colorBlue.Render("▎"))
		b.WriteString(" ")
		b.WriteString(phaseNameStyle.Render(p.Name))
		b.WriteString(statusStyle.Render(StatusColor(p.Status)))
		b.WriteString(durationStyle.Render(duration))
		b.WriteString("\n")

		if p.Error != "" {
			b.WriteString("  ")
			b.WriteString(errorStyle.Render(fmt.Sprintf("└─ Error: %s", p.Error)))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(colorGray.Render(fmt.Sprintf("Last updated: %s", state.UpdatedAt)))
	b.WriteString("\n")

	return b.String()
}

// PrintStatus renders the provisioning state to stdout.
// Falls back to plain text if the terminal doesn't support colors.
func PrintStatus(state *provision.State) {
	// Detect if we're in a TTY
	stat, _ := os.Stdout.Stat()
	isTerminal := (stat.Mode() & os.ModeCharDevice) != 0

	if !isTerminal {
		// Plain text fallback
		if state == nil {
			fmt.Println("No provisioning state found.")
			fmt.Println("Run 'ubiquity init' then 'ubiquity up' to start deployment.")
			return
		}
		fmt.Print(state.Summary())
		return
	}

	program := tea.NewProgram(statusModel{state: state}, tea.WithOutput(os.Stdout))
	if _, err := program.Run(); err != nil {
		fmt.Print(RenderStatus(state))
	}
}
