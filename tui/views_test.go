package tui

import (
	"context"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/opentdf/platform/protocol/go/common"
	"github.com/opentdf/platform/protocol/go/policy"
)

var ctx = context.Background()

// windowMsg is a standard resize event used to size views in tests.
var windowMsg = tea.WindowSizeMsg{Width: 100, Height: 30}

// sendKey is a convenience helper.
func sendKey(m tea.Model, key string) (tea.Model, tea.Cmd) {
	var msg tea.KeyMsg
	switch key {
	case "enter":
		msg = tea.KeyMsg{Type: tea.KeyEnter}
	case "backspace":
		msg = tea.KeyMsg{Type: tea.KeyBackspace}
	case "q":
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}
	case "ctrl+c":
		msg = tea.KeyMsg{Type: tea.KeyCtrlC}
	default:
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
	}
	return m.Update(msg)
}

// containsStatus checks if any command in the chain produces a StatusMsg with IsError.
func cmdProducesErrorStatus(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}
	msg := cmd()
	sm, ok := msg.(StatusMsg)
	return ok && sm.IsError
}

// ---------------------------------------------------------------------------
// Placeholder
// ---------------------------------------------------------------------------

func TestPlaceholder(t *testing.T) {
	t.Run("renders coming soon text", func(t *testing.T) {
		p := NewPlaceholder("Foo Resource")
		p.width, p.height = 80, 20
		v := p.View()
		if !strings.Contains(v, "Foo Resource") {
			t.Errorf("placeholder should contain resource name, got: %q", v)
		}
		if !strings.Contains(v, "coming soon") {
			t.Errorf("placeholder should say 'coming soon', got: %q", v)
		}
	})

	t.Run("window resize updates size without panic", func(t *testing.T) {
		p := NewPlaceholder("X")
		model, _ := p.Update(windowMsg)
		ph := model.(Placeholder)
		if ph.width != 100 || ph.height != 30 {
			t.Errorf("expected 100x30, got %dx%d", ph.width, ph.height)
		}
	})

	t.Run("KeyBindings returns nil (no keys)", func(t *testing.T) {
		if NewPlaceholder("X").KeyBindings() != nil {
			t.Error("placeholder should have no key bindings")
		}
	})
}

// ---------------------------------------------------------------------------
// AttributeList
// ---------------------------------------------------------------------------

func loadedAttributeList(h TUIHandler, items []list.Item) AttributeList {
	l := list.New(items, list.NewDefaultDelegate(), 100, 28)
	l.Title = "Attributes"
	return AttributeList{h: h, list: l, loading: false}
}

func TestAttributeList(t *testing.T) {
	t.Run("loading state renders spinner text", func(t *testing.T) {
		m, _ := InitAttributeList(ctx, "", newMock())
		al := m.(AttributeList)
		if !al.loading {
			t.Error("list should start in loading state")
		}
		v := al.View()
		if !strings.Contains(v, "Loading") {
			t.Errorf("loading view should mention loading, got: %q", v)
		}
	})

	t.Run("attributesLoadedMsg with items transitions out of loading", func(t *testing.T) {
		m, _ := InitAttributeList(ctx, "", newMock())
		items := []list.Item{AttributeItem{id: "1", name: "alpha"}, AttributeItem{id: "2", name: "beta"}}
		model, _ := m.Update(attributesLoadedMsg{items: items})
		al := model.(AttributeList)
		if al.loading {
			t.Error("should not be loading after receiving items")
		}
		if al.list.Items() == nil || len(al.list.Items()) != 2 {
			t.Errorf("expected 2 items, got %d", len(al.list.Items()))
		}
	})

	t.Run("attributesLoadedMsg with error sends StatusMsg", func(t *testing.T) {
		m, _ := InitAttributeList(ctx, "", newMock())
		model, cmd := m.Update(attributesLoadedMsg{err: errTest})
		al := model.(AttributeList)
		if al.loading {
			t.Error("loading should be false even on error")
		}
		if !cmdProducesErrorStatus(cmd) {
			t.Error("error load should produce an error StatusMsg")
		}
	})

	t.Run("enter on empty list is a no-op", func(t *testing.T) {
		m := loadedAttributeList(newMock(), nil)
		model, _ := sendKey(m, "enter")
		if _, ok := model.(AttributeList); !ok {
			t.Error("enter on empty list should remain AttributeList")
		}
	})

	t.Run("enter navigates to AttributeView", func(t *testing.T) {
		mock := newMock()
		mock.attributes = []*policy.Attribute{testAttr("attr-1", "classification")}
		items := []list.Item{AttributeItem{id: "attr-1", name: "classification"}}
		m := loadedAttributeList(mock, items)
		model, _ := sendKey(m, "enter")
		if _, ok := model.(AttributeView); !ok {
			t.Errorf("enter should navigate to AttributeView, got %T", model)
		}
	})

	t.Run("q key quits", func(t *testing.T) {
		m := loadedAttributeList(newMock(), nil)
		_, cmd := sendKey(m, "q")
		if cmd == nil {
			t.Fatal("q should return a command")
		}
		if _, ok := cmd().(tea.QuitMsg); !ok {
			t.Error("q should produce QuitMsg")
		}
	})

	t.Run("window resize updates list size", func(t *testing.T) {
		m := loadedAttributeList(newMock(), nil)
		model, _ := m.Update(windowMsg)
		al := model.(AttributeList)
		if al.list.Width() != 100 {
			t.Errorf("list width should be 100, got %d", al.list.Width())
		}
	})

	t.Run("KeyBindings returns non-empty slice", func(t *testing.T) {
		m := loadedAttributeList(newMock(), nil)
		if len(m.KeyBindings()) == 0 {
			t.Error("AttributeList should have key bindings")
		}
	})
}

