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

// This is a model for main narrative view containing
// file list, transcription, playback info etc.
type mainViewModel struct {
	showTranscript bool
	sources        Sources
	currentSource  *Source
	list           list.Model
	progress       ProgressModel
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
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if k := msg.String(); k == "ctrl+c" || k == "q" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()

		m.progress.SetWidth(msg.Width - 5)
		m.list.SetSize(msg.Width-h, msg.Height-v-10) // 10 is progress
	// case ToggleSourceCmd:
	// 	panic(m.currentSource)
	case UpdateSourceCmd:
		m.updateSource(msg.source)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
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
	v.SetContent(list + "\n" + progress)
	v.AltScreen = true
	return v
}
