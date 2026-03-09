package tui

import (
	"context"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/opentdf/otdfctl/pkg/handlers"
	"github.com/opentdf/otdfctl/tui/constants"
)

type KASRegistryList struct {
	list list.Model
	h    handlers.Handler
}

type KASRegistryItem struct {
	id   string
	name string
	uri  string
}

func (m KASRegistryItem) FilterValue() string { return m.name + " " + m.uri }
func (m KASRegistryItem) Title() string {
	if m.name != "" {
		return m.name
	}
	return m.uri
}
func (m KASRegistryItem) Description() string { return m.uri }

func InitKASRegistryList(ctx context.Context, h handlers.Handler) (tea.Model, tea.Cmd) {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), constants.WindowSize.Width, constants.WindowSize.Height)
	l.Title = "KAS Registry"

	res, _ := h.ListKasRegistryEntries(ctx, 250, 0)
	var items []list.Item
	for _, kas := range res.GetKeyAccessServers() {
		items = append(items, KASRegistryItem{
			id:   kas.GetId(),
			name: kas.GetName(),
			uri:  kas.GetUri(),
		})
	}
	l.SetItems(items)

	m := KASRegistryList{h: h, list: l}
	return m.Update(WindowMsg())
}

func (m KASRegistryList) Init() tea.Cmd { return nil }

func (m KASRegistryList) KeyBindings() []KeyBinding {
	return []KeyBinding{
		{Key: "enter", Help: "view"},
		{Key: "/", Help: "filter"},
	}
}

func (m KASRegistryList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			id := m.list.SelectedItem().(KASRegistryItem).id
			return InitKASRegistryView(ctx, id, m.h)
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m KASRegistryList) View() string {
	return ViewList(m.list)
}
