package narrative

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// How long we show modals for (in seconds).
const defaultModalTimeout = 4

var (
	// Channel holding errors for the app.
	ErrorCh = make(chan error)
	// Channel for sending messages to be shown
	// to the user
	MessageCh = make(chan string)
)

// Defines mode app is running.
type mode int

const (
	// Default display mode (sources + transcript).
	modeDefault mode = iota
	// Wizard mode for downloading models/libs during init.
	modeDownloader
	// Mode used for converting sources to mp3.
	modeConverter
	// Show detailed help.
	modeHelp
)

// Defines available types of the modal popup.
type modalType int

const (
	modalInfo modalType = iota
	modalError
)

// Contains information stored in the modal popup.
type Modal struct {
	t       modalType
	content string
}

// Main model used by narrative, used primarily for dispatching.
type narrativeModel struct {
	mode         mode
	ctrl         *NarrativeCtrl
	mainView     tea.Model
	modal        *Modal
	modalTimeout int
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
		if m.modal != nil {
			m.modalTimeout -= 1
			if m.modalTimeout < 0 {
				m.modal = nil
			}
		}
		cmds = append(cmds, UpdateTick())
	case ErrorCmd:
		m.addModal(modalError, error(msg).Error())
		cmds = append(cmds, WaitForError(ErrorCh))
	case MessageCmd:
		m.addModal(modalInfo, string(msg))
		cmds = append(cmds, WaitForMessage(MessageCh))
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

func (m *narrativeModel) addModal(t modalType, content string) {
	m.modal = &Modal{
		t:       t,
		content: content,
	}
	m.modalTimeout = defaultModalTimeout
}

func (m narrativeModel) Init() tea.Cmd {
	return tea.Batch(
		UpdateTick(),
		WaitForPlayback(PlaybackCh),
		WaitForError(ErrorCh),
		WaitForMessage(MessageCh),
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
	var v tea.View
	main_v := m.mainView.View()
	layers := []*lipgloss.Layer{
		lipgloss.NewLayer(main_v.Content),
		// todo add modal
	}
	comp := lipgloss.NewCompositor(layers...)
	v.SetContent(comp.Render())
	v.AltScreen = true
	return v
}
