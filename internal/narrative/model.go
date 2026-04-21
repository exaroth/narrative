package narrative

import (
	tea "charm.land/bubbletea/v2"
)

type mode int

const (
	// Default display mode (sources + transcript).
	modeDefault mode = iota
	// Default mode without transcript.
	modeDefaultMin
	// Wizard mode for downloading models/libs during init.
	modeDownloader
	// Mode used for converting sources to mp3.
	modeConverter
	modeHelp
)

// Main model used by narrative, used primarily for dispatching.
type narrativeModel struct {
	mode mode
	ctrl *NarrativeCtrl
}

// Initialize new narrative model.
func NewModel(ctrl *NarrativeCtrl) *narrativeModel {
	return &narrativeModel{
		ctrl: ctrl,
		mode: modeDefault,
	}
}

func (m narrativeModel) Init() tea.Cmd {
	return nil
}

func (m narrativeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m narrativeModel) View() tea.View {
	var v tea.View
	return v
}
