package tui

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/opentdf/otdfctl/pkg/handlers"
	"github.com/opentdf/otdfctl/tui/constants"
)

type ResourceMappingList struct {
	list list.Model
	h    handlers.Handler
}

type ResourceMappingItem struct {
	id    string
	terms string
}

func (m ResourceMappingItem) FilterValue() string { return m.terms }
func (m ResourceMappingItem) Title() string       { return m.terms }
func (m ResourceMappingItem) Description() string { return m.id }

func InitResourceMappingList(ctx context.Context, h handlers.Handler) (tea.Model, tea.Cmd) {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), constants.WindowSize.Width, constants.WindowSize.Height)
	l.Title = "Resource Mappings"

	res, _ := h.ListResourceMappings(ctx, 250, 0)
	var items []list.Item
	for _, rm := range res.GetResourceMappings() {
		items = append(items, ResourceMappingItem{
			id:    rm.GetId(),
			terms: strings.Join(rm.GetTerms(), ", "),
		})
	}
	l.SetItems(items)

	m := ResourceMappingList{h: h, list: l}
	return m.Update(WindowMsg())
}

func (m ResourceMappingList) Init() tea.Cmd { return nil }

func (m ResourceMappingList) KeyBindings() []KeyBinding {
	return []KeyBinding{
		{Key: "enter", Help: "view"},
		{Key: "/", Help: "filter"},
	}
}

func (m ResourceMappingList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			id := m.list.SelectedItem().(ResourceMappingItem).id
			return InitResourceMappingView(ctx, id, m.h)
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m ResourceMappingList) View() string {
	return ViewList(m.list)
}
