package tui

import (
	"context"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/opentdf/otdfctl/tui/constants"
	"github.com/opentdf/platform/protocol/go/policy"
)

type NamespaceView struct {
	ns   *policy.Namespace
	read Read
	h    TUIHandler
}

func InitNamespaceView(ctx context.Context, id string, h TUIHandler) (tea.Model, tea.Cmd) {
	ns, _ := h.GetNamespace(ctx, id)
	active := "false"
	if ns.GetActive().GetValue() {
		active = "true"
	}
	items := []list.Item{
		AttributeSubItem{title: "ID", description: ns.GetId()},
		AttributeSubItem{title: "Name", description: ns.GetName()},
		AttributeSubItem{title: "Active", description: active},
		AttributeSubItem{title: "FQN", description: ns.GetFqn()},
	}
	model, _ := InitRead("Namespace Detail", items)
	mod := model.(Read)
	m := NamespaceView{h: h, ns: ns, read: mod}
	model, msg := m.Update(WindowMsg())
	return model, msg
}

func (m NamespaceView) Init() tea.Cmd { return nil }

func (m NamespaceView) KeyBindings() []KeyBinding {
	return []KeyBinding{
		{Key: "backspace", Help: "back"},
	}
}

func (m NamespaceView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	ctx := context.Background()

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		constants.WindowSize = msg
		m.read.list.SetSize(msg.Width, msg.Height)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "backspace":
			return InitNamespaceList(ctx, m.h)
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.read.list, cmd = m.read.list.Update(msg)
	return m, cmd
}

func (m NamespaceView) View() string {
	return m.read.View()
}
