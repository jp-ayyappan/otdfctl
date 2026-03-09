package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/opentdf/otdfctl/pkg/handlers"
	"github.com/opentdf/otdfctl/tui/constants"
)

type focusTarget int

const (
	focusSidebar focusTarget = iota
	focusContent
)

// Root is the top-level BubbleTea model. It owns the persistent layout:
// header, sidebar, content area, status bar, and keybar. Only the content
// area changes as the user navigates.
type Root struct {
	h         handlers.Handler
	header    Header
	sidebar   Sidebar
	content   tea.Model
	statusbar StatusBar
	keybar    Keybar
	focus     focusTarget
	width     int
	height    int
}

func NewRoot(h handlers.Handler, profile, endpoint string) (Root, tea.Cmd) {
	sidebar := NewSidebar()

	// Load default content (Attributes, index 1 in navItems)
	defaultItem := navItems[1]
	content, contentCmd := defaultItem.load(context.Background(), h)

	m := Root{
		h:         h,
		header:    NewHeader(profile, endpoint),
		sidebar:   sidebar,
		content:   content,
		statusbar: NewStatusBar(),
		keybar:    NewKeybar(globalBindings),
		focus:     focusContent,
	}
	return m, contentCmd
}

func (m Root) Init() tea.Cmd {
	return m.content.Init()
}

const statusBarHeight = 1

func (m Root) contentSize() (width, height int) {
	width = m.width - SidebarWidth - SidebarBorder
	height = m.height - HeaderHeight - statusBarHeight - KeybarHeight
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}
	return
}

func (m Root) contentSizeMsg() tea.WindowSizeMsg {
	w, h := m.contentSize()
	return tea.WindowSizeMsg{Width: w, Height: h}
}

func (m Root) activeBindings() []KeyBinding {
	var bindings []KeyBinding
	if m.focus == focusSidebar {
		bindings = append(bindings, sidebarBindings...)
	} else {
		if keyed, ok := m.content.(Keyed); ok {
			bindings = append(bindings, keyed.KeyBindings()...)
		}
	}
	bindings = append(bindings, globalBindings...)
	return bindings
}

func (m Root) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case StatusMsg:
		m.statusbar = m.statusbar.Set(msg)
		if msg.Text != "" {
			return m, clearStatusCmd(4 * 1000000000) // 4 seconds
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		constants.WindowSize = msg

		m.header = m.header.SetWidth(msg.Width)
		m.statusbar = m.statusbar.SetWidth(msg.Width)
		m.keybar = m.keybar.SetWidth(msg.Width)
		m.sidebar = m.sidebar.SetHeight(m.height - HeaderHeight - statusBarHeight - KeybarHeight)

		// Forward adjusted size to content
		newContent, cmd := m.content.Update(m.contentSizeMsg())
		m.content = newContent
		return m, cmd

	case NavSelectMsg:
		content, cmd := msg.item.load(context.Background(), m.h)
		m.content = content
		m.focus = focusContent
		m.keybar = m.keybar.SetBindings(m.activeBindings())
		// Send current content size to newly loaded view
		return m, tea.Batch(cmd, func() tea.Msg { return m.contentSizeMsg() })

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "tab":
			if m.focus == focusSidebar {
				m.focus = focusContent
				m.sidebar = m.sidebar.SetFocused(false)
			} else {
				m.focus = focusSidebar
				m.sidebar = m.sidebar.SetFocused(true)
			}
			m.keybar = m.keybar.SetBindings(m.activeBindings())
			return m, nil

		default:
			if m.focus == focusSidebar {
				var cmd tea.Cmd
				m.sidebar, cmd = m.sidebar.Update(msg)
				return m, cmd
			}
			// Route to content; content may swap itself out
			newContent, cmd := m.content.Update(msg)
			m.content = newContent
			m.keybar = m.keybar.SetBindings(m.activeBindings())
			return m, cmd
		}
	}

	// Propagate all other messages (spinner ticks, etc.) to content
	newContent, cmd := m.content.Update(msg)
	m.content = newContent
	return m, cmd
}

func (m Root) View() string {
	if m.width == 0 {
		return ""
	}

	cw, ch := m.contentSize()

	header := m.header.View()
	sidebar := m.sidebar.View()
	content := lipgloss.NewStyle().Width(cw).Height(ch).Render(m.content.View())
	body := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, content)
	status := m.statusbar.View()
	keybar := m.keybar.View()

	return lipgloss.JoinVertical(lipgloss.Left, header, body, status, keybar)
}
