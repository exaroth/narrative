package narrative

import (
	"fmt"
	"reflect"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/sirupsen/logrus"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

const (
	appTitle = "Narrative v0.1"
)

// This is a model for main narrative view containing
// file list, transcription, playback info etc.
type mainViewModel struct {
	sources        Sources
	currentSource  *Source
	list           list.Model
	progress       ProgressModel
	showTranscript bool
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
	logrus.Info(fmt.Sprintf("Msg: %v", reflect.TypeOf(msg)))
	logrus.Info(msg)
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if k := msg.String(); k == "ctrl+c" || k == "q" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()

		m.progress.SetWidth(msg.Width - 5)
		m.list.SetSize(msg.Width-h, msg.Height-v-10) // 10 is progress
	case SelectSourceCmd:
		if msg.id == m.currentSource.id {
			cmds = append(cmds, TogglePlayback())
		} else {
			cmds = append(cmds, LoadSource(msg.id))
			// cmds = append(cmds, m.startPlayback())
		}
	case TogglePlaybackCmd:
		switch PSM.Status() {
		case playbackPlaying:
			PlaybackCh <- 0
		case playbackPaused:
			PlaybackCh <- 1
		case playbackIdle:

			cmds = append(cmds, LoadSource(m.currentSource.id))
		}
	case UpdateSourceCmd:
		m.updateSource(msg.source)
	}

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
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
	v.SetContent(list + "\n" + progress + "\n" + m.currentSource.id + "/" + PSM.Status().String())
	return v
}
