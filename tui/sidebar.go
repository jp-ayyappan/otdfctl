package tui

import (
	"context"
	"strings"

	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/opentdf/otdfctl/pkg/handlers"
)

// NavSelectMsg is sent when a sidebar item is selected.
type NavSelectMsg struct {
	item navItem
}

type navItem struct {
	title string
	group string
	load  func(ctx context.Context, h handlers.Handler) (tea.Model, tea.Cmd)
}

// navItems defines the full navigation tree. Items without a load fn get a placeholder.
var navItems = []navItem{
	{
		group: "Policy",
		title: "Namespaces",
		load: func(ctx context.Context, h handlers.Handler) (tea.Model, tea.Cmd) {
			return NewPlaceholder("Namespaces"), nil
		},
	},
	{
		group: "Policy",
		title: "Attributes",
		load: func(ctx context.Context, h handlers.Handler) (tea.Model, tea.Cmd) {
			return InitAttributeList(ctx, "", h)
		},
	},
	{
		group: "Policy",
		title: "Attr Values",
		load: func(ctx context.Context, h handlers.Handler) (tea.Model, tea.Cmd) {
			return NewPlaceholder("Attribute Values"), nil
		},
	},
	{
		group: "Policy",
		title: "Subject Mappings",
		load: func(ctx context.Context, h handlers.Handler) (tea.Model, tea.Cmd) {
			return NewPlaceholder("Subject Mappings"), nil
		},
	},
	{
		group: "Policy",
		title: "KAS Registry",
		load: func(ctx context.Context, h handlers.Handler) (tea.Model, tea.Cmd) {
			return NewPlaceholder("KAS Registry"), nil
		},
	},
	{
		group: "Policy",
		title: "KAS Grants",
		load: func(ctx context.Context, h handlers.Handler) (tea.Model, tea.Cmd) {
			return NewPlaceholder("KAS Grants"), nil
		},
	},
	{
		group: "Policy",
		title: "Resource Mappings",
		load: func(ctx context.Context, h handlers.Handler) (tea.Model, tea.Cmd) {
			return NewPlaceholder("Resource Mappings"), nil
		},
	},
}

// Sidebar holds the left-hand navigation panel.
type Sidebar struct {
	items    []navItem
	selected int
	focused  bool
	height   int
}

func NewSidebar() Sidebar {
	// default selection: Attributes (index 1)
	return Sidebar{items: navItems, selected: 1}
}

func (s Sidebar) SetHeight(h int) Sidebar {
	s.height = h
	return s
}

func (s Sidebar) SetFocused(f bool) Sidebar {
	s.focused = f
	return s
}

func (s Sidebar) KeyBindings() []KeyBinding {
	return sidebarBindings
}

func (s Sidebar) Update(msg tea.Msg) (Sidebar, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "up", "k":
			if s.selected > 0 {
				s.selected--
			}
		case "down", "j":
			if s.selected < len(s.items)-1 {
				s.selected++
			}
		case "enter":
			item := s.items[s.selected]
			return s, func() tea.Msg { return NavSelectMsg{item: item} }
		}
	}
	return s, nil
}

func (s Sidebar) View() string {
	var lines []string
	var lastGroup string

	for i, item := range s.items {
		if item.group != lastGroup {
			lastGroup = item.group
			lines = append(lines, sidebarGroupStyle.Render(item.group))
		}
		var style lipgloss.Style
		switch {
		case i == s.selected && s.focused:
			style = sidebarItemFocusedStyle.Width(SidebarWidth - 2)
		case i == s.selected:
			style = sidebarItemSelectedStyle.Width(SidebarWidth - 2)
		default:
			style = sidebarItemStyle.Width(SidebarWidth - 2)
		}
		prefix := "  "
		if i == s.selected {
			prefix = "▶ "
		}
		lines = append(lines, style.Render(prefix+item.title))
	}

	content := strings.Join(lines, "\n")

	borderColor := ColorDim
	if s.focused {
		borderColor = ColorPrimary
	}

	return lipgloss.NewStyle().
		Width(SidebarWidth).
		Height(s.height).
		BorderRight(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(borderColor).
		Render(content)
}
