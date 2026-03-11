package tui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/opentdf/otdfctl/tui/constants"
)

type SCSList struct {
	list    list.Model
	h       TUIHandler
	spinner spinner.Model
	loading bool
}

type SCSItem struct {
	id         string
	setCount   int
	condCount  int
}

func (m SCSItem) FilterValue() string { return m.id }
func (m SCSItem) Title() string       { return m.id }
func (m SCSItem) Description() string {
	return fmt.Sprintf("%d subject sets  %d total conditions", m.setCount, m.condCount)
}

type scsLoadedMsg struct {
	items []list.Item
	err   error
}

func loadSCS(ctx context.Context, h TUIHandler) tea.Cmd {
	return func() tea.Msg {
		res, err := h.ListSubjectConditionSets(ctx, 250, 0)
		if err != nil {
			return scsLoadedMsg{err: err}
		}
		var items []list.Item
		for _, scs := range res.GetSubjectConditionSets() {
			conds := 0
			for _, ss := range scs.GetSubjectSets() {
				for _, cg := range ss.GetConditionGroups() {
					conds += len(cg.GetConditions())
				}
			}
			items = append(items, SCSItem{
				id:        scs.GetId(),
				setCount:  len(scs.GetSubjectSets()),
				condCount: conds,
			})
		}
		return scsLoadedMsg{items: items}
	}
}

func InitSCSList(ctx context.Context, h TUIHandler) (tea.Model, tea.Cmd) {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), constants.WindowSize.Width, constants.WindowSize.Height)
	l.Title = "Subject Condition Sets"

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorPrimary)

	m := SCSList{h: h, list: l, spinner: s, loading: true}
	return m, tea.Batch(s.Tick, loadSCS(ctx, h))
}

func (m SCSList) Init() tea.Cmd { return nil }

func (m SCSList) KeyBindings() []KeyBinding {
	return []KeyBinding{
		{Key: "enter", Help: "view"},
		{Key: "/", Help: "filter"},
	}
}

func (m SCSList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	ctx := context.Background()

	switch msg := msg.(type) {
	case scsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			return m, func() tea.Msg {
				return StatusMsg{Text: "Error loading subject condition sets: " + msg.err.Error(), IsError: true}
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
			id := m.list.SelectedItem().(SCSItem).id
			return InitSCSView(ctx, id, m.h, nil)
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m SCSList) View() string {
	if m.loading {
		return lipgloss.NewStyle().Padding(1, 2).Render(m.spinner.View() + " Loading subject condition sets…")
	}
	return ViewList(m.list)
}
