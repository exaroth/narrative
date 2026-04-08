package debugger

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

type input struct {
	model    textinput.Model
	err      error
	quitting bool
}

func getInput(placeholder string, width int) input {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.SetVirtualCursor(false)
	ti.Focus()
	ti.CharLimit = 156
	ti.SetWidth(width)

	return input{model: ti}
}

func (m input) Init() tea.Cmd {
	return textinput.Blink
}
