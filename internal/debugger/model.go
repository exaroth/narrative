package debugger

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
)

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type mode int

const (
	listMode mode = iota
	phonemeMode
	helpMode
)

type debugModel struct {
	mode         mode
	width        int
	height       int
	ctrl         *Debugger
	sentenceList tea.Model
	phonemePanel tea.Model
}

func NewDebuggerModel(controller *Debugger, sentence_idx int) tea.Model {
	return &debugModel{
		ctrl:         controller,
		mode:         listMode,
		sentenceList: NewSentenceList(controller, sentence_idx),
	}
}

func (m debugModel) Init() tea.Cmd {
	// m.selectSentence(m.currentSentence)
	// return tick()
	return nil
}

func (m debugModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.setTermDimensions(msg.Width, msg.Height)
	case OpenPhonemePanelCmd:
		m.setPhonemeMode(msg.currentSentence, msg.currentWord)
	case ClosePhonemePanelCmd:
		m.setListMode()
	}

	switch m.mode {
	case listMode:
		m.sentenceList, cmd = m.sentenceList.Update(msg)
	case phonemeMode:
		m.phonemePanel, cmd = m.phonemePanel.Update(msg)
	}

	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m debugModel) View() tea.View {
	var v tea.View
	switch m.mode {
	case listMode:
		v.SetContent(fmt.Sprintf("%s\n", m.sentenceList.View()))
	case phonemeMode:
		v.SetContent(fmt.Sprintf("%s", m.phonemePanel.View()))
	}
	return v
}

func (m *debugModel) updateExtDict(word, phoneme string, sentence_n int) {
	m.ctrl.updateExtDict(word, phoneme)
	// m.ctrl.clearCache(m.currentSentence)
}

func (m *debugModel) setTermDimensions(w int, h int) {
	m.width = w
	m.height = h
}

func (m *debugModel) setPhonemeMode(sentence_num, word_num int) {
	m.phonemePanel = OpenPhonemePanel(m.ctrl, sentence_num, word_num)
	m.mode = phonemeMode
}

func (m *debugModel) setListMode() {
	m.mode = listMode
	m.phonemePanel = nil
}
