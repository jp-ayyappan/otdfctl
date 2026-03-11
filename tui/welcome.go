package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Welcome is the initial content pane shown before the user selects a resource.
type Welcome struct {
	width  int
	height int
}

func NewWelcome() Welcome { return Welcome{} }

func (w Welcome) Init() tea.Cmd { return nil }

func (w Welcome) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		w.width = msg.Width
		w.height = msg.Height
	}
	return w, nil
}

func (w Welcome) View() string {
	body := lipgloss.NewStyle().
		Foreground(ColorAccent).
		Bold(true).
		Render("OpenTDF Control Plane") +
		"\n\n" +
		lipgloss.NewStyle().
			Foreground(ColorMuted).
			Render("Use ↑/↓ or j/k to navigate the sidebar, then press enter to load a resource.") +
		"\n\n" +
		lipgloss.NewStyle().
			Foreground(ColorDim).
			Render("tab → switch between sidebar and content\nctrl+c → quit")

	return lipgloss.NewStyle().
		Width(w.width).
		Height(w.height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(body)
}