// ---------------------------------------------------------------------------
// AttributeView
// ---------------------------------------------------------------------------

func TestAttributeView(t *testing.T) {
	mock := newMock()
	mock.attributes = []*policy.Attribute{testAttr("attr-1", "classification")}

	t.Run("renders attribute fields", func(t *testing.T) {
		m, _ := InitAttributeView(ctx, "attr-1", mock)
		v := m.View()
		if !strings.Contains(v, "classification") {
			t.Errorf("view should contain attribute name, got: %q", v)
		}
	})

	t.Run("q key quits", func(t *testing.T) {
		m, _ := InitAttributeView(ctx, "attr-1", mock)
		_, cmd := sendKey(m, "q")
		if cmd == nil {
			t.Fatal("q should return a command")
		}
		if _, ok := cmd().(tea.QuitMsg); !ok {
			t.Error("q should produce QuitMsg")
		}
	})

	t.Run("backspace navigates to AttributeList", func(t *testing.T) {
		m, _ := InitAttributeView(ctx, "attr-1", mock)
		model, _ := sendKey(m, "backspace")
		if _, ok := model.(AttributeList); !ok {
			t.Errorf("backspace should navigate to AttributeList, got %T", model)
		}
	})

	t.Run("KeyBindings returns non-empty slice", func(t *testing.T) {
		m, _ := InitAttributeView(ctx, "attr-1", mock)
		if len(m.KeyBindings()) == 0 {
			t.Error("AttributeView should have key bindings")
		}
	})
}

// ---------------------------------------------------------------------------
// LabelList
// ---------------------------------------------------------------------------

func TestLabelList(t *testing.T) {
	attr := testAttr("attr-1", "classification")
	attr.Metadata = &common.Metadata{Labels: map[string]string{"env": "prod", "tier": "secret"}}

	t.Run("renders all labels", func(t *testing.T) {
		m, _ := InitLabelList(attr, newMock())
		v := m.View()
		if !strings.Contains(v, "env") {
			t.Errorf("label list should show 'env', got: %q", v)
		}
	})

	t.Run("backspace navigates to AttributeView", func(t *testing.T) {
		m, _ := InitLabelList(attr, newMock())
		model, _ := sendKey(m, "backspace")
		if _, ok := model.(AttributeView); !ok {
			t.Errorf("backspace should go to AttributeView, got %T", model)
		}
	})

	t.Run("c key opens empty LabelUpdate", func(t *testing.T) {
		m, _ := InitLabelList(attr, newMock())
		model, _ := sendKey(m, "c")
		if _, ok := model.(LabelUpdate); !ok {
			t.Errorf("c should open LabelUpdate, got %T", model)
		}
	})

	t.Run("KeyBindings returns non-empty slice", func(t *testing.T) {
		m, _ := InitLabelList(attr, newMock())
		ll := m.(LabelList)
		if len(ll.KeyBindings()) == 0 {
			t.Error("LabelList should have key bindings")
		}
	})
}

