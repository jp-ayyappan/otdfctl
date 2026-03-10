package tui

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/opentdf/otdfctl/tui/constants"
	"github.com/opentdf/platform/protocol/go/policy"
)

type ResourceMappingView struct {
	rm   *policy.ResourceMapping
	read Read
	h    TUIHandler
}

func InitResourceMappingView(ctx context.Context, id string, h TUIHandler) (tea.Model, tea.Cmd) {
	rm, _ := h.GetResourceMapping(id)

	items := []list.Item{
		AttributeSubItem{title: "ID", description: rm.GetId()},
		AttributeSubItem{title: "Attribute Value", description: rm.GetAttributeValue().GetFqn()},
		AttributeSubItem{title: "Terms", description: strings.Join(rm.GetTerms(), ", ")},
		AttributeSubItem{title: "Group ID", description: rm.GetGroup().GetId()},
	}

	model, _ := InitRead("Resource Mapping Detail", items)
	mod := model.(Read)
	m := ResourceMappingView{h: h, rm: rm, read: mod}
	model, msg := m.Update(WindowMsg())
	return model, msg
}

func (m ResourceMappingView) Init() tea.Cmd { return nil }

func (m ResourceMappingView) KeyBindings() []KeyBinding {
	return []KeyBinding{
		{Key: "backspace", Help: "back"},
	}
}

func (m ResourceMappingView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	ctx := context.Background()

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		constants.WindowSize = msg
		m.read.list.SetSize(msg.Width, msg.Height)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "backspace":
			return InitResourceMappingList(ctx, m.h)
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.read.list, cmd = m.read.list.Update(msg)
	return m, cmd
}

func (m ResourceMappingView) View() string {
	return m.read.View()
}
