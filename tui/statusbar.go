package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// StatusMsg sets a message in the root status bar.
type StatusMsg struct {
	Text    string
	IsError bool
}

// clearStatusCmd clears the status bar after a delay.
func clearStatusCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(_ time.Time) tea.Msg {
		return StatusMsg{}
	})
}

// StatusBar renders a single line below the content area for feedback messages.
type StatusBar struct {
	text    string
	isError bool
	width   int
}

func NewStatusBar() StatusBar {
	return StatusBar{}
}

func (s StatusBar) SetWidth(w int) StatusBar {
	s.width = w
	return s
}

func (s StatusBar) Set(msg StatusMsg) StatusBar {
	s.text = msg.Text
	s.isError = msg.IsError
	return s
}

func (s StatusBar) View() string {
	if s.text == "" {
		return lipgloss.NewStyle().
			Width(s.width).
			Background(lipgloss.Color("#111122")).
			Render("")
	}

	color := ColorSuccess
	if s.isError {
		color = ColorError
	}

	return lipgloss.NewStyle().
		Width(s.width).
		Background(lipgloss.Color("#111122")).
		Foreground(color).
		Padding(0, 1).
		Render(s.text)
}