// ---------------------------------------------------------------------------
// NamespaceList
// ---------------------------------------------------------------------------

func loadedNamespaceList(h TUIHandler, items []list.Item) NamespaceList {
	l := list.New(items, list.NewDefaultDelegate(), 100, 28)
	l.Title = "Namespaces"
	return NamespaceList{h: h, list: l, loading: false}
}

func TestNamespaceList(t *testing.T) {
	t.Run("starts in loading state", func(t *testing.T) {
		m, _ := InitNamespaceList(ctx, newMock())
		nl := m.(NamespaceList)
		if !nl.loading {
			t.Error("namespace list should start loading")
		}
	})

	t.Run("namespacesLoadedMsg populates items", func(t *testing.T) {
		m, _ := InitNamespaceList(ctx, newMock())
		items := []list.Item{NamespaceItem{id: "ns-1", name: "opentdf.io"}}
		model, _ := m.Update(namespacesLoadedMsg{items: items})
		nl := model.(NamespaceList)
		if nl.loading || len(nl.list.Items()) != 1 {
			t.Errorf("expected 1 item, loading=false; got loading=%v, items=%d", nl.loading, len(nl.list.Items()))
		}
	})

	t.Run("namespacesLoadedMsg with error sends StatusMsg", func(t *testing.T) {
		m, _ := InitNamespaceList(ctx, newMock())
		_, cmd := m.Update(namespacesLoadedMsg{err: errTest})
		if !cmdProducesErrorStatus(cmd) {
			t.Error("should emit error StatusMsg on load failure")
		}
	})

	t.Run("enter navigates to NamespaceView", func(t *testing.T) {
		items := []list.Item{NamespaceItem{id: "ns-1", name: "opentdf.io"}}
		m := loadedNamespaceList(newMock(), items)
		model, _ := sendKey(m, "enter")
		if _, ok := model.(NamespaceView); !ok {
			t.Errorf("enter should go to NamespaceView, got %T", model)
		}
	})

	t.Run("enter on empty list is a no-op", func(t *testing.T) {
		m := loadedNamespaceList(newMock(), nil)
		model, _ := sendKey(m, "enter")
		if _, ok := model.(NamespaceList); !ok {
			t.Error("enter on empty list should remain NamespaceList")
		}
	})
}

// ---------------------------------------------------------------------------
// NamespaceView
// ---------------------------------------------------------------------------

func TestNamespaceView(t *testing.T) {
	mock := newMock()
	mock.namespaces = []*policy.Namespace{testNamespace("ns-1", "opentdf.io")}

	t.Run("renders namespace fields", func(t *testing.T) {
		m, _ := InitNamespaceView(ctx, "ns-1", mock)
		v := m.View()
		if !strings.Contains(v, "opentdf.io") {
			t.Errorf("namespace view should show name, got: %q", v)
		}
	})

	t.Run("backspace navigates to NamespaceList", func(t *testing.T) {
		m, _ := InitNamespaceView(ctx, "ns-1", mock)
		model, _ := sendKey(m, "backspace")
		if _, ok := model.(NamespaceList); !ok {
			t.Errorf("backspace should go to NamespaceList, got %T", model)
		}
	})
}

// ---------------------------------------------------------------------------
// SubjectMappingList
// ---------------------------------------------------------------------------

func loadedSMList(h TUIHandler, items []list.Item) SubjectMappingList {
	l := list.New(items, list.NewDefaultDelegate(), 100, 28)
	l.Title = "Subject Mappings"
	return SubjectMappingList{h: h, list: l, loading: false}
}

