package tui

import (
	"context"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/opentdf/otdfctl/pkg/handlers"
	"github.com/opentdf/otdfctl/tui/constants"
	"github.com/opentdf/platform/protocol/go/common"
)

type NamespaceList struct {
	list list.Model
	h    handlers.Handler
}

type NamespaceItem struct {
	id   string
	name string
}

func (m NamespaceItem) FilterValue() string { return m.name }
func (m NamespaceItem) Title() string       { return m.name }
func (m NamespaceItem) Description() string { return m.id }

func InitNamespaceList(ctx context.Context, h handlers.Handler) (tea.Model, tea.Cmd) {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), constants.WindowSize.Width, constants.WindowSize.Height)
	l.Title = "Namespaces"

	res, _ := h.ListNamespaces(ctx, common.ActiveStateEnum_ACTIVE_STATE_ENUM_ANY, 250, 0)
	var items []list.Item
	for _, ns := range res.GetNamespaces() {
		items = append(items, NamespaceItem{id: ns.GetId(), name: ns.GetName()})
	}
	l.SetItems(items)

	m := NamespaceList{h: h, list: l}
	return m.Update(WindowMsg())
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
	case tea.WindowSizeMsg:
		constants.WindowSize = msg
		m.list.SetSize(msg.Width, msg.Height)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter", "e":
			if len(m.list.Items()) == 0 {
				return m, nil
			}
			id := m.list.SelectedItem().(NamespaceItem).id
			return InitNamespaceView(ctx, id, m.h)
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m NamespaceList) View() string {
	return ViewList(m.list)
}
