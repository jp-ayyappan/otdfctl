package tui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/opentdf/otdfctl/tui/constants"
	"github.com/opentdf/platform/protocol/go/policy"
)

// AttrValueList lists the values for a single attribute, drilled into from AttributeView.
type AttrValueList struct {
	attr    *policy.Attribute
	list    list.Model
	h       TUIHandler
	spinner spinner.Model
	loading bool
}

type AttrValueItem struct {
	id    string
	value string
	fqn   string
}

func (m AttrValueItem) FilterValue() string { return m.value }
func (m AttrValueItem) Title() string       { return m.value }
func (m AttrValueItem) Description() string { return m.fqn }

type attrValuesLoadedMsg struct {
	items []*policy.Value
	err   error
}

func loadAttrValues(ctx context.Context, h TUIHandler, attrID string) tea.Cmd {
	return func() tea.Msg {
		vals, err := h.ListAttributeValues(ctx, attrID)
		return attrValuesLoadedMsg{items: vals, err: err}
	}
}

func InitAttrValueList(ctx context.Context, attr *policy.Attribute, h TUIHandler) (tea.Model, tea.Cmd) {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), constants.WindowSize.Width, constants.WindowSize.Height)
	l.Title = fmt.Sprintf("Values: %s", attr.GetName())

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorPrimary)

	m := AttrValueList{attr: attr, h: h, list: l, spinner: s, loading: true}
	return m, tea.Batch(s.Tick, loadAttrValues(ctx, h, attr.GetId()))
}

func (m AttrValueList) Init() tea.Cmd { return nil }

func (m AttrValueList) KeyBindings() []KeyBinding {
	return []KeyBinding{
		{Key: "enter", Help: "view"},
		{Key: "/", Help: "filter"},
		{Key: "backspace", Help: "back"},
	}
}

func (m AttrValueList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	ctx := context.Background()

	switch msg := msg.(type) {
	case attrValuesLoadedMsg:
		m.loading = false
		if msg.err != nil {
			return m, func() tea.Msg {
				return StatusMsg{Text: "Error loading attribute values: " + msg.err.Error(), IsError: true}
			}
		}
		var items []list.Item
		for _, v := range msg.items {
			items = append(items, AttrValueItem{
				id:    v.GetId(),
				value: v.GetValue(),
				fqn:   v.GetFqn(),
			})
		}
		m.list.SetItems(items)
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
			return InitAttributeView(ctx, m.attr.GetId(), m.h)
		case "enter", "e":
			if m.loading || len(m.list.Items()) == 0 {
				return m, nil
			}
			selected := m.list.SelectedItem().(AttrValueItem)
			return InitAttrValueView(ctx, selected, m.attr, m.h)
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m AttrValueList) View() string {
	if m.loading {
		return lipgloss.NewStyle().Padding(1, 2).Render(m.spinner.View() + " Loading values…")
	}
	return ViewList(m.list)
}

// ---- AttrValueView ----

type AttrValueView struct {
	item AttrValueItem
	attr *policy.Attribute
	read Read
	h    TUIHandler
}

func InitAttrValueView(ctx context.Context, item AttrValueItem, attr *policy.Attribute, h TUIHandler) (tea.Model, tea.Cmd) {
	val, _ := h.GetAttributeValue(ctx, item.id)

	active := "false"
	if val.GetActive().GetValue() {
		active = "true"
	}

	items := []list.Item{
		AttributeSubItem{title: "ID", description: val.GetId()},
		AttributeSubItem{title: "Value", description: val.GetValue()},
		AttributeSubItem{title: "FQN", description: val.GetFqn()},
		AttributeSubItem{title: "Active", description: active},
	}

	model, _ := InitRead(fmt.Sprintf("Value: %s", val.GetValue()), items)
	mod := model.(Read)
	m := AttrValueView{item: item, attr: attr, h: h, read: mod}
	model, msg := m.Update(WindowMsg())
	return model, msg
}

func (m AttrValueView) Init() tea.Cmd { return nil }

func (m AttrValueView) KeyBindings() []KeyBinding {
	return []KeyBinding{
		{Key: "backspace", Help: "back to values"},
	}
}

func (m AttrValueView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	ctx := context.Background()

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		constants.WindowSize = msg
		m.read.list.SetSize(msg.Width, msg.Height)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "backspace":
			return InitAttrValueList(ctx, m.attr, m.h)
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.read.list, cmd = m.read.list.Update(msg)
	return m, cmd
}

func (m AttrValueView) View() string {
	return m.read.View()
}
