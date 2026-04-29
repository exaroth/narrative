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
func InitNarrativeModel(ctrl *NarrativeCtrl) *narrativeModel {
	return &narrativeModel{
		ctrl:     ctrl,
		mode:     modeDefault,
		mainView: NewMainViewModel(ctrl.dataCfg.Sources),
	}
}

// Initialize model with given source.
func (m *narrativeModel) initSource(source *Source) {
	m.mainView, _ = m.mainView.Update(UpdateSourceCmd{source: source})
}

// Update actions for main narrative model.
func (m *narrativeModel) mainUpdate(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case PlaybackCmd:
		cmds = append(cmds, m.handlePlayback(int(msg)))

	case UpdateTickMsg:
		if m.ctrl.currentSource != nil && PSM.AllowsBuffering() {
			buf_size := m.ctrl.currentSource.BufferSize()
			if buf_size < m.ctrl.cfg.BufferSize {
				// todo handle errors
				go m.ctrl.currentSource.updateCacheBuffer()
			}
		}
		return UpdateTick()
	case LoadSourceCmd:
		m.ctrl.selectSource(msg.id)

	case SetSourceCmd:
		m.mainView, cmd = m.mainView.Update(UpdateSourceCmd{source: msg.source})
		cmds = append(cmds, cmd)
	}

	m.mainView, cmd = m.mainView.Update(msg)
	cmds = append(cmds, cmd)
	return tea.Batch(cmds...)
}

// Process playback command.
func (m *narrativeModel) handlePlayback(playbackC int) tea.Cmd {
	m.ctrl.remote.HandleCommand(playbackC)
	return WaitForPlayback(PlaybackCh)
}

func (m narrativeModel) Init() tea.Cmd {
	return tea.Batch(
		UpdateTick(),
		WaitForPlayback(PlaybackCh),
	)
}

func (m narrativeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.mode {
	case modeDefault:
		cmd = m.mainUpdate(msg)
	}
	return m, cmd
}

func (m narrativeModel) View() tea.View {
	return m.mainView.View()
	// var v tea.View
	// return v
}
