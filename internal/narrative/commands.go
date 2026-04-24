package narrative

import tea "charm.land/bubbletea/v2"

// Command used for toggling source with given id.
// This command will switch current source and start
// automatic playback.
type ToggleSourceCmd struct {
	id string
}

// Toggle source with given id
func ToggleSource(id string) tea.Cmd {
	return func() tea.Msg {
		return ToggleSourceCmd{
			id: id,
		}
	}
}

// Command for starting playback for currently selected
// source. If already playing will do nothing.
type PlaySourceCmd struct{}

func PlaySource(id string) tea.Cmd {
	return func() tea.Msg {
		return PlaySourceCmd{}
	}
}

// Pause source playback (unless paused).
type PauseSourceCmd struct{}

func PauseSource(id string) tea.Cmd {
	return func() tea.Msg {
		return PauseSourceCmd{}
	}
}

// Update current source
type UpdateSourceCmd struct {
	source *Source
}
