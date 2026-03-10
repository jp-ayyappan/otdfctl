package tui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/opentdf/otdfctl/pkg/handlers"
	"github.com/opentdf/otdfctl/tui/constants"
	"github.com/opentdf/platform/protocol/go/policy"
)

// KASKeyList lists the cryptographic keys for a single KAS server.
type KASKeyList struct {
	kas     *policy.KeyAccessServer
	list    list.Model
	h       TUIHandler
	spinner spinner.Model
	loading bool
}

type KASKeyItem struct {
	systemID  string
	kid       string
	algorithm string
	status    string
	mode      string
}

func (m KASKeyItem) FilterValue() string { return m.kid + " " + m.algorithm }
func (m KASKeyItem) Title() string {
	if m.kid != "" {
		return m.kid
	}
	return m.systemID
}
func (m KASKeyItem) Description() string {
	return fmt.Sprintf("%s  %s  %s", m.algorithm, m.status, m.mode)
}

type kasKeysLoadedMsg struct {
	items []list.Item
	err   error
}

func loadKASKeys(ctx context.Context, h TUIHandler, kasID string) tea.Cmd {
	return func() tea.Msg {
		res, err := h.ListKasKeys(ctx, 250, 0, policy.Algorithm_ALGORITHM_UNSPECIFIED,
			handlers.KasIdentifier{ID: kasID}, nil)
		if err != nil {
			return kasKeysLoadedMsg{err: err}
		}
		var items []list.Item
		for _, k := range res.GetKasKeys() {
			key := k.GetKey()
			items = append(items, KASKeyItem{
				systemID:  key.GetId(),
				kid:       key.GetKeyId(),
				algorithm: key.GetKeyAlgorithm().String(),
				status:    key.GetKeyStatus().String(),
				mode:      key.GetKeyMode().String(),
			})
		}
		return kasKeysLoadedMsg{items: items}
	}
}

func InitKASKeyList(ctx context.Context, kas *policy.KeyAccessServer, h TUIHandler) (tea.Model, tea.Cmd) {
	title := kas.GetName()
	if title == "" {
		title = kas.GetUri()
	}
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), constants.WindowSize.Width, constants.WindowSize.Height)
	l.Title = fmt.Sprintf("Keys: %s", title)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorPrimary)

	m := KASKeyList{kas: kas, h: h, list: l, spinner: s, loading: true}
	return m, tea.Batch(s.Tick, loadKASKeys(ctx, h, kas.GetId()))
}

func (m KASKeyList) Init() tea.Cmd { return nil }

func (m KASKeyList) KeyBindings() []KeyBinding {
	return []KeyBinding{
		{Key: "enter", Help: "view"},
		{Key: "/", Help: "filter"},
		{Key: "backspace", Help: "back"},
	}
}

func (m KASKeyList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	ctx := context.Background()

	switch msg := msg.(type) {
	case kasKeysLoadedMsg:
		m.loading = false
		if msg.err != nil {
			return m, func() tea.Msg {
				return StatusMsg{Text: "Error loading KAS keys: " + msg.err.Error(), IsError: true}
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
		case "backspace":
			return InitKASRegistryView(ctx, m.kas.GetId(), m.h)
		case "enter", "e":
			if m.loading || len(m.list.Items()) == 0 {
				return m, nil
			}
			selected := m.list.SelectedItem().(KASKeyItem)
			return InitKASKeyView(selected, m.kas, m.h)
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m KASKeyList) View() string {
	if m.loading {
		return lipgloss.NewStyle().Padding(1, 2).Render(m.spinner.View() + " Loading KAS keys…")
	}
	return ViewList(m.list)
}

// ---- KASKeyView ----

type KASKeyView struct {
	item KASKeyItem
	kas  *policy.KeyAccessServer
	read Read
	h    TUIHandler
}

func InitKASKeyView(item KASKeyItem, kas *policy.KeyAccessServer, h TUIHandler) (tea.Model, tea.Cmd) {
	items := []list.Item{
		AttributeSubItem{title: "System ID", description: item.systemID},
		AttributeSubItem{title: "KID", description: item.kid},
		AttributeSubItem{title: "KAS URI", description: kas.GetUri()},
		AttributeSubItem{title: "Algorithm", description: item.algorithm},
		AttributeSubItem{title: "Status", description: item.status},
		AttributeSubItem{title: "Mode", description: item.mode},
	}

	model, _ := InitRead(fmt.Sprintf("Key: %s", item.kid), items)
	mod := model.(Read)
	m := KASKeyView{item: item, kas: kas, h: h, read: mod}
	model, msg := m.Update(WindowMsg())
	return model, msg
}

func (m KASKeyView) Init() tea.Cmd { return nil }

func (m KASKeyView) KeyBindings() []KeyBinding {
	return []KeyBinding{
		{Key: "backspace", Help: "back to keys"},
	}
}

func (m KASKeyView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	ctx := context.Background()

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		constants.WindowSize = msg
		m.read.list.SetSize(msg.Width, msg.Height)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "backspace":
			return InitKASKeyList(ctx, m.kas, m.h)
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.read.list, cmd = m.read.list.Update(msg)
	return m, cmd
}

func (m KASKeyView) View() string {
	return m.read.View()
}
