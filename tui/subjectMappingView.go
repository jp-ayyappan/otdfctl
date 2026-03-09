package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/opentdf/otdfctl/pkg/handlers"
	"github.com/opentdf/otdfctl/tui/constants"
	"github.com/opentdf/platform/protocol/go/policy"
)

type SubjectMappingView struct {
	sm   *policy.SubjectMapping
	read Read
	h    handlers.Handler
}

func InitSubjectMappingView(ctx context.Context, id string, h handlers.Handler) (tea.Model, tea.Cmd) {
	sm, _ := h.GetSubjectMapping(ctx, id)

	var actions []string
	for _, a := range sm.GetActions() {
		actions = append(actions, a.GetName())
	}

	var scs string
	if sm.GetSubjectConditionSet() != nil {
		scs = sm.GetSubjectConditionSet().GetId()
	}

	items := []list.Item{
		AttributeSubItem{title: "ID", description: sm.GetId()},
		AttributeSubItem{title: "Attribute Value", description: sm.GetAttributeValue().GetFqn()},
		AttributeSubItem{title: "Actions", description: strings.Join(actions, ", ")},
		AttributeSubItem{title: "Subject Condition Set", description: scs},
	}
	for i, ss := range sm.GetSubjectConditionSet().GetSubjectSets() {
		for j, sc := range ss.GetConditionGroups() {
			for k, c := range sc.GetConditions() {
				label := fmt.Sprintf("Condition [%d][%d][%d]", i, j, k)
				desc := fmt.Sprintf("%s %s %v", c.GetSubjectExternalSelectorValue(), c.GetOperator(), c.GetSubjectExternalValues())
				items = append(items, AttributeSubItem{title: label, description: desc})
			}
		}
	}

	model, _ := InitRead("Subject Mapping Detail", items)
	mod := model.(Read)
	m := SubjectMappingView{h: h, sm: sm, read: mod}
	model, msg := m.Update(WindowMsg())
	return model, msg
}

func (m SubjectMappingView) Init() tea.Cmd { return nil }

func (m SubjectMappingView) KeyBindings() []KeyBinding {
	return []KeyBinding{
		{Key: "backspace", Help: "back"},
	}
}

func (m SubjectMappingView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	ctx := context.Background()

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		constants.WindowSize = msg
		m.read.list.SetSize(msg.Width, msg.Height)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "backspace":
			return InitSubjectMappingList(ctx, m.h)
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.read.list, cmd = m.read.list.Update(msg)
	return m, cmd
}

func (m SubjectMappingView) View() string {
	return m.read.View()
}
