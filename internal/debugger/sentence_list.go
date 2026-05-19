package debugger

import (
	"fmt"
	"slices"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/list"
)

var listHelpText = "<h/j/k/l>:Nav  <CR>:Select  <Space>:Play  <c>:Cont.Mode  <Arrows>:Scroll  Q:Quit  ?:Help"

// Commands

type (
	OpenPhonemePanelCmd struct {
		currentSentence int
		currentWord     int
	}

	PlaySentenceCmd struct {
		sentence   string
		continuous bool
	}

	UpdateMissingDictCmd struct {
		values map[string]string
	}

	StopPlaybackCmd struct{}
	PlayNextCmd     struct{}
)

// channel controlling when we should send next sentence
// to play.
var playbackCh = make(chan struct{})

type sentenceList struct {
	sentences       *[]string
	currentSentence int
	currentWord     int
	list            viewport.Model
	ready           bool
	ctrl            *Debugger
	height          int
	width           int
	continuousMode  bool
	// list of sentences marked for review
	marks []int
}

func waitForPlaybackEnd(sub chan struct{}) tea.Cmd {
	return func() tea.Msg {
		return PlayNextCmd(<-sub)
	}
}

func NewSentenceList(ctrl *Debugger, sentence_num int) *sentenceList {
	return &sentenceList{
		currentSentence: sentence_num,
		ctrl:            ctrl,
	}
}

func (s sentenceList) Init() tea.Cmd {

	return tea.Batch(
		waitForPlaybackEnd(playbackCh),
	)
}

func (s sentenceList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if k := msg.String(); k == "ctrl+c" || k == "Q" {
			return s, tea.Quit
		}
		if !s.continuousMode {
			if k := msg.String(); k == "tab" {
				cmds = append(cmds, s.selectNextSentence(true))
				s.selectWord(0)
				cmds = append(cmds, s.playCurrentSentence())
			}
			if k := msg.String(); k == "j" {
				cmds = append(cmds, s.selectNextSentence(true))
				s.selectWord(0)
			}
			if k := msg.String(); k == "k" {
				cmds = append(cmds, s.selectPrevSentence(true))
				s.selectWord(0)
			}
			if k := msg.String(); k == "l" {
				s.selectNextWord()
			}
			if k := msg.String(); k == "h" {
				s.selectPrevWord()
			}
			if k := msg.String(); k == "enter" {
				cmds = append(cmds, s.openPhonemePanel())
			}
		}
		if k := msg.String(); k == "m" {
			s.markForReview()
		}

		if k := msg.String(); k == "c" {
			if s.continuousMode {
				cmds = append(cmds, s.stopContinuousMode())
			} else {
				cmds = append(cmds, s.startContinuousPlay())
			}
		}
		if k := msg.String(); k == "space" {
			if s.continuousMode {
				cmds = append(cmds, s.stopContinuousMode())
			} else {
				cmds = append(cmds, s.playCurrentSentence())
			}
		}

	case PlayNextCmd:
		cmds = append(
			cmds,
			s.playNextSentence(),
			waitForPlaybackEnd(playbackCh))

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
		var r_contents string
		if s.continuousMode {
			r_contents = "C"
		}
		v.SetContent(lipgloss.Sprintf("%s\n%s\n%s",
			s.list.View(),
			s.sentencePanelView(),
			getStatusBar(listHelpText, r_contents, s.width),
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

// Play Current sentence only.
func (s *sentenceList) playCurrentSentence() tea.Cmd {

	sentence_data := s.ctrl.sentenceData[s.currentSentence]
	return func() tea.Msg {
		return PlaySentenceCmd{
			sentence:   sentence_data.phonemized,
			continuous: s.continuousMode,
		}
	}
}

func (s *sentenceList) playNextSentence() tea.Cmd {
	return tea.Batch(
		s.selectNextSentence(true),
		s.playCurrentSentence(),
	)
}

// Start playing sentences, starting with current one.
func (s *sentenceList) startContinuousPlay() tea.Cmd {

	s.continuousMode = true
	sentence_data := s.ctrl.sentenceData[s.currentSentence]
	return func() tea.Msg {
		return PlaySentenceCmd{
			sentence:   sentence_data.phonemized,
			continuous: true,
		}
	}
}

func (s *sentenceList) stopContinuousMode() tea.Cmd {
	s.continuousMode = false
	return func() tea.Msg {
		return StopPlaybackCmd{}
	}
}

func (s *sentenceList) markForReview() {
	if slices.Contains(s.marks, s.currentSentence) {
		// unmark if exists
		idx := slices.Index(s.marks, s.currentSentence)
		s.marks = append(s.marks[:idx], s.marks[idx+1:]...)
		return
	}
	s.marks = append(s.marks, s.currentSentence)
}

// Render sentence list.
func (s sentenceList) renderList() string {
	l := list.New().
		Enumerator(list.Arabic).
		ItemStyleFunc(func(_ list.Items, i int) lipgloss.Style {
			st := sentenceListBaseStyle
			if slices.Contains(s.marks, i) {
				return st.Foreground(sentenceListMarkedColor)
			} else if s.currentSentence == i {
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

// Render introspection panel below the list containing words and selectable phonemes.
func (s *sentenceList) generateSentenceTranscription(use_phonemes bool) string {
	var builder strings.Builder
	s.selectSentence(s.currentSentence, false)
	sentence_data := s.ctrl.sentenceData[s.currentSentence]
	if sentence_data == nil {
		panic("No opts found")
	}
	var words []string
	var style lipgloss.Style
	if use_phonemes {
		style = sentencePanelPhonemeStyle
		words = s.getSelectedPhonemes()
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

func (s *sentenceList) selectPrevSentence(update bool) tea.Cmd {
	return s.selectSentence(s.currentSentence-1, update)
}

func (s *sentenceList) selectNextSentence(update bool) tea.Cmd {
	return s.selectSentence(s.currentSentence+1, update)
}

func (s *sentenceList) selectSentence(n int, update bool) tea.Cmd {
	if n < 0 {
		n = 0
	}
	if n >= len(s.ctrl.source) {
		n = len(s.ctrl.source) - 1
	}
	s.currentSentence = n
	s.ctrl.getSentenceData(uint(n))

	s.list.SetContent(s.renderList())
	if update {
		return s.updateMissingDict()
	}
	return nil
}

func (s *sentenceList) updateMissingDict() tea.Cmd {
	sentence_data := s.ctrl.sentenceData[s.currentSentence]
	if sentence_data == nil {
		return nil
	}
	res := make(map[string]string)
	for idx, t := range *sentence_data.opts.Tags {
		// if theres more than 1 phoneme available
		// inferrence would not run (kw)
		if len(t) > 1 {
			continue
		}
		for ph, tt := range t {
			// this will happen for lone punctuation
			if len(ph) == 0 || len(ph) == 1 {
				break
			}
			if len(tt) == 0 {
				res[(*sentence_data.opts.WordOrigins)[idx]] = ph
			}
			break
		}
	}

	return func() tea.Msg {
		return UpdateMissingDictCmd{res}
	}
}

// Get list of selected phonemes for current sentence.
func (s *sentenceList) getSelectedPhonemes() []string {
	result := []string{}
	sentence_data := s.ctrl.sentenceData[s.currentSentence]
	for _, e := range sentence_data.selectedPhonemes {
		result = append(result, e[1])
	}
	return result
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
