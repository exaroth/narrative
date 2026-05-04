package narrative

import (
	"fmt"
	"time"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var (
	SpinnerCloseCh   = make(chan struct{})
	SpinnerWaitCh    = make(chan struct{})
	SpinnerMessageCh = make(chan string)
	spinnerP         *tea.Program
)

// Gracefully close spinner.
func CloseSpinner() {
	go func() {
		SpinnerCloseCh <- struct{}{}
	}()
	SpinnerWaitCh <- struct{}{}
}

// Model for displaying spinner during loading.
type spinnerModel struct {
	text     string
	spinner  spinner.Model
	quitting bool
}

func startSpinner(text string) {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(col.Color2)
	model := spinnerModel{spinner: s, text: text}
	spinnerP = tea.NewProgram(model)
	spinnerP.Run()
}

func (m spinnerModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		WaitForSpinnerClose(SpinnerCloseCh),
		WaitForSpinnerMessage(SpinnerMessageCh),
	)
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case SpinnerCloseCmd:
		m.quitting = true
		go func() {
			time.Sleep(100 * time.Millisecond)
			spinnerP.Kill()
			<-SpinnerWaitCh
		}()
		return m, tea.Quit
	case SpinnerMsgCmd:
		m.text = string(msg)
	}
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m spinnerModel) View() tea.View {
	if m.quitting {
		return tea.NewView("\n")
	}
	str := fmt.Sprintf("\n\n   %s %s\n\n", m.spinner.View(), m.text)
	return tea.NewView(str)
}
