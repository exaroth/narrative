package narrative

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/exaroth/narrative/internal/common"
)

// How long we show modals for (in seconds).
const (
	defaultModalTimeout = 4
	defaultModalXOffset = 2
	defaultModalYOffset = 1
	defaultModalWidth   = 30
	windowTitle         = "Narrative"
)

var (
	// Channel holding errors for the app.
	ErrorCh = make(chan error)
	// Channel for sending messages to be shown
	// to the user
	MessageCh = make(chan string)
	// Channel for fast forwarding or rewinding text
	// sources.
	FastForwardCh = make(chan int)
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
	mode          mode
	ctrl          *NarrativeCtrl
	mainView      tea.Model
	modal         *Modal
	showInfo      bool
	modalTimeout  int
	width, height int
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
			if m.modalTimeout < 0 {
				m.modal = nil
			}
			m.modalTimeout -= 1
		}
		cmds = append(cmds, UpdateTick())
	case ErrorCmd:
		m.addModal(modalError, msg.Error())
		cmds = append(cmds, WaitForError(ErrorCh))
	case MessageCmd:
		if msg == "" {
			m.removeModal()
		} else {
			m.addModal(modalInfo, string(msg))
		}
		cmds = append(cmds, WaitForMessage(MessageCh))
	case FastForwardCmd:
		cmds = append(cmds, m.handleRewind(int(msg)))
	case RemoveModalCmd:
		m.removeModal()
	case AddBookmarkCmd:
		m.ctrl.AddBookmark()
		cmds = append(cmds, SetBookmarks(m.ctrl.currentSource.BookmarksPerc()))
	case LoadSourceCmd:
		err := m.ctrl.selectSource(msg.id)
		if err != nil {
			ErrorCh <- err
		}
		cmds = append(cmds, SetBookmarks(m.ctrl.Source().BookmarksPerc()))
	case SetSourceCmd:
		m.mainView, cmd = m.mainView.Update(UpdateSourceCmd{source: msg.source})
		cmds = append(cmds, cmd)
	case DeleteSourceCmd:
		err := m.ctrl.deleteSource(msg.id)
		if err != nil {
			ErrorCh <- err
		}
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

// Rewind playback to given sentence number.
func (m *narrativeModel) handleRewind(sNum int) tea.Cmd {
	go m.ctrl.remote.Rewind(sNum)
	return WaitForFastForward(FastForwardCh)
}

func (m *narrativeModel) setTermDimensions(width, height int) {
	m.width = width
	m.height = height
}

// Render modal currently attached to the model.
func (m narrativeModel) renderModal() *lipgloss.Layer {
	if m.modal == nil {
		return nil
	}
	var style lipgloss.Style
	switch m.modal.t {
	case modalInfo:
		style = modalInfoStyle
	case modalError:
		style = modalErrorStyle
	default:
		panic("Unknown modal type")
	}
	modal := style.Width(defaultModalWidth).Render(m.modal.content)
	x := m.width - (lipgloss.Width(modal) + defaultModalXOffset)
	return lipgloss.NewLayer(modal).X(x).Y(defaultModalYOffset)
}

