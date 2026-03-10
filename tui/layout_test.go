package tui

import (
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Header
// ---------------------------------------------------------------------------

func TestHeader(t *testing.T) {
	t.Run("renders app name", func(t *testing.T) {
		h := NewHeader("", "")
		h = h.SetWidth(80)
		if !strings.Contains(h.View(), "OpenTDF") {
			t.Error("header should contain 'OpenTDF'")
		}
	})

	t.Run("renders profile name", func(t *testing.T) {
		h := NewHeader("my-profile", "")
		h = h.SetWidth(80)
		if !strings.Contains(h.View(), "my-profile") {
			t.Errorf("header should contain profile name, got: %q", h.View())
		}
	})

	t.Run("renders endpoint", func(t *testing.T) {
		h := NewHeader("", "https://platform.example.com")
		h = h.SetWidth(100)
		if !strings.Contains(h.View(), "https://platform.example.com") {
			t.Errorf("header should contain endpoint, got: %q", h.View())
		}
	})

	t.Run("renders both profile and endpoint", func(t *testing.T) {
		h := NewHeader("prod", "https://kas.example.com")
		h = h.SetWidth(120)
		v := h.View()
		if !strings.Contains(v, "prod") || !strings.Contains(v, "https://kas.example.com") {
			t.Errorf("header missing profile or endpoint: %q", v)
		}
	})

	t.Run("zero width does not panic", func(t *testing.T) {
		h := NewHeader("p", "e")
		h = h.SetWidth(0)
		_ = h.View() // must not panic
	})
}

// ---------------------------------------------------------------------------
// Keybar
// ---------------------------------------------------------------------------

func TestKeybar(t *testing.T) {
	t.Run("renders provided bindings", func(t *testing.T) {
		kb := NewKeybar([]KeyBinding{{Key: "enter", Help: "select"}, {Key: "q", Help: "quit"}})
		kb = kb.SetWidth(80)
		v := kb.View()
		if !strings.Contains(v, "enter") || !strings.Contains(v, "select") {
			t.Errorf("keybar missing 'enter'/'select': %q", v)
		}
		if !strings.Contains(v, "q") || !strings.Contains(v, "quit") {
			t.Errorf("keybar missing 'q'/'quit': %q", v)
		}
	})

	t.Run("empty bindings renders without panic", func(t *testing.T) {
		kb := NewKeybar(nil).SetWidth(80)
		_ = kb.View()
	})

	t.Run("SetBindings replaces bindings", func(t *testing.T) {
		kb := NewKeybar([]KeyBinding{{Key: "old", Help: "old-help"}})
		kb = kb.SetBindings([]KeyBinding{{Key: "new", Help: "new-help"}})
		kb = kb.SetWidth(80)
		v := kb.View()
		if strings.Contains(v, "old") {
			t.Error("old binding should have been replaced")
		}
		if !strings.Contains(v, "new") {
			t.Error("new binding should be present")
		}
	})
}

// ---------------------------------------------------------------------------
// StatusBar
// ---------------------------------------------------------------------------

func TestStatusBar(t *testing.T) {
	t.Run("empty renders blank line without panic", func(t *testing.T) {
		sb := NewStatusBar().SetWidth(80)
		_ = sb.View()
	})

	t.Run("shows text on set", func(t *testing.T) {
		sb := NewStatusBar().SetWidth(80)
		sb = sb.Set(StatusMsg{Text: "hello world"})
		if !strings.Contains(sb.View(), "hello world") {
			t.Errorf("status bar should show message: %q", sb.View())
		}
	})

	t.Run("error message is stored", func(t *testing.T) {
		sb := NewStatusBar().SetWidth(80)
		sb = sb.Set(StatusMsg{Text: "something failed", IsError: true})
		if !sb.isError {
			t.Error("isError should be true")
		}
		if !strings.Contains(sb.View(), "something failed") {
			t.Error("error text should be visible")
		}
	})

	t.Run("clears on empty StatusMsg", func(t *testing.T) {
		sb := NewStatusBar().SetWidth(80)
		sb = sb.Set(StatusMsg{Text: "temp"})
		sb = sb.Set(StatusMsg{})
		if strings.Contains(sb.View(), "temp") {
			t.Error("status bar should be cleared")
		}
	})

	t.Run("clearStatusCmd fires after duration", func(t *testing.T) {
		cmd := clearStatusCmd(1 * time.Millisecond)
		msg := cmd()
		if _, ok := msg.(StatusMsg); !ok {
			t.Errorf("clearStatusCmd should return StatusMsg, got %T", msg)
		}
	})
}
