package tui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/opentdf/otdfctl/tui/constants"
	"github.com/opentdf/platform/protocol/go/policy"
)

type SubjectMappingView struct {
	sm   *policy.SubjectMapping
	read Read
	h    TUIHandler
}

func InitSubjectMappingView(ctx context.Context, id string, h TUIHandler) (tea.Model, tea.Cmd) {
	sm, _ := h.GetSubjectMapping(ctx, id)
	smID := sm.GetId()

	var items []list.Item
	items = append(items, AttributeSubItem{title: "ID", description: smID})

	// Attribute Value — navigable chip
	av := sm.GetAttributeValue()
	avID := av.GetId()
	items = append(items, NewChip("Attribute Value", av.GetFqn(),
		func(ctx context.Context, h TUIHandler) (tea.Model, tea.Cmd) {
			val, _ := h.GetAttributeValue(ctx, avID)
			back := func(ctx context.Context, h TUIHandler) (tea.Model, tea.Cmd) {
				return InitSubjectMappingView(ctx, smID, h)
			}
			return initAttrValueViewWithBack(ctx, AttrValueItem{
				id: val.GetId(), value: val.GetValue(), fqn: val.GetFqn(),
			}, val.GetAttribute(), h, back)
		},
	))

	// Actions — info chips (names only; actions are not a complex navigable object)
	if len(sm.GetActions()) == 0 {
		items = append(items, AttributeSubItem{title: "Actions", description: "(none)"})
	}
	for i, a := range sm.GetActions() {
		items = append(items, NewInfoChip(fmt.Sprintf("Action %d", i+1), a.GetName()))
	}

	// Subject Condition Set — navigable chip
	if scs := sm.GetSubjectConditionSet(); scs != nil {
		scsID := scs.GetId()
		items = append(items, NewChip("Subject Condition Set", scsID,
			func(ctx context.Context, h TUIHandler) (tea.Model, tea.Cmd) {
				back := func(ctx context.Context, h TUIHandler) (tea.Model, tea.Cmd) {
					return InitSubjectMappingView(ctx, smID, h)
				}
				return InitSCSView(ctx, scsID, h, back)
			},
		))
	} else {
		items = append(items, AttributeSubItem{title: "Subject Condition Set", description: "(none)"})
	}

	model, _ := InitRead("Subject Mapping", items)
	mod := model.(Read)
	m := SubjectMappingView{h: h, sm: sm, read: mod}
	model, msg := m.Update(WindowMsg())
	m = model.(SubjectMappingView)
	return m, msg
}

func (m SubjectMappingView) Init() tea.Cmd { return nil }

func (m SubjectMappingView) KeyBindings() []KeyBinding {
	return []KeyBinding{
		{Key: "enter", Help: "drill in"},
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
		case "enter":
			if chip, ok := m.read.list.SelectedItem().(ChipItem); ok && chip.Navigable() {
				return chip.navigate(ctx, m.h)
			}
		}
	}

	var cmd tea.Cmd
	m.read.list, cmd = m.read.list.Update(msg)
	return m, cmd
}

func (m SubjectMappingView) View() string {
	return m.read.View()
}

// initAttrValueViewWithBack creates a value detail view whose backspace
// navigates to back() instead of defaulting to AttrValueList.
func initAttrValueViewWithBack(ctx context.Context, item AttrValueItem, attr *policy.Attribute, h TUIHandler, back backFn) (tea.Model, tea.Cmd) {
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
	bv := backableValueView{item: item, attr: attr, h: h, read: mod, backTo: back}
	model, out := bv.Update(WindowMsg())
	return model, out
}

// backableValueView is an AttrValueView with a configurable back destination,
// used when navigating to a value via a chip from another resource's detail.
type backableValueView struct {
	item   AttrValueItem
	attr   *policy.Attribute
	h      TUIHandler
	read   Read
	backTo backFn
}

func (b backableValueView) Init() tea.Cmd { return nil }

func (b backableValueView) KeyBindings() []KeyBinding {
	return []KeyBinding{{Key: "backspace", Help: "back"}}
}

func (b backableValueView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	ctx := context.Background()
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		constants.WindowSize = msg
		b.read.list.SetSize(msg.Width, msg.Height)
		return b, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "backspace":
			if b.backTo != nil {
				return b.backTo(ctx, b.h)
			}
			return InitAttrValueList(ctx, b.attr, b.h)
		case "ctrl+c", "q":
			return b, tea.Quit
		}
	}
	var cmd tea.Cmd
	b.read.list, cmd = b.read.list.Update(msg)
	return b, cmd
}

func (b backableValueView) View() string { return b.read.View() }
