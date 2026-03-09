package tui

import (
	"context"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/opentdf/otdfctl/pkg/handlers"
	"github.com/opentdf/otdfctl/tui/constants"
)

type KASRegistryList struct {
	list    list.Model
	h       handlers.Handler
	spinner spinner.Model
	loading bool
}

type KASRegistryItem struct {
	id   string
	name string
	uri  string
}

func (m KASRegistryItem) FilterValue() string { return m.name + " " + m.uri }
func (m KASRegistryItem) Title() string {
	if m.name != "" {
		return m.name
	}
	return m.uri
}
func (m KASRegistryItem) Description() string { return m.uri }

type kasRegistryLoadedMsg struct {
	items []list.Item
	err   error
}

func loadKASRegistry(ctx context.Context, h handlers.Handler) tea.Cmd {
	return func() tea.Msg {
		res, err := h.ListKasRegistryEntries(ctx, 250, 0)
		if err != nil {
			return kasRegistryLoadedMsg{err: err}
		}
		var items []list.Item
		for _, kas := range res.GetKeyAccessServers() {
			items = append(items, KASRegistryItem{
				id:   kas.GetId(),
				name: kas.GetName(),
				uri:  kas.GetUri(),
			})
		}
		return kasRegistryLoadedMsg{items: items}
	}
}

func InitKASRegistryList(ctx context.Context, h handlers.Handler) (tea.Model, tea.Cmd) {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), constants.WindowSize.Width, constants.WindowSize.Height)
	l.Title = "KAS Registry"

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorPrimary)

	m := KASRegistryList{h: h, list: l, spinner: s, loading: true}
	return m, tea.Batch(s.Tick, loadKASRegistry(ctx, h))
}

func (m KASRegistryList) Init() tea.Cmd { return nil }

func (m KASRegistryList) KeyBindings() []KeyBinding {
	return []KeyBinding{
		{Key: "enter", Help: "view"},
		{Key: "/", Help: "filter"},
	}
}

func (m KASRegistryList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	ctx := context.Background()

	switch msg := msg.(type) {
	case kasRegistryLoadedMsg:
		m.loading = false
		if msg.err != nil {
			return m, func() tea.Msg {
				return StatusMsg{Text: "Error loading KAS registry: " + msg.err.Error(), IsError: true}
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
			return InitKASRegistryView(ctx, m.list.SelectedItem().(KASRegistryItem).id, m.h)
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m KASRegistryList) View() string {
	if m.loading {
		return lipgloss.NewStyle().Padding(1, 2).Render(m.spinner.View() + " Loading KAS registry…")
	}
	return ViewList(m.list)
}
