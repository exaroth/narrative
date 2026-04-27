package narrative

import tea "charm.land/bubbletea/v2"

// Simple wrapper for returning tea commands.
func teaCmd(payload interface{}) tea.Cmd {
	return func() tea.Msg {
		return payload
	}
}

// Command used for toggling source with given id.
// This command will switch current source and start
// automatic playback.
type ToggleSourceCmd struct {
	id string
}

// Toggle source with given id
func ToggleSource(id string) tea.Cmd {
	return teaCmd(ToggleSourceCmd{
		id: id,
	})
}

// Command for loading source with given id.
type LoadSourceCmd struct {
	id string
}

func LoadSource(id string) tea.Cmd {
	return teaCmd(LoadSourceCmd{
		id: id,
	})
}

// Command for starting playback for currently selected
// source. If already playing will do nothing.
type PlaySourceCmd struct{}

func PlaySource() tea.Cmd {
	return teaCmd(PlaySourceCmd{})
}

// Pause source playback (unless paused).
type PauseSourceCmd struct{}

func PauseSource() tea.Cmd {
	return teaCmd(PauseSourceCmd{})
}

// Update current source
type UpdateSourceCmd struct {
	source *Source
}

// Command for setting current source in the model.
type SetSourceCmd struct {
	source *Source
}
