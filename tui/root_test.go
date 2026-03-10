package tui

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// newTestRoot returns a Root with a placeholder as content (no live API needed).
func newTestRoot(h TUIHandler) Root {
	m := Root{
		h:         h,
		header:    NewHeader(h.GetProfileName(), h.GetEndpoint()),
		sidebar:   NewSidebar(),
		content:   NewPlaceholder("Test"),
		statusbar: NewStatusBar(),
		keybar:    NewKeybar(globalBindings),
		focus:     focusContent,
	}
	return m
}

func TestRoot(t *testing.T) {
	resize := tea.WindowSizeMsg{Width: 120, Height: 40}

	t.Run("window resize updates dimensions", func(t *testing.T) {
		m := newTestRoot(newMock())
		model, _ := m.Update(resize)
		r := model.(Root)
		if r.width != 120 || r.height != 40 {
			t.Errorf("expected 120x40, got %dx%d", r.width, r.height)
		}
	})

	t.Run("view returns empty string before first resize", func(t *testing.T) {
		m := newTestRoot(newMock())
		if m.View() != "" {
			t.Error("view should be empty before window size is known")
		}
	})

	t.Run("view renders after resize", func(t *testing.T) {
		m := newTestRoot(newMock())
		model, _ := m.Update(resize)
		r := model.(Root)
		v := r.View()
		if v == "" {
			t.Error("view should be non-empty after resize")
		}
		if !strings.Contains(v, "OpenTDF") {
			t.Error("view should contain app name in header")
		}
	})

	t.Run("tab switches focus from content to sidebar", func(t *testing.T) {
		m := newTestRoot(newMock())
		m.focus = focusContent
		m.width, m.height = 120, 40
		model, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
		r := model.(Root)
		if r.focus != focusSidebar {
			t.Error("tab should switch focus to sidebar")
		}
		if !r.sidebar.focused {
			t.Error("sidebar.focused should be true")
		}
	})

	t.Run("tab switches focus from sidebar back to content", func(t *testing.T) {
		m := newTestRoot(newMock())
		m.focus = focusSidebar
		m.sidebar = m.sidebar.SetFocused(true)
		m.width, m.height = 120, 40
		model, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
		r := model.(Root)
		if r.focus != focusContent {
			t.Error("tab should switch focus back to content")
		}
	})

	t.Run("ctrl+c returns quit command", func(t *testing.T) {
		m := newTestRoot(newMock())
		_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
		if cmd == nil {
			t.Fatal("ctrl+c should return a command")
		}
		msg := cmd()
		if _, ok := msg.(tea.QuitMsg); !ok {
			t.Errorf("ctrl+c should produce QuitMsg, got %T", msg)
		}
	})

	t.Run("NavSelectMsg swaps content and focuses content", func(t *testing.T) {
		m := newTestRoot(newMock())
		m.width, m.height = 120, 40
		m.focus = focusSidebar
		ph := NewPlaceholder("Namespaces")
		navMsg := NavSelectMsg{item: navItem{
			title: "Namespaces",
			load: func(_ context.Context, _ TUIHandler) (tea.Model, tea.Cmd) {
				return ph, nil
			},
		}}
		model, _ := m.Update(navMsg)
		r := model.(Root)
		if r.focus != focusContent {
			t.Error("after NavSelect, focus should be content")
		}
		if _, ok := r.content.(Placeholder); !ok {
			t.Errorf("content should be Placeholder after NavSelect, got %T", r.content)
		}
	})

	t.Run("StatusMsg is displayed in statusbar", func(t *testing.T) {
		m := newTestRoot(newMock())
		model, _ := m.Update(StatusMsg{Text: "loaded OK"})
		r := model.(Root)
		if r.statusbar.text != "loaded OK" {
			t.Errorf("statusbar text should be 'loaded OK', got %q", r.statusbar.text)
		}
	})

	t.Run("error StatusMsg sets isError flag", func(t *testing.T) {
		m := newTestRoot(newMock())
		model, _ := m.Update(StatusMsg{Text: "boom", IsError: true})
		r := model.(Root)
		if !r.statusbar.isError {
			t.Error("statusbar.isError should be true for error messages")
		}
	})

	t.Run("empty StatusMsg clears statusbar", func(t *testing.T) {
		m := newTestRoot(newMock())
		m.statusbar = m.statusbar.Set(StatusMsg{Text: "old"})
		model, _ := m.Update(StatusMsg{})
		r := model.(Root)
		if r.statusbar.text != "" {
			t.Errorf("empty StatusMsg should clear statusbar, got %q", r.statusbar.text)
		}
	})

	t.Run("header shows profile and endpoint from handler", func(t *testing.T) {
		mock := newMock()
		mock.endpoint = "https://platform.test"
		mock.profileName = "dev"
		m := newTestRoot(mock)
		model, _ := m.Update(resize)
		r := model.(Root)
		v := r.View()
		if !strings.Contains(v, "dev") {
			t.Error("header should contain profile name")
		}
		if !strings.Contains(v, "https://platform.test") {
			t.Error("header should contain endpoint")
		}
	})

	t.Run("content receives key events when focused", func(t *testing.T) {
		m := newTestRoot(newMock())
		m.focus = focusContent
		m.width, m.height = 120, 40
		// Placeholder handles window resize and just returns self for other keys
		m.content = NewPlaceholder("P")
		model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
		r := model.(Root)
		if _, ok := r.content.(Placeholder); !ok {
			t.Error("content should remain Placeholder after unhandled key")
		}
	})

	t.Run("keybar updates with sidebar bindings when sidebar focused", func(t *testing.T) {
		m := newTestRoot(newMock())
		m.width, m.height = 120, 40
		// tab to sidebar
		model, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
		r := model.(Root)
		v := r.keybar.View()
		if !strings.Contains(v, "↑/↓") || !strings.Contains(v, "navigate") {
			t.Errorf("keybar should show sidebar bindings, got: %q", v)
		}
	})
}