// Render internal information about playback
func (m narrativeModel) renderInfo() *lipgloss.Layer {
	var builder strings.Builder
	builder.WriteString("Stats for nerds:\n\n")
	builder.WriteString("Source:\n")
	if m.ctrl.Source().IsDummy() {
		builder.WriteString(fmt.Sprintf("  N/A\n"))
	} else {
		builder.WriteString(fmt.Sprintf("  Title: %s\n", m.ctrl.Source().ts.Title))
		builder.WriteString(fmt.Sprintf("  Author: %s\n", m.ctrl.Source().ts.Author))
		builder.WriteString(fmt.Sprintf("  ID: %s\n", m.ctrl.Source().id))
		builder.WriteString(fmt.Sprintf("  Path: %s\n", m.ctrl.Source().ts.Path))
		builder.WriteString(fmt.Sprintf("  Type: %s\n", m.ctrl.Source().ts.SourceType))
		builder.WriteString(fmt.Sprintf("  Created: %s\n",
			time.Unix(m.ctrl.Source().ts.Added, 0).Format("2006-01-02 15:04:05"),
		))
		builder.WriteString(fmt.Sprintf("  Total Sentences: %d\n", m.ctrl.Source().length))
		builder.WriteString(fmt.Sprintf("  Current Sentence: %d\n", m.ctrl.Source().SNum()))
		bmarks := []string{}
		for _, b := range m.ctrl.Source().Bookmarks() {
			bmarks = append(bmarks, strconv.Itoa(b))
		}
		builder.WriteString(fmt.Sprintf("  Bookmarks: %s", strings.Join(bmarks, ", ")))
	}

	builder.WriteString("\n")

	builder.WriteString(fmt.Sprintf("Selected Voice: %s\n", m.ctrl.cfg.Voice))
	builder.WriteString(fmt.Sprintf("Playback Speed: %.2f\n", m.ctrl.cfg.Speed))

	builder.WriteString("\n")

	model_n := m.ctrl.dataCfg.SelectedModel
	model_t, _ := common.KittenTypeFromString(model_n)
	model := common.GetKittenModel(model_t)
	builder.WriteString("Model:\n")
	builder.WriteString(fmt.Sprintf("  Name: %s\n", model_n))
	builder.WriteString(fmt.Sprintf("  Type: %d\n", model_t))
	builder.WriteString(fmt.Sprintf("  Remote: %s\n", model.Remote))
	builder.WriteString(fmt.Sprintf("  Fname: %s\n", model.Fname))

	builder.WriteString("\n")

	lib_n := m.ctrl.dataCfg.SelectedLib
	lib_d := common.GetOnnxLib(runtime.GOOS, runtime.GOARCH, lib_n)
	builder.WriteString("Library:\n")
	builder.WriteString(fmt.Sprintf("  Name: %s\n", lib_n))
	builder.WriteString(fmt.Sprintf("  OS: %s\n", lib_d.Os))
	builder.WriteString(fmt.Sprintf("  Arch: %s\n", lib_d.Arch))
	builder.WriteString(fmt.Sprintf("  Filename: %s\n", lib_d.Filename))
	builder.WriteString(fmt.Sprintf("  Remote: %s\n", lib_d.Remote))

	style := infoPanelStyle.Width(m.width - 8).Height(m.height - 8)
	return lipgloss.NewLayer(style.Render(builder.String())).X(4).Y(4)
}

// Attach new modal to the model.
func (m *narrativeModel) addModal(t modalType, content string) {
	m.modal = &Modal{
		t:       t,
		content: content,
	}
	m.modalTimeout = defaultModalTimeout
}

// Remove currently attached modal.
func (m *narrativeModel) removeModal() {
	m.modal = nil
	m.modalTimeout = 0
}

func (m narrativeModel) Init() tea.Cmd {
	return tea.Batch(
		UpdateTick(),
		WaitForPlayback(PlaybackCh),
		WaitForError(ErrorCh),
		WaitForMessage(MessageCh),
		WaitForFastForward(FastForwardCh),
	)
}

func (m narrativeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.setTermDimensions(msg.Width, msg.Height)
	case tea.KeyPressMsg:
		k := msg.String()
		if k == "f2" && !m.ctrl.Source().IsDummy() {
			m.ctrl.runDebugger = true
			return m, tea.Quit
		}
		if k == "f3" {
			m.showInfo = !m.showInfo
		}
	}

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
	if m.modal != nil {
		layers = append(
			layers,
			m.renderModal(),
		)
	}
	if m.showInfo {
		layers = append(layers, m.renderInfo())
	}
	comp := lipgloss.NewCompositor(layers...)
	v.SetContent(comp.Render())
	v.AltScreen = true
	v.WindowTitle = windowTitle
	return v
}
