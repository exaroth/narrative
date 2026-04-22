package narrative

import tea "charm.land/bubbletea/v2"

// This is a model for main narrative view containing
// file list, transcription, playback info etc.
type mainViewModel struct {
	showTranscript bool
}

// Initialize new narrative model.
func NewMainViewModel() *mainViewModel {
	// TODO.
	return &mainViewModel{
		showTranscript: true,
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
	}
	return m, nil
}

func (m mainViewModel) View() tea.View {
	var v tea.View
	v.SetContent("\n main view")
	return v
}
