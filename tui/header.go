package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Header renders a single-line top bar with the app name on the left
// and the active profile + endpoint on the right.
type Header struct {
	profile  string
	endpoint string
	width    int
}

func NewHeader(profile, endpoint string) Header {
	return Header{profile: profile, endpoint: endpoint}
}

func (h Header) SetWidth(w int) Header {
	h.width = w
	return h
}

func (h Header) View() string {
	appName := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorText).
		Render(" OpenTDF")

	right := ""
	if h.profile != "" || h.endpoint != "" {
		meta := h.profile
		if h.endpoint != "" {
			if meta != "" {
				meta += " @ "
			}
			meta += h.endpoint
		}
		right = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Render(meta + " ")
	}

	leftWidth := h.width - lipgloss.Width(right)
	if leftWidth < 0 {
		leftWidth = 0
	}

	left := lipgloss.NewStyle().
		Background(ColorPrimary).
		Width(leftWidth).
		Render(appName)

	rightStyled := lipgloss.NewStyle().
		Background(ColorPrimary).
		Render(right)

	return lipgloss.JoinHorizontal(lipgloss.Top, left, rightStyled)
}
