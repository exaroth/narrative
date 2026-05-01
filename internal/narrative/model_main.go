package narrative

import (
	"fmt"
	"reflect"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/sirupsen/logrus"
)

const (
	appTitle        = "ℕarrative v0.1"
	statusHelpShort = "short"
	statusHelpLong  = "long"
)

// This is a model for main narrative view containing
// file list, transcription, playback info etc.
type mainViewModel struct {
	sources        Sources
	currentSource  *Source
	list           list.Model
	progress       ProgressModel
	perc           *percRead
	statusBar      *statusBar
	showTranscript bool
}

// Initialize new narrative model.
func NewMainViewModel(source_list Sources) *mainViewModel {
	l := list.New(source_list.ListItems(), NewDelegate(), 0, 0)
	l.SetShowHelp(false)
	l.Title = appTitle
	l.Styles.Title = TitleStyle

	p := NewProgress(WithoutPercentage(), WithColors(col.Color2))
	return &mainViewModel{
		showTranscript: true,
		sources:        source_list,
		currentSource:  nil,
		list:           l,
		progress:       p,
		perc:           &percRead{},
		statusBar:      &statusBar{},
	}
}

// Update current source.
func (m *mainViewModel) updateSource(source *Source) {
	m.currentSource = source
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
		m.perc = m.perc.SetWidth(msg.Width)
		m.statusBar = m.statusBar.SetWidth(msg.Width)
	case UpdateTickMsg:
		cmds = append(
			cmds,
			m.progress.SetPercent(m.currentSource.PercRead()),
		)
		m.perc = m.perc.SetPerc(m.currentSource.PercRead())
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
	m.progress, cmd = m.progress.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m mainViewModel) View() tea.View {
	list := docStyle.Render(m.list.View())
	progress := docStyle.Render(m.progress.View())
	var v tea.View
	v.SetContent(m.statusBar.Render() + "\n" + list + "\n" + progress + "\n" + m.perc.Render())
	return v
}

// Displays percent of source read.
type percRead struct {
	width int
	perc  float64
}

func (p *percRead) SetWidth(w int) *percRead {
	p.width = w
	return p
}

func (p *percRead) SetPerc(perc float64) *percRead {
	p.perc = perc
	return p
}

func (p *percRead) Render() string {
	perc := percReadStyle.Render(fmt.Sprintf(" %3.0f%% Read", p.perc*100))
	return lipgloss.Place(p.width, 1, lipgloss.Center, lipgloss.Center, perc)
}

// Status bar for the main app.
type statusBar struct {
	width  int
	status playbackStatus
}

func (s *statusBar) SetWidth(w int) *statusBar {
	s.width = w
	return s
}

func (s *statusBar) Render() string {
	col := GetStatusBarStatusColor(PSM.Status())
	status := statusBarStatusStyle.Background(col).Render(" " + PSM.Status().StatusString())
	w := s.width - lipgloss.Width(status)
	h := lipgloss.Place(w, 1, lipgloss.Left, lipgloss.Center, "  "+statusHelpShort)

	help := statusBarStatusText.Render(h)

	bar := lipgloss.JoinHorizontal(lipgloss.Top,
		help,
		status,
	)
	return statusBarStyle.Width(s.width).Render(bar)
}
