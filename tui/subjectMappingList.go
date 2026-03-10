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

type SubjectMappingList struct {
	list    list.Model
	h       TUIHandler
	spinner spinner.Model
	loading bool
}

type SubjectMappingItem struct {
	id       string
	attrVal  string
	actionCt int
}

func (m SubjectMappingItem) FilterValue() string { return m.attrVal }
func (m SubjectMappingItem) Title() string       { return m.attrVal }
func (m SubjectMappingItem) Description() string {
	return fmt.Sprintf("%s  (%d actions)", m.id, m.actionCt)
}

type subjectMappingsLoadedMsg struct {
	items []list.Item
	err   error
}

func loadSubjectMappings(ctx context.Context, h TUIHandler) tea.Cmd {
	return func() tea.Msg {
		res, err := h.ListSubjectMappings(ctx, 250, 0)
		if err != nil {
			return subjectMappingsLoadedMsg{err: err}
		}
		var items []list.Item
		for _, sm := range res.GetSubjectMappings() {
			items = append(items, SubjectMappingItem{
				id:       sm.GetId(),
				attrVal:  sm.GetAttributeValue().GetFqn(),
				actionCt: len(sm.GetActions()),
			})
		}
		return subjectMappingsLoadedMsg{items: items}
	}
}

func InitSubjectMappingList(ctx context.Context, h TUIHandler) (tea.Model, tea.Cmd) {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), constants.WindowSize.Width, constants.WindowSize.Height)
	l.Title = "Subject Mappings"

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorPrimary)

	m := SubjectMappingList{h: h, list: l, spinner: s, loading: true}
	return m, tea.Batch(s.Tick, loadSubjectMappings(ctx, h))
}

func (m SubjectMappingList) Init() tea.Cmd { return nil }

func (m SubjectMappingList) KeyBindings() []KeyBinding {
	return []KeyBinding{
		{Key: "enter", Help: "view"},
		{Key: "/", Help: "filter"},
	}
}

func (m SubjectMappingList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	ctx := context.Background()

	switch msg := msg.(type) {
	case subjectMappingsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			return m, func() tea.Msg {
				return StatusMsg{Text: "Error loading subject mappings: " + msg.err.Error(), IsError: true}
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
			return InitSubjectMappingView(ctx, m.list.SelectedItem().(SubjectMappingItem).id, m.h)
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m SubjectMappingList) View() string {
	if m.loading {
		return lipgloss.NewStyle().Padding(1, 2).Render(m.spinner.View() + " Loading subject mappings…")
	}
	return ViewList(m.list)
}
