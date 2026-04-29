package narrative

import (
	"strconv"
	"sync"

	tea "charm.land/bubbletea/v2"
	"github.com/sirupsen/logrus"
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

type playbackStatus int

const (
	playbackPlaying playbackStatus = iota
	playbackPaused
	playbackBuffering
	playbackIdle
)

func (p playbackStatus) String() string {
	switch p {
	case playbackPlaying:
		return "playing"
	case playbackPaused:
		return "paused"
	case playbackBuffering:
		return "buffering"
	case playbackIdle:
		return "idle"
	}
	return "unknown"
}

type playbackSM struct {
	status playbackStatus
	mut    sync.Mutex
}

func (p *playbackSM) SetStatus(status playbackStatus) {
	p.mut.Lock()
	defer p.mut.Unlock()
	p.status = status
}

func (p *playbackSM) Status() playbackStatus {
	p.mut.Lock()
	defer p.mut.Unlock()
	return p.status
}
func (p *playbackSM) AllowsBuffering() bool {
	return p.status != playbackBuffering && p.status != playbackIdle
}

var PSM = playbackSM{status: playbackIdle}

// channel controlling playback.
var playbackCh = make(chan int)

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

	case StartPlaybackCmd:
		PSM.SetStatus(playbackPlaying)
		m.ctrl.play(func() {
			playbackCh <- 1
		})

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
	switch playbackC {
	// continuous playback
	case 1:
		if PSM.Status() == playbackPaused {
			logrus.Info("handlePlayback: resume")
			PSM.SetStatus(playbackPlaying)
			m.ctrl.resume()
		} else {
			logrus.Info("handlePlayback: play")
			sn := m.ctrl.currentSource.incrementSentenceNum()
			// at the end of playback.
			if sn == -1 {
				m.ctrl.stop()
				break
			}
			PSM.SetStatus(playbackPlaying)
			m.ctrl.play(func() {
				playbackCh <- 1
			})
		}
	// pause
	case 0:
		logrus.Info("handlePlayback: pause")
		PSM.SetStatus(playbackPaused)
		m.ctrl.pause()
	// stop
	case -1:
		logrus.Info("handlePlayback: stop")
		PSM.SetStatus(playbackIdle)
		// todo
		m.ctrl.stop()
	default:
		panic("Invalid playback cmd passed: " + strconv.Itoa(playbackC))
	}
	return WaitForPlayback(playbackCh)
}

func (m narrativeModel) Init() tea.Cmd {
	return tea.Batch(
		UpdateTick(),
		WaitForPlayback(playbackCh),
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
