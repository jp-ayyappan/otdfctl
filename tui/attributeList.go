package tui

import (
	"context"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/opentdf/otdfctl/tui/constants"
	"github.com/opentdf/platform/protocol/go/common"
)

type AttributeList struct {
	list    list.Model
	h       TUIHandler
	spinner spinner.Model
	loading bool
}

type AttributeItem struct {
	id   string
	name string
}

func (m AttributeItem) FilterValue() string { return m.name }
func (m AttributeItem) Title() string       { return m.name }
func (m AttributeItem) Description() string { return m.id }

type attributesLoadedMsg struct {
	items []list.Item
	err   error
}

func loadAttributes(ctx context.Context, h TUIHandler, selectID string) tea.Cmd {
	return func() tea.Msg {
		res, err := h.ListAttributes(ctx, common.ActiveStateEnum_ACTIVE_STATE_ENUM_ANY, 250, 0)
		if err != nil {
			return attributesLoadedMsg{err: err}
		}
		var items []list.Item
		for _, attr := range res.GetAttributes() {
			items = append(items, AttributeItem{id: attr.GetId(), name: attr.GetName()})
		}
		return attributesLoadedMsg{items: items}
	}
}

func InitAttributeList(ctx context.Context, id string, h TUIHandler) (tea.Model, tea.Cmd) {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), constants.WindowSize.Width, constants.WindowSize.Height)
	l.Title = "Attributes"

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorPrimary)

	m := AttributeList{h: h, list: l, spinner: s, loading: true}
	return m, tea.Batch(s.Tick, loadAttributes(ctx, h, id))
}

func (m AttributeList) Init() tea.Cmd { return nil }

func (m AttributeList) KeyBindings() []KeyBinding {
	return []KeyBinding{
		{Key: "enter", Help: "view"},
		{Key: "/", Help: "filter"},
	}
}

func StyleAttr(attr string) string {
	return lipgloss.NewStyle().
		Foreground(constants.Magenta).
		Render(attr)
}

func CreateViewFormat(num int) string {
	var format string
	for i := 0; i < num; i++ {
		format += "%s %s\n"
	}
	return format
}

func (m AttributeList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	ctx := context.Background()

	switch msg := msg.(type) {
	case attributesLoadedMsg:
		m.loading = false
		if msg.err != nil {
			return m, func() tea.Msg {
				return StatusMsg{Text: "Error loading attributes: " + msg.err.Error(), IsError: true}
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
			return InitAttributeView(ctx, m.list.SelectedItem().(AttributeItem).id, m.h)
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m AttributeList) View() string {
	if m.loading {
		return lipgloss.NewStyle().
			Padding(1, 2).
			Render(m.spinner.View() + " Loading attributes…")
	}
	return ViewList(m.list)
}
