package narrative

import (
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// This is a model for main narrative view containing
// file list, transcription, playback info etc.
type mainViewModel struct {
	showTranscript bool
	sources        Sources
	currentSource  *Source
	list           list.Model
}

// Initialize new narrative model.
func NewMainViewModel(source_list Sources, source *Source) *mainViewModel {
	l := list.New(source_list.ListItems(), NewDelegate(), 0, 0)
	l.Title = "Narrative v0.1"
	l.Styles.Title = TitleStyle
	return &mainViewModel{
		showTranscript: true,
		sources:        source_list,
		currentSource:  source,
		list:           l,
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
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

var docStyle = lipgloss.NewStyle().Margin(1, 2)

func (m mainViewModel) View() tea.View {
	v := tea.NewView(docStyle.Render(m.list.View()))
	v.AltScreen = true
	return v
}
