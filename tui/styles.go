package tui

import "github.com/charmbracelet/lipgloss"

const (
	SidebarWidth  = 24
	HeaderHeight  = 1
	KeybarHeight  = 1
	SidebarBorder = 1 // right border
)

var (
	ColorPrimary = lipgloss.Color("#7D56F4")
	ColorAccent  = lipgloss.Color("#EE6FF8")
	ColorDim     = lipgloss.Color("#4a4a5a")
	ColorText    = lipgloss.Color("#FAFAFA")
	ColorMuted   = lipgloss.Color("#888888")
	ColorSuccess = lipgloss.Color("#73F59F")
	ColorError   = lipgloss.Color("#FF5555")

	sidebarGroupStyle = lipgloss.NewStyle().
				Padding(0, 1).
				Foreground(ColorMuted).
				Bold(true)

	sidebarItemStyle = lipgloss.NewStyle().
				Padding(0, 2)

	sidebarItemSelectedStyle = lipgloss.NewStyle().
					Padding(0, 1).
					Foreground(ColorPrimary).
					Bold(true)

	sidebarItemFocusedStyle = lipgloss.NewStyle().
				Padding(0, 1).
				Background(ColorPrimary).
				Foreground(ColorText).
				Bold(true)

	keyStyle = lipgloss.NewStyle().
			Background(ColorDim).
			Foreground(ColorText).
			Padding(0, 1)

	keyHelpStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			PaddingRight(2)

	placeholderStyle = lipgloss.NewStyle().
				Foreground(ColorMuted)
)
