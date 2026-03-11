package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/opentdf/otdfctl/tui/constants"
	"github.com/opentdf/platform/protocol/go/entity"
	ersv2 "github.com/opentdf/platform/protocol/go/entityresolution/v2"
)

// ersResolvedMsg carries the ERS response back to the model.
type ersResolvedMsg struct {
	reps []*ersv2.EntityRepresentation
	err  error
}

// ---- ERSInputView — entity identifier input form ----

// ERSInputView collects an entity identifier and, on submit, loads results.
type ERSInputView struct {
	input      textinput.Model
	inputType  string // "email", "client-id", "username"
	typeToggle int    // cycles through types
	h          TUIHandler
	width      int
	height     int
}

var ersInputTypes = []string{"email", "client-id", "username"}

func NewERSInputView(h TUIHandler) ERSInputView {
	ti := textinput.New()
	ti.Placeholder = "alice@example.com"
	ti.Focus()
	ti.Width = 50
	ti.Prompt = "  "

	return ERSInputView{
		input:     ti,
		inputType: ersInputTypes[0],
		h:         h,
	}
}

func (m ERSInputView) Init() tea.Cmd { return textinput.Blink }

func (m ERSInputView) KeyBindings() []KeyBinding {
	return []KeyBinding{
		{Key: "enter", Help: "resolve"},
		{Key: "tab", Help: "toggle type"},
		{Key: "backspace", Help: "clear"},
	}
}

func (m ERSInputView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "tab":
			m.typeToggle = (m.typeToggle + 1) % len(ersInputTypes)
			m.inputType = ersInputTypes[m.typeToggle]
			m.input.Placeholder = ersInputPlaceholder(m.inputType)
			return m, nil

		case "enter":
			val := strings.TrimSpace(m.input.Value())
			if val == "" {
				return m, nil
			}
			ent := buildEntity(m.inputType, val)
			results := NewERSResultView([]string{val}, m.h)
			return results, func() tea.Msg {
				resp, err := m.h.ResolveEntities(context.Background(), []*entity.Entity{ent})
				if err != nil {
					return ersResolvedMsg{err: err}
				}
				return ersResolvedMsg{reps: resp.GetEntityRepresentations()}
			}
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m ERSInputView) View() string {
	typeLabel := lipgloss.NewStyle().
		Foreground(ColorAccent).Bold(true).
		Render(m.inputType)

	tabHint := lipgloss.NewStyle().Foreground(ColorDim).
		Render("  tab to switch type")

	inputRow := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Padding(0, 1).
		Width(54).
		Render(m.input.View())

	body := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Foreground(ColorMuted).Render("Resolve Entity Attributes via ERS"),
		"",
		"  Type: "+typeLabel+tabHint,
		"",
		inputRow,
		"",
		lipgloss.NewStyle().Foreground(ColorDim).
			Render("  enter to resolve  •  ctrl+c to quit"),
	)

	return lipgloss.NewStyle().
		Width(m.width).Height(m.height).
		Padding(2, 4).
		Render(body)
}

// ---- ERSResultView — entitlement list after resolution ----

type ERSResultView struct {
	entities []string
	list     list.Model
	spinner  spinner.Model
	loading  bool
	h        TUIHandler
}

type ERSEntitlementItem struct {
	entity  string
	attrFQN string
	actions string
}

func (e ERSEntitlementItem) FilterValue() string { return e.entity + " " + e.attrFQN }
func (e ERSEntitlementItem) Title() string       { return e.attrFQN }
func (e ERSEntitlementItem) Description() string {
	return fmt.Sprintf("entity: %s  actions: %s", e.entity, e.actions)
}

func NewERSResultView(entities []string, h TUIHandler) ERSResultView {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), constants.WindowSize.Width, constants.WindowSize.Height)
	l.Title = "ERS — Entitlements"

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorPrimary)

	return ERSResultView{entities: entities, h: h, list: l, spinner: s, loading: true}
}

func (m ERSResultView) Init() tea.Cmd { return m.spinner.Tick }

func (m ERSResultView) KeyBindings() []KeyBinding {
	return []KeyBinding{
		{Key: "/", Help: "filter"},
		{Key: "backspace", Help: "new search"},
	}
}

func (m ERSResultView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ersResolvedMsg:
		m.loading = false
		if msg.err != nil {
			return m, func() tea.Msg {
				return StatusMsg{Text: "ERS error: " + msg.err.Error(), IsError: true}
			}
		}
		var items []list.Item
		for _, rep := range msg.reps {
			id := rep.GetOriginalId()
			for _, e := range rep.GetDirectEntitlements() {
				items = append(items, ERSEntitlementItem{
					entity:  id,
					attrFQN: e.GetAttributeValueFqn(),
					actions: strings.Join(e.GetActions(), ", "),
				})
			}
		}
		if len(items) == 0 {
			return m, func() tea.Msg {
				return StatusMsg{Text: "No entitlements found for the given entity.", IsError: false}
			}
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
			return NewERSInputView(m.h), textinput.Blink
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m ERSResultView) View() string {
	if m.loading {
		label := strings.Join(m.entities, ", ")
		return lipgloss.NewStyle().Padding(1, 2).
			Render(m.spinner.View() + fmt.Sprintf(" Resolving %s…", label))
	}
	return ViewList(m.list)
}

// ---- helpers ----

func buildEntity(inputType, value string) *entity.Entity {
	switch inputType {
	case "client-id":
		return &entity.Entity{EntityType: &entity.Entity_ClientId{ClientId: value}}
	case "username":
		return &entity.Entity{EntityType: &entity.Entity_UserName{UserName: value}}
	default: // email
		return &entity.Entity{EntityType: &entity.Entity_EmailAddress{EmailAddress: value}}
	}
}

func ersInputPlaceholder(inputType string) string {
	switch inputType {
	case "client-id":
		return "my-service-account"
	case "username":
		return "alice"
	default:
		return "alice@example.com"
	}
}
