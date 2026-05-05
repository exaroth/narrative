package narrative

import (
	"time"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const (
	appTitle         = "ℕarrative v0.1"
	transcriptHeight = 4
)

var statusHelp = [][2]string{
	{"?", "help"},
	{"/", "search"},
	{"space", "play/pause"},
	{"enter", "select"},
	{"j/k", "nav"},
	{"h/l", "seek"},
	{"d", "delete"},
	{"q", "quit"},
	{"ctrl-b", "set bookmark"},
	{"b", "go to bookmark"},
}

// This is a model for main narrative view containing
// file list, transcription, playback info etc.
type mainViewModel struct {
	sources        Sources
	currentSource  *Source
	list           list.Model
	progress       ProgressModel
	perc           *percRead
	statusBar      *statusBar
	fForwarder     *fastForwarder
	transcript     *transcript
	helpPanel      *helpPanel
	prompt         *confirmPrompt
	showTranscript bool
	showHelp       bool
}

// Initialize new narrative model.
func NewMainViewModel(source_list Sources) *mainViewModel {
	l := list.New(source_list.ListItems(), NewDelegate(), 0, 0)
	l.SetShowHelp(false)
	l.Title = appTitle
	l.Styles.Title = TitleStyle
	l.KeyMap = ListKeymap()
	p := NewProgress(col.Color2)
	return &mainViewModel{
		sources:        source_list,
		currentSource:  nil,
		list:           l,
		progress:       p,
		perc:           &percRead{},
		statusBar:      &statusBar{},
		transcript:     &transcript{height: transcriptHeight},
		showTranscript: true,
	}
}

// Update current source.
func (m *mainViewModel) updateSource(source *Source) {
	m.currentSource = source
}

// Toggle help panel on/off.
func (m *mainViewModel) toggleHelp() {
	m.showHelp = !m.showHelp
}

func (m mainViewModel) Init() tea.Cmd {
	return nil
}

func (m mainViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	// logrus.Info(fmt.Sprintf("Msg: %v", reflect.TypeOf(msg)))
	// logrus.Info(msg)
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		k := msg.String()
		if m.prompt != nil {
			if k == "y" || k == "enter" {
				cmds = append(cmds, m.prompt.Confirm(), ClosePrompt())
			}
			if k == "n" || k == "esc" {
				cmds = append(cmds, m.prompt.Decline(), ClosePrompt())
			}
		} else {
			if k == "ctrl+c" || k == "q" {
				cmds = append(cmds, ShowPrompt("Quit?", tea.Quit, nil))
			}
			if k == "?" || k == "f1" {
				m.toggleHelp()
			}
			if k == "esc" {
				if m.showHelp {
					m.toggleHelp()
				}
			}
			if PSM.AllowsRewinding() && k == "l" || k == "h" {
				if m.fForwarder == nil {
					m.fForwarder = newFastForwarder(
						0,
						m.currentSource.Length()-1,
						m.currentSource.SNum(),
					)
				}
				m.fForwarder.Update(k == "l")
			}
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		listWidth, listHeight := msg.Width-h, msg.Height-v-10 // 10 is progress
		m.progress.SetWidth(msg.Width - 5)
		m.transcript.SetWidth(msg.Width)
		if m.showTranscript {
			listHeight = listHeight - lipgloss.Height(m.transcript.Render(m.currentSource))
		}
		m.list.SetSize(listWidth, listHeight)
		m.perc = m.perc.SetWidth(msg.Width)
		m.statusBar = m.statusBar.SetWidth(msg.Width)
		m.helpPanel = &helpPanel{msg.Width, msg.Height}
	case UpdateTickMsg:
		cmds = append(
			cmds,
			m.progress.SetPercent(m.currentSource.PercRead()),
		)
		m.perc = m.perc.SetPerc(m.currentSource.PercRead())
		if m.fForwarder != nil {
			c, done := m.fForwarder.Tick(time.Time(msg))
			if done {
				FastForwardCh <- c
				m.fForwarder = nil
			}
		}
	case ShowPromptCmd:
		m.showPrompt(msg)
	case ClosePromptCmd:
		m.closePrompt()
	case UpdateSourceListCmd:
		cmds = append(cmds, m.list.SetItems(msg.items))
	case SelectSourceCmd:
		if msg.id == m.currentSource.id {
			if msg.autoplay {
				cmds = append(cmds, TogglePlayback())
			}
			if msg.index > -1 {
				m.list.Select(msg.index)
			}
		} else {
			cmds = append(cmds, LoadSource(msg.id))
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
	var v tea.View
	if m.showHelp {
		v.SetContent(m.helpPanel.Render())
		return v
	}
	list := docStyle.Render(m.list.View())
	progress := docStyle.Render(m.progress.View())
	var status_b string
	if m.prompt != nil {
		status_b = m.statusBar.Prompt(m.prompt.text)
	} else {
		status_b = m.statusBar.RenderHelp()
	}
	v.SetContent(status_b + "\n" +
		list + "\n" +
		m.transcript.Render(m.currentSource) + "\n" +
		progress + "\n" +
		m.perc.Render())
	return v
}

// Set prompt data on the model.
func (m *mainViewModel) showPrompt(msg ShowPromptCmd) {
	if m.prompt == nil {
		m.prompt = &confirmPrompt{
			text:    msg.text,
			okFunc:  msg.okFunc,
			nayFunc: msg.nayFunc,
		}
	}
}

// Remove prompt data from the model.
func (m *mainViewModel) closePrompt() {
	m.prompt = nil
}
