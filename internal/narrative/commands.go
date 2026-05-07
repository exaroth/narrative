package narrative

import (
	"time"

	"charm.land/bubbles/v2/list"
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
// automatic playback. Set idle to true to disable
// autoplay.
type SelectSourceCmd struct {
	id       string
	index    int
	autoplay bool
}

// Toggle source with given id, also pass index to move the
// cursor to selected source position (set it to -1 to ignore).
func SelectSource(id string, index int) tea.Cmd {
	return teaCmd(SelectSourceCmd{
		id:       id,
		index:    index,
		autoplay: true,
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

// Update current source
type UpdateSourceCmd struct {
	source *Source
}

// Command for setting current source in the model.
type SetSourceCmd struct {
	source *Source
}

// Delete source with given id.
type DeleteSourceCmd struct {
	id string
}

func DeleteSource(id string) tea.Cmd {
	return teaCmd(DeleteSourceCmd{
		id: id,
	})
}

type UpdateSourceListCmd struct {
	items []list.Item
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

// Stores errors send in the app
// to be shown to the user.
type ErrorCmd error

// ErrorCmd receiver func.
func WaitForError(sub chan error) tea.Cmd {
	return func() tea.Msg {
		return ErrorCmd(<-sub)
	}
}

// Stores messages to be be sent and shown
// to the user.
type MessageCmd string

// MessageCmd receiver func.
func WaitForMessage(sub chan string) tea.Cmd {
	return func() tea.Msg {
		return MessageCmd(<-sub)
	}
}

// Stores sentence number we ought to rewind
// text source to.
type FastForwardCmd int

func WaitForFastForward(sub chan int) tea.Cmd {
	return func() tea.Msg {
		return FastForwardCmd(<-sub)
	}
}

// Remove any modal currently being displayed
type RemoveModalCmd struct{}

// Command for closing spinner loader.
type SpinnerCloseCmd struct{}

// Close currently running spinner
func WaitForSpinnerClose(sub chan struct{}) tea.Cmd {
	return func() tea.Msg {
		return SpinnerCloseCmd(<-sub)
	}
}

// Command used for updating running spinners message.
type SpinnerMsgCmd string

func WaitForSpinnerMessage(sub chan string) tea.Cmd {
	return func() tea.Msg {
		return SpinnerMsgCmd(<-sub)
	}
}

// Command for displaying prompt, pass
// funcs for yes and no, you can pass nil
// in which case no action will be taken.
// Also pass text to be displayed alongside
// the prompt.
type ShowPromptCmd struct {
	text    string
	okFunc  tea.Cmd
	nayFunc tea.Cmd
}

func ShowPrompt(text string, okFunc, nayFunc tea.Cmd) tea.Cmd {
	return teaCmd(ShowPromptCmd{
		text:    text,
		okFunc:  okFunc,
		nayFunc: nayFunc,
	})
}

// Command for closing currently displayed prompt
type ClosePromptCmd struct{}

func ClosePrompt() tea.Cmd {
	return teaCmd(ClosePromptCmd{})
}

// Message containing download progress.
type DownloadProgressMsg float64

// Message containing error that occurred during downloading.
type DownloadProgressErr error

func WaitForDownloadProgressMsg(sub chan float64) tea.Cmd {
	return func() tea.Msg {
		return DownloadProgressMsg(<-sub)
	}
}

func WaitForDownloadProgressErr(sub chan error) tea.Cmd {
	return func() tea.Msg {
		return DownloadProgressErr(<-sub)
	}
}
