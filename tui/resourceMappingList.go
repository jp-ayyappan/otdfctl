package tui

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/opentdf/otdfctl/tui/constants"
)

type ResourceMappingList struct {
	list    list.Model
	h       TUIHandler
	spinner spinner.Model
	loading bool
}

type ResourceMappingItem struct {
	id    string
	terms string
}

func (m ResourceMappingItem) FilterValue() string { return m.terms }
func (m ResourceMappingItem) Title() string       { return m.terms }
func (m ResourceMappingItem) Description() string { return m.id }

type resourceMappingsLoadedMsg struct {
	items []list.Item
	err   error
}

func loadResourceMappings(ctx context.Context, h TUIHandler) tea.Cmd {
	return func() tea.Msg {
		res, err := h.ListResourceMappings(ctx, 250, 0)
		if err != nil {
			return resourceMappingsLoadedMsg{err: err}
		}
		var items []list.Item
		for _, rm := range res.GetResourceMappings() {
			items = append(items, ResourceMappingItem{
				id:    rm.GetId(),
				terms: strings.Join(rm.GetTerms(), ", "),
			})
		}
		return resourceMappingsLoadedMsg{items: items}
	}
}

func InitResourceMappingList(ctx context.Context, h TUIHandler) (tea.Model, tea.Cmd) {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), constants.WindowSize.Width, constants.WindowSize.Height)
	l.Title = "Resource Mappings"

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorPrimary)

	m := ResourceMappingList{h: h, list: l, spinner: s, loading: true}
	return m, tea.Batch(s.Tick, loadResourceMappings(ctx, h))
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
	case resourceMappingsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			return m, func() tea.Msg {
				return StatusMsg{Text: "Error loading resource mappings: " + msg.err.Error(), IsError: true}
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
			return InitResourceMappingView(ctx, m.list.SelectedItem().(ResourceMappingItem).id, m.h)
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m ResourceMappingList) View() string {
	if m.loading {
		return lipgloss.NewStyle().Padding(1, 2).Render(m.spinner.View() + " Loading resource mappings…")
	}
	return ViewList(m.list)
}
