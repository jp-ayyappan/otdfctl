package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Keybar renders a bottom bar showing context-sensitive key bindings.
type Keybar struct {
	bindings []KeyBinding
	width    int
}

func NewKeybar(bindings []KeyBinding) Keybar {
	return Keybar{bindings: bindings}
}

func (k Keybar) SetWidth(w int) Keybar {
	k.width = w
	return k
}

func (k Keybar) SetBindings(bindings []KeyBinding) Keybar {
	k.bindings = bindings
	return k
}

func (k Keybar) View() string {
	var parts []string
	for _, b := range k.bindings {
		key := keyStyle.Render(b.Key)
		help := keyHelpStyle.Render(b.Help)
		parts = append(parts, key+help)
	}
	content := strings.Join(parts, " ")
	return lipgloss.NewStyle().
		Background(lipgloss.Color("#1a1a2e")).
		Width(k.width).
		Render(content)
}

// globalBindings are always shown regardless of context.
var globalBindings = []KeyBinding{
	{Key: "tab", Help: "switch panel"},
	{Key: "ctrl+c", Help: "quit"},
}

// sidebarBindings are shown when the sidebar has focus.
var sidebarBindings = []KeyBinding{
	{Key: "↑/↓", Help: "navigate"},
	{Key: "enter", Help: "open"},
}