func TestSubjectMappingList(t *testing.T) {
	t.Run("starts in loading state", func(t *testing.T) {
		m, _ := InitSubjectMappingList(ctx, newMock())
		sl := m.(SubjectMappingList)
		if !sl.loading {
			t.Error("should start loading")
		}
	})

	t.Run("loaded items populated", func(t *testing.T) {
		m, _ := InitSubjectMappingList(ctx, newMock())
		items := []list.Item{SubjectMappingItem{id: "sm-1", attrVal: "https://ns.io/attr/a/value/b", actionCt: 2}}
		model, _ := m.Update(subjectMappingsLoadedMsg{items: items})
		sl := model.(SubjectMappingList)
		if sl.loading || len(sl.list.Items()) != 1 {
			t.Error("should have 1 item and not be loading")
		}
	})

	t.Run("error load emits StatusMsg", func(t *testing.T) {
		m, _ := InitSubjectMappingList(ctx, newMock())
		_, cmd := m.Update(subjectMappingsLoadedMsg{err: errTest})
		if !cmdProducesErrorStatus(cmd) {
			t.Error("should emit error StatusMsg")
		}
	})

	t.Run("enter navigates to SubjectMappingView", func(t *testing.T) {
		items := []list.Item{SubjectMappingItem{id: "sm-1", attrVal: "fqn", actionCt: 1}}
		m := loadedSMList(newMock(), items)
		model, _ := sendKey(m, "enter")
		if _, ok := model.(SubjectMappingView); !ok {
			t.Errorf("enter should go to SubjectMappingView, got %T", model)
		}
	})
}

// ---------------------------------------------------------------------------
// SubjectMappingView
// ---------------------------------------------------------------------------

func TestSubjectMappingView(t *testing.T) {
	mock := newMock()
	mock.subjectMappings = []*policy.SubjectMapping{testSubjectMapping("sm-1", "https://ns.io/attr/a/value/b")}

	t.Run("renders mapping fields", func(t *testing.T) {
		m, _ := InitSubjectMappingView(ctx, "sm-1", mock)
		v := m.View()
		if !strings.Contains(v, "sm-1") {
			t.Errorf("view should contain mapping ID, got: %q", v)
		}
	})

	t.Run("backspace navigates to SubjectMappingList", func(t *testing.T) {
		m, _ := InitSubjectMappingView(ctx, "sm-1", mock)
		model, _ := sendKey(m, "backspace")
		if _, ok := model.(SubjectMappingList); !ok {
			t.Errorf("backspace should go to SubjectMappingList, got %T", model)
		}
	})
}

// ---------------------------------------------------------------------------
// KASRegistryList
// ---------------------------------------------------------------------------

func loadedKASList(h TUIHandler, items []list.Item) KASRegistryList {
	l := list.New(items, list.NewDefaultDelegate(), 100, 28)
	l.Title = "KAS Registry"
	return KASRegistryList{h: h, list: l, loading: false}
}

func TestKASRegistryList(t *testing.T) {
	t.Run("starts in loading state", func(t *testing.T) {
		m, _ := InitKASRegistryList(ctx, newMock())
		kl := m.(KASRegistryList)
		if !kl.loading {
			t.Error("should start loading")
		}
	})

	t.Run("loaded items populated", func(t *testing.T) {
		m, _ := InitKASRegistryList(ctx, newMock())
		items := []list.Item{KASRegistryItem{id: "kas-1", name: "primary", uri: "https://kas.example.com"}}
		model, _ := m.Update(kasRegistryLoadedMsg{items: items})
		kl := model.(KASRegistryList)
		if kl.loading || len(kl.list.Items()) != 1 {
			t.Error("should have 1 item and not be loading")
		}
	})

	t.Run("error load emits StatusMsg", func(t *testing.T) {
		m, _ := InitKASRegistryList(ctx, newMock())
		_, cmd := m.Update(kasRegistryLoadedMsg{err: errTest})
		if !cmdProducesErrorStatus(cmd) {
			t.Error("should emit error StatusMsg")
		}
	})

	t.Run("KASRegistryItem title falls back to URI when name is empty", func(t *testing.T) {
		item := KASRegistryItem{id: "k", name: "", uri: "https://kas.io"}
		if item.Title() != "https://kas.io" {
			t.Errorf("should use URI as title when name empty, got %q", item.Title())
		}
	})

	t.Run("enter navigates to KASRegistryView", func(t *testing.T) {
		items := []list.Item{KASRegistryItem{id: "kas-1", name: "primary", uri: "https://kas.example.com"}}
		m := loadedKASList(newMock(), items)
		model, _ := sendKey(m, "enter")
		if _, ok := model.(KASRegistryView); !ok {
			t.Errorf("enter should go to KASRegistryView, got %T", model)
		}
	})
}

