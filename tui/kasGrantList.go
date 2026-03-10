package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/opentdf/otdfctl/tui/constants"
)

// KASGrantList shows all key→policy-object mappings (key grants).
// Each row is one KAS key with the namespaces, attributes, and values it covers.
type KASGrantList struct {
	list    list.Model
	h       TUIHandler
	spinner spinner.Model
	loading bool
}

type KASGrantItem struct {
	kid        string
	kasURI     string
	namespaces []string
	attributes []string
	values     []string
}

func (m KASGrantItem) FilterValue() string {
	return m.kid + " " + m.kasURI + " " + strings.Join(m.values, " ")
}

func (m KASGrantItem) Title() string {
	if m.kid != "" {
		return fmt.Sprintf("%s  @  %s", m.kid, m.kasURI)
	}
	return m.kasURI
}

func (m KASGrantItem) Description() string {
	parts := []string{}
	if len(m.namespaces) > 0 {
		parts = append(parts, fmt.Sprintf("ns:%d", len(m.namespaces)))
	}
	if len(m.attributes) > 0 {
		parts = append(parts, fmt.Sprintf("attr:%d", len(m.attributes)))
	}
	if len(m.values) > 0 {
		parts = append(parts, fmt.Sprintf("val:%d", len(m.values)))
	}
	if len(parts) == 0 {
		return "(no mappings)"
	}
	return strings.Join(parts, "  ")
}

type kasGrantsLoadedMsg struct {
	items []list.Item
	err   error
}

func loadKASGrants(ctx context.Context, h TUIHandler) tea.Cmd {
	return func() tea.Msg {
		res, err := h.ListKeyMappings(ctx, 250, 0, "", nil)
		if err != nil {
			return kasGrantsLoadedMsg{err: err}
		}
		var items []list.Item
		for _, km := range res.GetKeyMappings() {
			item := KASGrantItem{
				kid:    km.GetKid(),
				kasURI: km.GetKasUri(),
			}
			for _, n := range km.GetNamespaceMappings() {
				item.namespaces = append(item.namespaces, n.GetFqn())
			}
			for _, a := range km.GetAttributeMappings() {
				item.attributes = append(item.attributes, a.GetFqn())
			}
			for _, v := range km.GetValueMappings() {
				item.values = append(item.values, v.GetFqn())
			}
			items = append(items, item)
		}
		return kasGrantsLoadedMsg{items: items}
	}
}

func InitKASGrantList(ctx context.Context, h TUIHandler) (tea.Model, tea.Cmd) {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), constants.WindowSize.Width, constants.WindowSize.Height)
	l.Title = "KAS Grants (Key Mappings)"

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorPrimary)

	m := KASGrantList{h: h, list: l, spinner: s, loading: true}
	return m, tea.Batch(s.Tick, loadKASGrants(ctx, h))
}

func (m KASGrantList) Init() tea.Cmd { return nil }

func (m KASGrantList) KeyBindings() []KeyBinding {
	return []KeyBinding{
		{Key: "enter", Help: "view detail"},
		{Key: "/", Help: "filter"},
	}
}

func (m KASGrantList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case kasGrantsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			return m, func() tea.Msg {
				return StatusMsg{Text: "Error loading KAS grants: " + msg.err.Error(), IsError: true}
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
			selected := m.list.SelectedItem().(KASGrantItem)
			return InitKASGrantView(selected, m.h)
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m KASGrantList) View() string {
	if m.loading {
		return lipgloss.NewStyle().Padding(1, 2).Render(m.spinner.View() + " Loading KAS grants…")
	}
	return ViewList(m.list)
}

// ---- KASGrantView ---- shows full mapping detail for one key

type KASGrantView struct {
	item KASGrantItem
	read Read
	h    TUIHandler
}

func InitKASGrantView(item KASGrantItem, h TUIHandler) (tea.Model, tea.Cmd) {
	mkObj := func(label, val string) list.Item {
		return AttributeSubItem{title: label, description: val}
	}

	items := []list.Item{
		mkObj("KID", item.kid),
		mkObj("KAS URI", item.kasURI),
		mkObj("Namespaces", fmtList(item.namespaces)),
		mkObj("Attributes", fmtList(item.attributes)),
		mkObj("Values", fmtList(item.values)),
	}

	model, _ := InitRead(fmt.Sprintf("Grant: %s", item.kid), items)
	mod := model.(Read)
	m := KASGrantView{item: item, h: h, read: mod}
	model, msg := m.Update(WindowMsg())
	return model, msg
}

func (m KASGrantView) Init() tea.Cmd { return nil }

func (m KASGrantView) KeyBindings() []KeyBinding {
	return []KeyBinding{
		{Key: "backspace", Help: "back"},
	}
}

func (m KASGrantView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	ctx := context.Background()

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		constants.WindowSize = msg
		m.read.list.SetSize(msg.Width, msg.Height)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "backspace":
			return InitKASGrantList(ctx, m.h)
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.read.list, cmd = m.read.list.Update(msg)
	return m, cmd
}

func (m KASGrantView) View() string {
	return m.read.View()
}

// fmtList formats a string slice as a comma-separated string or "(none)".
func fmtList(ss []string) string {
	if len(ss) == 0 {
		return "(none)"
	}
	return strings.Join(ss, ", ")
}

