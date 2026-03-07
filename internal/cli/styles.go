package cli

import "github.com/charmbracelet/lipgloss"

func newStyles() *Styles {
	// move this to internal/tui/styles.go later and import from there
	return &Styles{
		Title:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")),
		Header:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("141")),
		Success: lipgloss.NewStyle().Foreground(lipgloss.Color("76")),
		Error:   lipgloss.NewStyle().Foreground(lipgloss.Color("203")),
		Path:    lipgloss.NewStyle().Foreground(lipgloss.Color("45")),
		Dim:     lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
	}
}

// Styles struct (can be shared)
type Styles struct {
	Title   lipgloss.Style
	Header  lipgloss.Style
	Success lipgloss.Style
	Error   lipgloss.Style
	Path    lipgloss.Style
	Dim     lipgloss.Style
}
