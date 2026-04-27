package narrative

import (
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

const (
	appTitle = "Narrative v0.1"
)

type playbackStatus int

const (
	playbackPlaying playbackStatus = iota
	playbackPaused
)

// This is a model for main narrative view containing
// file list, transcription, playback info etc.
type mainViewModel struct {
	sources        Sources
	currentSource  *Source
	list           list.Model
	progress       ProgressModel
	showTranscript bool
	playbackStatus playbackStatus
}

// Initialize new narrative model.
func NewMainViewModel(source_list Sources) *mainViewModel {
	l := list.New(source_list.ListItems(), NewDelegate(), 0, 0)
	l.SetShowHelp(false)
	l.Title = appTitle
	l.Styles.Title = TitleStyle

	p := NewProgress(WithDefaultBlend())

	return &mainViewModel{
		showTranscript: true,
		sources:        source_list,
		currentSource:  nil,
		list:           l,
		progress:       p,
	}
}

func (m mainViewModel) Init() tea.Cmd {
	return nil
}

func (m mainViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if k := msg.String(); k == "ctrl+c" || k == "q" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()

		m.progress.SetWidth(msg.Width - 5)
		m.list.SetSize(msg.Width-h, msg.Height-v-10) // 10 is progress
	case ToggleSourceCmd:
		if msg.id == m.currentSource.id {
			if m.playbackStatus == playbackPlaying {
				cmds = append(cmds, m.pausePlayback())
			} else {
				cmds = append(cmds, m.startPlayback())
			}
		} else {
			cmds = append(cmds, LoadSource(msg.id))
			// cmds = append(cmds, m.startPlayback())
		}
	case UpdateSourceCmd:
		m.updateSource(msg.source)
	}

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

// Start playback of current source.
func (m *mainViewModel) startPlayback() tea.Cmd {
	m.playbackStatus = playbackPlaying
	return PlaySource()
}

// Pause current playback.
func (m *mainViewModel) pausePlayback() tea.Cmd {
	m.playbackStatus = playbackPaused
	return PauseSource()
}

// Update current source.
func (m *mainViewModel) updateSource(source *Source) {
	m.currentSource = source
}

func (m mainViewModel) View() tea.View {
	// v := tea.NewView(docStyle.Render(m.list.View()))
	list := docStyle.Render(m.list.View())
	progress := docStyle.Render(m.progress.View())
	var v tea.View
	v.SetContent(list + "\n" + progress + "\n" + m.currentSource.id)
	v.AltScreen = true
	return v
}
