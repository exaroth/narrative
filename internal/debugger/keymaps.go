package debugger

import (
	"github.com/exaroth/narrative/bubbles/key"
	"github.com/exaroth/narrative/bubbles/viewport"
)

func GetSentenceListKeymap() viewport.KeyMap {
	return viewport.KeyMap{
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", "f"),
			key.WithHelp("f/pgdn", "page down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "b"),
			key.WithHelp("b/pgup", "page up"),
		),
		HalfPageUp: key.NewBinding(
			key.WithKeys("u", "ctrl+u"),
			key.WithHelp("u", "½ page up"),
		),
		HalfPageDown: key.NewBinding(
			key.WithKeys("d", "ctrl+d"),
			key.WithHelp("d", "½ page down"),
		),
		Up: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("up", "Up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("down", "Down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left"),
			key.WithHelp("left", "move left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right"),
			key.WithHelp("right", "move right"),
		),
	}
}
