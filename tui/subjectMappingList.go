package tui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/opentdf/otdfctl/pkg/handlers"
	"github.com/opentdf/otdfctl/tui/constants"
)

type SubjectMappingList struct {
	list list.Model
	h    handlers.Handler
}

type SubjectMappingItem struct {
	id       string
	attrVal  string
	actionCt int
}

func (m SubjectMappingItem) FilterValue() string { return m.attrVal }
func (m SubjectMappingItem) Title() string       { return m.attrVal }
func (m SubjectMappingItem) Description() string {
	return fmt.Sprintf("%s  (%d actions)", m.id, m.actionCt)
}

func InitSubjectMappingList(ctx context.Context, h handlers.Handler) (tea.Model, tea.Cmd) {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), constants.WindowSize.Width, constants.WindowSize.Height)
	l.Title = "Subject Mappings"

	res, _ := h.ListSubjectMappings(ctx, 250, 0)
	var items []list.Item
	for _, sm := range res.GetSubjectMappings() {
		items = append(items, SubjectMappingItem{
			id:       sm.GetId(),
			attrVal:  sm.GetAttributeValue().GetFqn(),
			actionCt: len(sm.GetActions()),
		})
	}
	l.SetItems(items)

	m := SubjectMappingList{h: h, list: l}
	return m.Update(WindowMsg())
}

func (m SubjectMappingList) Init() tea.Cmd { return nil }

func (m SubjectMappingList) KeyBindings() []KeyBinding {
	return []KeyBinding{
		{Key: "enter", Help: "view"},
		{Key: "/", Help: "filter"},
	}
}

func (m SubjectMappingList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			id := m.list.SelectedItem().(SubjectMappingItem).id
			return InitSubjectMappingView(ctx, id, m.h)
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m SubjectMappingList) View() string {
	return ViewList(m.list)
}
