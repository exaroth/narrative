package debugger

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/list"
)

var listHelpText = "<h/j/k/l>:Nav  <CR>:Select  <Space>:Play  <c>:Cont.Mode  ?:Help"

type OpenPhonemePanelCmd struct {
	currentSentence int
	currentWord     int
}

type sentenceList struct {
	sentences       *[]string
	currentSentence int
	currentWord     int
	list            viewport.Model
	ready           bool
	ctrl            *Debugger
	height          int
	width           int
}

func NewSentenceList(ctrl *Debugger, sentence_num int) *sentenceList {
	return &sentenceList{
		currentSentence: sentence_num,
		ctrl:            ctrl,
	}
}

func (s sentenceList) Init() tea.Cmd {
	return nil
}

func (s sentenceList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if k := msg.String(); k == "ctrl+c" || k == "q" || k == "esc" {
			return s, tea.Quit
		}
		if k := msg.String(); k == "j" {
			s.selectNextSentence()
			s.selectWord(0)
		}
		if k := msg.String(); k == "k" {
			s.selectPrevSentence()
			s.selectWord(0)
		}
		if k := msg.String(); k == "l" || k == "tab" {
			s.selectNextWord()
		}
		if k := msg.String(); k == "h" {
			s.selectPrevWord()
		}
		if k := msg.String(); k == "enter" {
			cmds = append(cmds, s.openPhonemePanel())
		}

	case tea.WindowSizeMsg:
		s.setTermDimensions(msg.Width, msg.Height)
		// assume status bar is height 1
		verticalMarginHeight := lipgloss.Height(s.sentencePanelView()) + 1

		if !s.ready {
			s.list = viewport.New(
				viewport.WithWidth(msg.Width),
				viewport.WithHeight(msg.Height-verticalMarginHeight),
			)
			s.list.KeyMap = GetSentenceListKeymap()
			s.list.SetContent(s.renderList())
			s.ready = true
		} else {
			s.list.SetWidth(msg.Width)
			s.list.SetHeight(msg.Height - verticalMarginHeight)
		}
	}

	s.list, cmd = s.list.Update(msg)
	cmds = append(cmds, cmd)

	return s, tea.Batch(cmds...)
}

func (s sentenceList) View() tea.View {
	var v tea.View
	if !s.ready {
		v.SetContent("\n  Initializing...")
	} else {
		v.SetContent(lipgloss.Sprintf("%s\n%s\n%s",
			s.list.View(),
			s.sentencePanelView(),
			getStatusBar(listHelpText, s.width),
		))
		s.ready = true
	}
	return v
}

func (s *sentenceList) setTermDimensions(w int, h int) {
	s.width = w
	s.height = h
}

func (s *sentenceList) openPhonemePanel() tea.Cmd {
	return func() tea.Msg {
		return OpenPhonemePanelCmd{
			currentSentence: s.currentSentence,
			currentWord:     s.currentWord,
		}
	}
}

func (s sentenceList) renderList() string {

	l := list.New().
		Enumerator(list.Arabic).
		ItemStyleFunc(func(_ list.Items, i int) lipgloss.Style {
			st := sentenceListBaseStyle
			if s.currentSentence == i {
				return st.Foreground(sentenceListHighlightColor)
			}
			return st.Foreground(sentenceListDimColor)
		}).
		EnumeratorStyleFunc(func(_ list.Items, i int) lipgloss.Style {
			if s.currentSentence == i {
				return lipgloss.NewStyle().Foreground(sentenceListHighlightColor)
			}
			return lipgloss.NewStyle().Foreground(sentenceListDimColor)
		})

	for _, d := range s.ctrl.source {
		l.Item(d)
	}

	return lipgloss.Sprint("\n", l, "\n")
}

func (s *sentenceList) generateSentenceTranscription(use_phonemes bool) string {
	var builder strings.Builder
	s.selectSentence(s.currentSentence)
	sentence_data := s.ctrl.sentenceData[s.currentSentence]
	if sentence_data == nil {
		panic("No opts found")
	}
	var words []string
	var style lipgloss.Style
	if use_phonemes {
		style = sentencePanelPhonemeStyle
		for _, e := range sentence_data.selectedPhonemes {
			words = append(words, e[1])
		}
	} else {
		style = sentencePanelWordStyle
		words = *sentence_data.opts.WordOrigins
	}
	punctuation := sentence_data.punctuation.AsArr(len(words))

	var w, pre_punct, post_punct, num string
	for idx, word := range words {
		num = lipgloss.Sprint(sentencePanelNumStyle.Render(fmt.Sprintf("(%d)", idx+1)))
		if idx == s.currentWord {
			w = lipgloss.Sprint(sentencePanelSelectedWordStyle.Render(word))
		} else {
			w = lipgloss.Sprint(style.Render(word))
		}
		pre_punct = lipgloss.Sprint(sentencePanelPunctStyle.Render(punctuation[idx][0]))
		post_punct = lipgloss.Sprint(sentencePanelPunctStyle.Render(punctuation[idx][1]))

		if len(word) > 0 {
			builder.WriteString(num)
		}

		builder.WriteString(pre_punct)
		builder.WriteString(w)
		builder.WriteString(post_punct)
		builder.WriteString(" ")
	}
	if use_phonemes {
		return lipgloss.Sprintf("Phonemes %s", builder.String())
	}
	return lipgloss.Sprintf(" Words    %s", builder.String())
}

func (s *sentenceList) selectPrevSentence() {
	s.selectSentence(s.currentSentence - 1)
}

func (s *sentenceList) selectNextSentence() {
	s.selectSentence(s.currentSentence + 1)
}

func (s *sentenceList) selectSentence(n int) {
	if n < 0 {
		n = 0
	}
	if n >= len(s.ctrl.source) {
		n = len(s.ctrl.source) - 1
	}
	s.currentSentence = n
	s.ctrl.getSentenceData(uint(n))

	s.list.SetContent(s.renderList())
}

func (s *sentenceList) selectWord(w_n int) {
	var data = (*s.ctrl.sentenceData[s.currentSentence])
	var words = data.opts.WordOrigins

	if w_n < 0 {
		w_n = len(*words) - 1
	}
	if w_n >= len(*words) {
		w_n = 0
	}
	s.currentWord = w_n
}

func (s *sentenceList) selectNextWord() {
	s.selectWord(s.currentWord + 1)
}

func (s *sentenceList) selectPrevWord() {
	s.selectWord(s.currentWord - 1)
}

func (s sentenceList) sentencePanelView() string {

	output := sentencePanelStyle.Width(s.list.Width()).Render(
		s.generateSentenceTranscription(false),
		"\n\n",
		s.generateSentenceTranscription(true),
	)
	return lipgloss.Sprint(output)
}
