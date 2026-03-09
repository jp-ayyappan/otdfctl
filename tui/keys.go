package tui

// KeyBinding describes a key and its action for display in the keybar.
type KeyBinding struct {
	Key  string
	Help string
}

// Keyed is implemented by view models that want to advertise their key bindings
// to the keybar at the bottom of the screen.
type Keyed interface {
	KeyBindings() []KeyBinding
}
