package tui

import (
	"context"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/opentdf/otdfctl/pkg/handlers"
	"github.com/opentdf/otdfctl/tui/constants"
	"github.com/opentdf/platform/protocol/go/common"
)

type NamespaceList struct {
	list    list.Model
	h       handlers.Handler
	spinner spinner.Model
	loading bool
}

type NamespaceItem struct {
	id   string
	name string
}

func (m NamespaceItem) FilterValue() string { return m.name }
func (m NamespaceItem) Title() string       { return m.name }
func (m NamespaceItem) Description() string { return m.id }

type namespacesLoadedMsg struct {
	items []list.Item
	err   error
}

func loadNamespaces(ctx context.Context, h handlers.Handler) tea.Cmd {
	return func() tea.Msg {
		res, err := h.ListNamespaces(ctx, common.ActiveStateEnum_ACTIVE_STATE_ENUM_ANY, 250, 0)
		if err != nil {
			return namespacesLoadedMsg{err: err}
		}
		var items []list.Item
		for _, ns := range res.GetNamespaces() {
			items = append(items, NamespaceItem{id: ns.GetId(), name: ns.GetName()})
		}
		return namespacesLoadedMsg{items: items}
	}
}

func InitNamespaceList(ctx context.Context, h handlers.Handler) (tea.Model, tea.Cmd) {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), constants.WindowSize.Width, constants.WindowSize.Height)
	l.Title = "Namespaces"

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorPrimary)

	m := NamespaceList{h: h, list: l, spinner: s, loading: true}
	return m, tea.Batch(s.Tick, loadNamespaces(ctx, h))
}

func (m NamespaceList) Init() tea.Cmd { return nil }

func (m NamespaceList) KeyBindings() []KeyBinding {
	return []KeyBinding{
		{Key: "enter", Help: "view"},
		{Key: "/", Help: "filter"},
	}
}

func (m NamespaceList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	ctx := context.Background()

	switch msg := msg.(type) {
	case namespacesLoadedMsg:
		m.loading = false
		if msg.err != nil {
			return m, func() tea.Msg {
				return StatusMsg{Text: "Error loading namespaces: " + msg.err.Error(), IsError: true}
			}
		}
		m.list.SetItems(msg.items)
		return m, nil

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case tea.WindowSizeMsg:
		constants.WindowSize = msg
		m.list.SetSize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter", "e":
			if m.loading || len(m.list.Items()) == 0 {
				return m, nil
			}
			return InitNamespaceView(ctx, m.list.SelectedItem().(NamespaceItem).id, m.h)
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m NamespaceList) View() string {
	if m.loading {
		return lipgloss.NewStyle().Padding(1, 2).Render(m.spinner.View() + " Loading namespaces…")
	}
	return ViewList(m.list)
}
