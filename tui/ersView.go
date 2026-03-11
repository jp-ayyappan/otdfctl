package tui

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/opentdf/otdfctl/tui/constants"
	"github.com/opentdf/platform/protocol/go/entity"
	ersv2 "github.com/opentdf/platform/protocol/go/entityresolution/v2"
	"google.golang.org/protobuf/types/known/structpb"
)

// ersResolvedMsg carries the ERS response back to the model.
type ersResolvedMsg struct {
	reps     []*ersv2.EntityRepresentation
	entities []*entity.Entity // original inputs for mapping entity_idx_N → identifier
	err      error
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
		{Key: "ctrl+t", Help: "toggle type"},
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

		case "ctrl+t":
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
			inputs := []*entity.Entity{ent}
			results := NewERSResultView([]string{val}, m.h)
			return results, func() tea.Msg {
				resp, err := m.h.ResolveEntities(context.Background(), inputs)
				if err != nil {
					return ersResolvedMsg{err: err, entities: inputs}
				}
				return ersResolvedMsg{reps: resp.GetEntityRepresentations(), entities: inputs}
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
		Render("  ctrl+t to switch type")

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

// ERSRowItem is a generic display row in the ERS result list.
type ERSRowItem struct {
	section string // "id", "idp-attr", "entitlement", "header"
	key     string
	value   string
}

func (r ERSRowItem) FilterValue() string { return r.key + " " + r.value }

func (r ERSRowItem) Title() string {
	switch r.section {
	case "header":
		return lipgloss.NewStyle().Bold(true).Foreground(ColorAccent).Render("▸ " + r.key)
	case "entitlement":
		return lipgloss.NewStyle().Foreground(ColorSuccess).Render("⊕ ") + r.key
	default:
		return "  " + r.key
	}
}

func (r ERSRowItem) Description() string { return r.value }

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
		items := buildERSItems(msg.reps, msg.entities)
		m.list.SetItems(items)
		if len(items) == 0 {
			return m, func() tea.Msg {
				return StatusMsg{Text: "ERS returned no data for the given entity.", IsError: false}
			}
		}
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

// internalERSKeys are Keycloak/LDAP bookkeeping fields we skip in TUI output.
var internalERSKeys = map[string]bool{
	"LDAP_ENTRY_DN":   true,
	"LDAP_ID":         true,
	"createTimestamp": true,
	"modifyTimestamp": true,
}

// buildERSItems converts ERS response representations into flat list items,
// showing identity fields, IdP attributes, and OpenTDF entitlements per entity.
func buildERSItems(reps []*ersv2.EntityRepresentation, inputs []*entity.Entity) []list.Item {
	var items []list.Item

	for i, rep := range reps {
		// Resolve the human-readable label for this entity
		label := ersEntityLabel(rep.GetOriginalId(), inputs, i)

		// Entity header
		items = append(items, ERSRowItem{section: "header", key: label})

		// Identity fields from additional_props
		idKeys := []string{"email", "username", "firstName", "lastName", "id"}
		for _, prop := range rep.GetAdditionalProps() {
			fields := prop.GetFields()
			for _, k := range idKeys {
				if v, ok := fields[k]; ok && v.GetStringValue() != "" {
					items = append(items, ERSRowItem{section: "id", key: k, value: v.GetStringValue()})
				}
			}
			// Custom attributes nested under "attributes"
			if attrField, ok := fields["attributes"]; ok {
				if attrStruct := attrField.GetStructValue(); attrStruct != nil {
					keys := make([]string, 0, len(attrStruct.GetFields()))
					for k := range attrStruct.GetFields() {
						keys = append(keys, k)
					}
					sort.Strings(keys)
					for _, k := range keys {
						if internalERSKeys[k] {
							continue
						}
						var vals []string
						for _, item := range attrStruct.GetFields()[k].GetListValue().GetValues() {
							vals = append(vals, item.GetStringValue())
						}
						if len(vals) > 0 {
							items = append(items, ERSRowItem{
								section: "idp-attr",
								key:     k,
								value:   strings.Join(vals, ", "),
							})
						}
					}
				}
			}
		}

		// OpenTDF entitlements
		entitlements := rep.GetDirectEntitlements()
		if len(entitlements) == 0 {
			items = append(items, ERSRowItem{
				section: "entitlement",
				key:     "No OpenTDF entitlements",
				value:   "check subject mappings",
			})
		} else {
			for _, e := range entitlements {
				items = append(items, ERSRowItem{
					section: "entitlement",
					key:     e.GetAttributeValueFqn(),
					value:   strings.Join(e.GetActions(), ", "),
				})
			}
		}
	}
	return items
}

// ersEntityLabel maps the ephemeral entity_idx_N back to the original input identifier.
func ersEntityLabel(originalID string, inputs []*entity.Entity, idx int) string {
	if idx < len(inputs) {
		switch et := inputs[idx].GetEntityType().(type) {
		case *entity.Entity_EmailAddress:
			return et.EmailAddress
		case *entity.Entity_ClientId:
			return et.ClientId
		case *entity.Entity_UserName:
			return et.UserName
		}
	}
	return originalID
}

// Keep structpb in scope — used via attrField.GetStructValue() above.
var _ *structpb.Struct
