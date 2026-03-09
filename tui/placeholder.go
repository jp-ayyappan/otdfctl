package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Placeholder is shown for resource views not yet implemented.
type Placeholder struct {
	title  string
	width  int
	height int
}

func NewPlaceholder(title string) Placeholder {
	return Placeholder{title: title}
}

func (p Placeholder) Init() tea.Cmd { return nil }

func (p Placeholder) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		p.width = msg.Width
		p.height = msg.Height
	}
	return p, nil
}

func (p Placeholder) View() string {
	return lipgloss.NewStyle().
		Width(p.width).
		Height(p.height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(
			placeholderStyle.Render("[ "+p.title+" ]") + "\n\n" +
				lipgloss.NewStyle().Foreground(ColorDim).Render("coming soon"),
		)
}

func (p Placeholder) KeyBindings() []KeyBinding {
	return nil
}
