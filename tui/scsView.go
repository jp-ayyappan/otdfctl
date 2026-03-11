package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/opentdf/otdfctl/tui/constants"
	"github.com/opentdf/platform/protocol/go/policy"
)

// backFn is the navigation target for the backspace key.
// nil means "go to SCSList".
type backFn func(ctx context.Context, h TUIHandler) (tea.Model, tea.Cmd)

// SCSView renders a SubjectConditionSet as a flat list of conditions,
// grouped by SubjectSet and ConditionGroup.
type SCSView struct {
	scs    *policy.SubjectConditionSet
	read   Read
	h      TUIHandler
	backTo backFn
}

func InitSCSView(ctx context.Context, id string, h TUIHandler, back backFn) (tea.Model, tea.Cmd) {
	scs, _ := h.GetSubjectConditionSet(ctx, id)
	items := buildSCSItems(scs)

	model, _ := InitRead(fmt.Sprintf("SCS: %s", truncate(id, 36)), items)
	mod := model.(Read)
	m := SCSView{scs: scs, h: h, read: mod, backTo: back}
	model, msg := m.Update(WindowMsg())
	return model, msg
}

// buildSCSItems flattens the nested SubjectSet → ConditionGroup → Condition
// tree into a flat list of readable rows.
func buildSCSItems(scs *policy.SubjectConditionSet) []list.Item {
	var items []list.Item

	items = append(items, AttributeSubItem{title: "ID", description: scs.GetId()})

	for si, ss := range scs.GetSubjectSets() {
		// Subject set header
		items = append(items, AttributeSubItem{
			title:       fmt.Sprintf("Subject Set %d", si+1),
			description: fmt.Sprintf("%d condition group(s)", len(ss.GetConditionGroups())),
		})

		for gi, cg := range ss.GetConditionGroups() {
			boolOp := cg.GetBooleanOperator().String()
			boolOp = strings.TrimPrefix(boolOp, "CONDITION_BOOLEAN_TYPE_ENUM_")

			// Condition group header
			items = append(items, AttributeSubItem{
				title:       fmt.Sprintf("  Group %d", gi+1),
				description: fmt.Sprintf("operator: %s  %d condition(s)", boolOp, len(cg.GetConditions())),
			})

			for ci, c := range cg.GetConditions() {
				op := c.GetOperator().String()
				op = strings.TrimPrefix(op, "SUBJECT_MAPPING_OPERATOR_ENUM_")

				vals := c.GetSubjectExternalValues()
				valStr := strings.Join(vals, ", ")
				if len(vals) > 3 {
					valStr = strings.Join(vals[:3], ", ") + fmt.Sprintf(" … +%d", len(vals)-3)
				}

				items = append(items, AttributeSubItem{
					title:       fmt.Sprintf("    [S%d/G%d/C%d] %s", si+1, gi+1, ci+1, c.GetSubjectExternalSelectorValue()),
					description: fmt.Sprintf("%s  [%s]", op, valStr),
				})
			}
		}
	}

	return items
}

func (m SCSView) Init() tea.Cmd { return nil }

func (m SCSView) KeyBindings() []KeyBinding {
	return []KeyBinding{
		{Key: "backspace", Help: "back"},
	}
}

func (m SCSView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	ctx := context.Background()

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		constants.WindowSize = msg
		m.read.list.SetSize(msg.Width, msg.Height)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "backspace":
			if m.backTo != nil {
				return m.backTo(ctx, m.h)
			}
			return InitSCSList(ctx, m.h)
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.read.list, cmd = m.read.list.Update(msg)
	return m, cmd
}

func (m SCSView) View() string {
	return m.read.View()
}

// truncate shortens a string to maxLen with "…" suffix.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "…"
}
