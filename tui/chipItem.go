package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ChipItem is a list.Item that optionally carries a navigation action.
// Navigable chips render with a "→" indicator and respond to enter.
type ChipItem struct {
	label    string
	detail   string
	navigate func(ctx context.Context, h TUIHandler) (tea.Model, tea.Cmd)
}

func NewChip(label, detail string, navigate func(ctx context.Context, h TUIHandler) (tea.Model, tea.Cmd)) ChipItem {
	return ChipItem{label: label, detail: detail, navigate: navigate}
}

func NewInfoChip(label, detail string) ChipItem {
	return ChipItem{label: label, detail: detail}
}

func (c ChipItem) FilterValue() string { return c.label + " " + c.detail }

func (c ChipItem) Title() string {
	if c.navigate != nil {
		return lipgloss.NewStyle().Foreground(ColorAccent).Render("→") + " " + c.label
	}
	return "  " + c.label
}

func (c ChipItem) Description() string { return c.detail }
func (c ChipItem) Navigable() bool     { return c.navigate != nil }
