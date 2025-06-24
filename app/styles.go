package app

import "github.com/charmbracelet/lipgloss/v2"

var (
	currentStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffd700"))
	shouldBeStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#228b22"))
	styleFocused   = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff69b4")).Bold(true)
	styleUnfocused = lipgloss.NewStyle().Foreground(lipgloss.Color("#606060"))
	appStyle       = lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
	paneStyle      = lipgloss.NewStyle().Padding(2)
)
