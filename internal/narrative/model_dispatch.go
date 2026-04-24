package narrative

import (
	tea "charm.land/bubbletea/v2"
)

type mode int

const (
	// Default display mode (sources + transcript).
	modeDefault mode = iota
	// Wizard mode for downloading models/libs during init.
	modeDownloader
	// Mode used for converting sources to mp3.
	modeConverter
	modeHelp
)

// Main model used by narrative, used primarily for dispatching.
type narrativeModel struct {
	mode     mode
	ctrl     *NarrativeCtrl
	mainView tea.Model
}

// Initialize new narrative model.
func NewModel(ctrl *NarrativeCtrl) *narrativeModel {
	return &narrativeModel{
		ctrl:     ctrl,
		mode:     modeDefault,
		mainView: NewMainViewModel(ctrl.dataCfg.Sources),
	}
}

func (m *narrativeModel) SetSource(source *Source) {
	m.mainView, _ = m.mainView.Update(UpdateSourceCmd{source: source})
}

func (m narrativeModel) Init() tea.Cmd {
	return nil
}

func (m narrativeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.mode {
	case modeDefault:
		m.mainView, cmd = m.mainView.Update(msg)
	}

	return m, cmd
}

func (m narrativeModel) View() tea.View {
	return m.mainView.View()
	// var v tea.View
	// return v
}
