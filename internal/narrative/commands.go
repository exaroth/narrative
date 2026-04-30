package narrative

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

// Simple wrapper for returning tea commands.
func teaCmd(payload interface{}) tea.Cmd {
	return func() tea.Msg {
		return payload
	}
}

// Command used for toggling source with given id.
// This command will switch current source and start
// automatic playback.
type SelectSourceCmd struct {
	id string
}

// Toggle source with given id
func SelectSource(id string) tea.Cmd {
	return teaCmd(SelectSourceCmd{
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

// // Command for starting playback for currently selected
// // source. If already playing will do nothing.
// type PlaySourceCmd struct{}

// func PlaySource() tea.Cmd {
// 	return teaCmd(PlaySourceCmd{})
// }

// // Pause source playback (unless paused).
// type PauseSourceCmd struct{}

// func PauseSource() tea.Cmd {
// 	return teaCmd(PauseSourceCmd{})
// }

// Update current source
type UpdateSourceCmd struct {
	source *Source
}

// Command for setting current source in the model.
type SetSourceCmd struct {
	source *Source
}

type UpdateTickMsg time.Time

func UpdateTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return UpdateTickMsg(t)
	})
}

// Command to start playback
type StartPlaybackCmd struct{}

// Enable disable playback
type TogglePlaybackCmd struct{}

func TogglePlayback() tea.Cmd {
	return teaCmd(TogglePlaybackCmd{})
}

type PlaybackCmd int

// Wait for playback command to be passed
// Commands are integer based, available values:
// 0 - pause
// 1 - play
// -1 - stop
// 2 - start new playback
func WaitForPlayback(sub chan int) tea.Cmd {
	return func() tea.Msg {
		return PlaybackCmd(<-sub)
	}
}

type ErrorCmd error

func WaitForError(sub chan error) tea.Cmd {
	return func() tea.Msg {
		return ErrorCmd(<-sub)
	}
}
