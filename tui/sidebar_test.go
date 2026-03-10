package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSidebar(t *testing.T) {
	t.Run("initial selection is Attributes (index 1)", func(t *testing.T) {
		s := NewSidebar()
		if s.selected != 1 {
			t.Errorf("expected default selection=1 (Attributes), got %d", s.selected)
		}
		if s.items[s.selected].title != "Attributes" {
			t.Errorf("expected 'Attributes' at index 1, got %q", s.items[s.selected].title)
		}
	})

	t.Run("down key moves selection forward", func(t *testing.T) {
		s := NewSidebar()
		s.selected = 0
		s, _ = s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
		if s.selected != 1 {
			t.Errorf("expected selection 1 after down, got %d", s.selected)
		}
	})

	t.Run("up key moves selection backward", func(t *testing.T) {
		s := NewSidebar()
		s.selected = 2
		s, _ = s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
		if s.selected != 1 {
			t.Errorf("expected selection 1 after up, got %d", s.selected)
		}
	})

	t.Run("up at top boundary stays at 0", func(t *testing.T) {
		s := NewSidebar()
		s.selected = 0
		s, _ = s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
		if s.selected != 0 {
			t.Errorf("selection should stay at 0, got %d", s.selected)
		}
	})

	t.Run("down at bottom boundary stays at max", func(t *testing.T) {
		s := NewSidebar()
		s.selected = len(s.items) - 1
		s, _ = s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
		if s.selected != len(s.items)-1 {
			t.Errorf("selection should stay at max, got %d", s.selected)
		}
	})

	t.Run("enter sends NavSelectMsg", func(t *testing.T) {
		s := NewSidebar()
		s.selected = 1
		_, cmd := s.Update(tea.KeyMsg{Type: tea.KeyEnter})
		if cmd == nil {
			t.Fatal("enter should produce a command")
		}
		msg := cmd()
		nav, ok := msg.(NavSelectMsg)
		if !ok {
			t.Fatalf("expected NavSelectMsg, got %T", msg)
		}
		if nav.item.title != "Attributes" {
			t.Errorf("expected Attributes nav item, got %q", nav.item.title)
		}
	})

	t.Run("view contains all resource group and item names", func(t *testing.T) {
		s := NewSidebar()
		s = s.SetHeight(40)
		v := s.View()
		// group header is title-case "Policy", items use their exact title strings
		for _, want := range []string{"Policy", "Namespaces", "Attributes", "KAS Registry"} {
			if !strings.Contains(v, want) {
				t.Errorf("sidebar view missing %q", want)
			}
		}
	})

	t.Run("SetFocused stores focused state", func(t *testing.T) {
		sf := NewSidebar().SetFocused(true)
		su := NewSidebar().SetFocused(false)
		if !sf.focused {
			t.Error("focused sidebar should have focused=true")
		}
		if su.focused {
			t.Error("unfocused sidebar should have focused=false")
		}
	})

	t.Run("selected item shows arrow indicator", func(t *testing.T) {
		s := NewSidebar()
		s.selected = 0
		s = s.SetHeight(40)
		v := s.View()
		if !strings.Contains(v, "▶") {
			t.Error("selected item should show ▶ indicator")
		}
	})

	t.Run("KeyBindings returns non-empty slice", func(t *testing.T) {
		s := NewSidebar()
		kb := s.KeyBindings()
		if len(kb) == 0 {
			t.Error("sidebar should advertise key bindings")
		}
	})
}
