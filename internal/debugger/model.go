package debugger

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

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
	isPlaying    bool
	ctrl         *Debugger
	sentenceList tea.Model
	phonemePanel tea.Model
	helpPanel    *helpPanel
}

type UpdateTickMsg time.Time

type SetPlaybackStatusMsg bool

func UpdateTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return UpdateTickMsg(t)
	})
}

func NewDebuggerModel(controller *Debugger, sentence_idx int) tea.Model {
	return &debugModel{
		ctrl:         controller,
		mode:         listMode,
		sentenceList: NewSentenceList(controller, sentence_idx),
		helpPanel:    &helpPanel{},
	}
}

func (m debugModel) Init() tea.Cmd {
	return tea.Batch(
		UpdateTick(),
		m.sentenceList.Init(),
	)
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
	case PlaySentenceCmd:
		if msg.continuous {
			m.ctrl.play(msg.sentence, "", func() {
				playbackCh <- struct{}{}
			})
		} else {
			if m.isPlaying {
				m.ctrl.stop()
			} else {
				go m.ctrl.play(msg.sentence, "", nil)
			}
		}
	case PlayPhonemeCmd:
		m.ctrl.play(msg.phoneme, ".", nil)
	case StopPlaybackCmd:
		m.ctrl.stop()
	case SetPlaybackStatusMsg:
		m.isPlaying = bool(msg)
	case UpdatePhonemeCmd:
		m.updateExtDict(msg.word, msg.phoneme, msg.sentenceNum)
		m.setListMode()
		if msg.reopen {
			m.setPhonemeMode(msg.sentenceNum, msg.wordNum)
		}
	case tea.KeyPressMsg:
		if k := msg.String(); k == "?" || k == "f1" {
			m.toggleHelp()
		}
	case DeletePhonemeCmd:
		m.deleteFromExtDict(msg.word, msg.sentenceNum)
		m.setListMode()
		m.setPhonemeMode(msg.sentenceNum, msg.wordNum)
	case UpdateMissingDictCmd:
		m.ctrl.updateMissingDict(msg.values)
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
	v.AltScreen = true
	switch m.mode {
	case listMode:
		v = m.sentenceList.View()
	case phonemeMode:
		v = m.phonemePanel.View()
	case helpMode:
		v.SetContent(m.helpPanel.Render())
	}
	return v
}

func (m *debugModel) updateExtDict(word, phoneme string, sentence_n int) {
	m.ctrl.updateExtDict(word, phoneme)
	m.ctrl.clearCache(sentence_n)
	m.ctrl.getSentenceData(uint(sentence_n))
}

func (m *debugModel) deleteFromExtDict(word string, sentence_n int) {
	m.ctrl.deleteFromExtDict(word)
	m.ctrl.clearCache(sentence_n)
	m.ctrl.getSentenceData(uint(sentence_n))
}

func (m *debugModel) setTermDimensions(w int, h int) {
	m.width = w
	m.height = h
	m.helpPanel.width = w
	m.helpPanel.height = h
}

func (m *debugModel) setPhonemeMode(sentence_num, word_num int) {
	m.phonemePanel = OpenPhonemePanel(m.ctrl, sentence_num, word_num, m.width, m.height)
	m.mode = phonemeMode
}

func (m *debugModel) setListMode() {
	m.mode = listMode
	m.phonemePanel = nil
}

func (m *debugModel) toggleHelp() {
	if m.mode == helpMode {
		m.mode = listMode
	} else {
		m.mode = helpMode
	}
}
