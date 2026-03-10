package tui

import (
	"context"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/opentdf/otdfctl/pkg/handlers"
	"github.com/opentdf/otdfctl/tui/constants"
	"github.com/opentdf/platform/protocol/go/policy"
)

type KASRegistryView struct {
	kas  *policy.KeyAccessServer
	read Read
	h    TUIHandler
}

func InitKASRegistryView(ctx context.Context, id string, h TUIHandler) (tea.Model, tea.Cmd) {
	kas, _ := h.GetKasRegistryEntry(ctx, handlers.KasIdentifier{ID: id})

	pubKey := ""
	if kas.GetPublicKey() != nil {
		if kas.GetPublicKey().GetCached() != nil {
			pubKey = "(cached)"
		} else if kas.GetPublicKey().GetRemote() != "" {
			pubKey = kas.GetPublicKey().GetRemote()
		}
	}

	items := []list.Item{
		AttributeSubItem{title: "ID", description: kas.GetId()},
		AttributeSubItem{title: "Name", description: kas.GetName()},
		AttributeSubItem{title: "URI", description: kas.GetUri()},
		AttributeSubItem{title: "Public Key", description: pubKey},
		AttributeSubItem{title: "Keys", description: "→ view cryptographic keys"},
	}

	model, _ := InitRead("KAS Registry Detail", items)
	mod := model.(Read)
	m := KASRegistryView{h: h, kas: kas, read: mod}
	model, msg := m.Update(WindowMsg())
	return model, msg
}

func (m KASRegistryView) Init() tea.Cmd { return nil }

func (m KASRegistryView) KeyBindings() []KeyBinding {
	return []KeyBinding{
		{Key: "enter", Help: "open Keys"},
		{Key: "backspace", Help: "back"},
	}
}

func (m KASRegistryView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	ctx := context.Background()

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		constants.WindowSize = msg
		m.read.list.SetSize(msg.Width, msg.Height)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "backspace":
			return InitKASRegistryList(ctx, m.h)
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			if m.read.list.SelectedItem().(AttributeSubItem).title == "Keys" {
				return InitKASKeyList(ctx, m.kas, m.h)
			}
		}
	}

	var cmd tea.Cmd
	m.read.list, cmd = m.read.list.Update(msg)
	return m, cmd
}

func (m KASRegistryView) View() string {
	return m.read.View()
}