// ---------------------------------------------------------------------------
// KASRegistryView
// ---------------------------------------------------------------------------

func TestKASRegistryView(t *testing.T) {
	mock := newMock()
	mock.kasEntries = []*policy.KeyAccessServer{testKAS("kas-1", "primary", "https://kas.example.com")}

	t.Run("renders KAS fields", func(t *testing.T) {
		m, _ := InitKASRegistryView(ctx, "kas-1", mock)
		v := m.View()
		if !strings.Contains(v, "kas-1") {
			t.Errorf("view should contain KAS ID, got: %q", v)
		}
	})

	t.Run("backspace navigates to KASRegistryList", func(t *testing.T) {
		m, _ := InitKASRegistryView(ctx, "kas-1", mock)
		model, _ := sendKey(m, "backspace")
		if _, ok := model.(KASRegistryList); !ok {
			t.Errorf("backspace should go to KASRegistryList, got %T", model)
		}
	})
}

// ---------------------------------------------------------------------------
// ResourceMappingList
// ---------------------------------------------------------------------------

func loadedRMList(h TUIHandler, items []list.Item) ResourceMappingList {
	l := list.New(items, list.NewDefaultDelegate(), 100, 28)
	l.Title = "Resource Mappings"
	return ResourceMappingList{h: h, list: l, loading: false}
}

func TestResourceMappingList(t *testing.T) {
	t.Run("starts in loading state", func(t *testing.T) {
		m, _ := InitResourceMappingList(ctx, newMock())
		rl := m.(ResourceMappingList)
		if !rl.loading {
			t.Error("should start loading")
		}
	})

	t.Run("loaded items populated", func(t *testing.T) {
		m, _ := InitResourceMappingList(ctx, newMock())
		items := []list.Item{ResourceMappingItem{id: "rm-1", terms: "SECRET, CONFIDENTIAL"}}
		model, _ := m.Update(resourceMappingsLoadedMsg{items: items})
		rl := model.(ResourceMappingList)
		if rl.loading || len(rl.list.Items()) != 1 {
			t.Error("should have 1 item and not be loading")
		}
	})

	t.Run("error load emits StatusMsg", func(t *testing.T) {
		m, _ := InitResourceMappingList(ctx, newMock())
		_, cmd := m.Update(resourceMappingsLoadedMsg{err: errTest})
		if !cmdProducesErrorStatus(cmd) {
			t.Error("should emit error StatusMsg")
		}
	})

	t.Run("enter navigates to ResourceMappingView", func(t *testing.T) {
		items := []list.Item{ResourceMappingItem{id: "rm-1", terms: "SECRET"}}
		m := loadedRMList(newMock(), items)
		model, _ := sendKey(m, "enter")
		if _, ok := model.(ResourceMappingView); !ok {
			t.Errorf("enter should go to ResourceMappingView, got %T", model)
		}
	})
}

// ---------------------------------------------------------------------------
// ResourceMappingView
// ---------------------------------------------------------------------------

func TestResourceMappingView(t *testing.T) {
	mock := newMock()
	mock.resourceMaps = []*policy.ResourceMapping{testResourceMapping("rm-1", []string{"SECRET", "CONFIDENTIAL"})}

	t.Run("renders mapping fields", func(t *testing.T) {
		m, _ := InitResourceMappingView(ctx, "rm-1", mock)
		v := m.View()
		if !strings.Contains(v, "rm-1") {
			t.Errorf("view should contain mapping ID, got: %q", v)
		}
	})

	t.Run("backspace navigates to ResourceMappingList", func(t *testing.T) {
		m, _ := InitResourceMappingView(ctx, "rm-1", mock)
		model, _ := sendKey(m, "backspace")
		if _, ok := model.(ResourceMappingList); !ok {
			t.Errorf("backspace should go to ResourceMappingList, got %T", model)
		}
	})
}
