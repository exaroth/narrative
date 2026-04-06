package debugger

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/list"
	"github.com/exaroth/narrative/bubbles/viewport"
)

type tickMsg time.Time

// sentence list styles
var (
	sentenceListDimColor       = lipgloss.Color("250")
	sentenceListHighlightColor = lipgloss.Color("228")
	sentenceListBaseStyle      = lipgloss.NewStyle().MarginBottom(1).MarginLeft(1).Inline(true)
)

// sentence introspection styles
var (
	sentencePanelPunctStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFF"))
	sentencePanelNumStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
	sentencePanelWordStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("110"))
	sentencePanelPhonemeStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("222"))
	sentencePanelSelectedWordStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#F00"))

	sentencePanelStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		return lipgloss.NewStyle().
			MarginBottom(1).
			Height(12).
			BorderStyle(b).
			BorderForeground(lipgloss.Color("237")).
			PaddingLeft(2).PaddingTop(1).PaddingRight(1)
	}()
)

// title styles
var (
	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().Bold(true).BorderStyle(b).Padding(0, 1)
	}()
)

// Phoneme view styles

var (
	phonemeViewBaseStyle          = lipgloss.NewStyle().Padding(1)
	phonemeViewHeaderWordStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFF"))
	phonemeViewHeaderPhonemeStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFF"))
	phonemeViewSelectedPhoneme    = lipgloss.NewStyle().Foreground(lipgloss.Color("#F00"))
)

type mainViewModel struct {
	ready           bool
	phonemeView     bool
	currentSentence int
	currentWord     int
	selectedPhoneme int
	width           int
	height          int
	ctrl            *Debugger
	sentenceList    viewport.Model
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func NewDebuggerModel(controller *Debugger, sentence_idx int) tea.Model {
	return &mainViewModel{
		ctrl:            controller,
		currentSentence: sentence_idx,
	}
}

func (m mainViewModel) Init() tea.Cmd {
	m.selectSentence(m.currentSentence)
	// return tick()
	return nil
}

func (m mainViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if m.phonemeView {
			if k := msg.String(); k == "q" || k == "esc" {
				m.togglePhonemeView()
			}
			if k := msg.String(); k == "j" {
				m.selectPrevPhoneme()
			}
			if k := msg.String(); k == "k" || k == "tab" {
				m.selectNextPhoneme()
			}
		} else {
			if k := msg.String(); k == "ctrl+c" || k == "q" || k == "esc" {
				return m, tea.Quit
			}
			if k := msg.String(); k == "j" {
				m.selectNextSentence()
				m.selectWord(0)
			}
			if k := msg.String(); k == "k" {
				m.selectPrevSentence()
				m.selectWord(0)
			}
			if k := msg.String(); k == "l" || k == "tab" {
				m.selectNextWord()
			}
			if k := msg.String(); k == "h" {
				m.selectPrevWord()
			}
			if k := msg.String(); k == "enter" {
				m.togglePhonemeView()
				m.selectDefaultPhoneme()
			}
		}
	case tea.WindowSizeMsg:
		m.setTermDimensions(msg.Width, msg.Height)
		headerHeight := lipgloss.Height(m.headerView())
		sentencePanelHeight := lipgloss.Height(m.sentencePanelView())
		verticalMarginHeight := headerHeight + sentencePanelHeight

		if !m.ready {
			m.sentenceList = viewport.New(
				viewport.WithWidth(msg.Width),
				viewport.WithHeight(msg.Height-verticalMarginHeight),
			)
			m.sentenceList.KeyMap = GetSentenceListKeymap()
			m.sentenceList.YPosition = headerHeight
			m.sentenceList.SetContent(m.renderSentenceList())
			m.ready = true
		} else {
			m.sentenceList.SetWidth(msg.Width)
			m.sentenceList.SetHeight(msg.Height - verticalMarginHeight)
		}
	case tickMsg:
		fmt.Println("tick")
		return m, tick()
	}

	// update viewport on input
	m.sentenceList, cmd = m.sentenceList.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m mainViewModel) View() tea.View {
	var v tea.View
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	if !m.ready {
		v.SetContent("\n  Initializing...")
	} else if m.phonemeView {
		v.SetContent(fmt.Sprintf("%s\n%s", m.headerView(), m.phonemePanelView()))
	} else {
		v.SetContent(fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.sentenceList.View(), m.sentencePanelView()))
		m.ready = true
	}
	return v
}

func (m *mainViewModel) selectPrevSentence() {
	m.selectSentence(m.currentSentence - 1)
}

func (m *mainViewModel) selectNextSentence() {
	m.selectSentence(m.currentSentence + 1)
}

func (m *mainViewModel) selectSentence(n int) {
	if n < 0 {
		n = 0
	}
	if n >= len(m.ctrl.source) {
		n = len(m.ctrl.source) - 1
	}
	m.currentSentence = n
	m.ctrl.getSentenceData(uint(n))

	m.sentenceList.SetContent(m.renderSentenceList())
}

func (m *mainViewModel) selectWord(w_n int) {
	var data = (*m.ctrl.sentenceData[m.currentSentence])
	var words = data.opts.WordOrigins

	if w_n < 0 {
		w_n = len(*words) - 1
	}
	if w_n >= len(*words) {
		w_n = 0
	}
	m.currentWord = w_n
}

func (m *mainViewModel) selectNextWord() {
	m.selectWord(m.currentWord + 1)
}

func (m *mainViewModel) selectPrevWord() {
	m.selectWord(m.currentWord - 1)
}

func (m *mainViewModel) setTermDimensions(w int, h int) {
	m.width = w
	m.height = h
}

func (m *mainViewModel) selectPhoneme(n int) {
	_, _, tags := m.getPhonemeOptionsForWord(m.currentWord)
	phonemes := slices.Sorted(maps.Keys(tags))
	if n < 0 {
		n = len(phonemes) - 1
	}
	if n >= len(phonemes) {
		n = 0
	}
	m.selectedPhoneme = n
}

func (m *mainViewModel) selectNextPhoneme() {
	m.selectPhoneme(m.selectedPhoneme + 1)
}

func (m *mainViewModel) selectPrevPhoneme() {
	m.selectPhoneme(m.selectedPhoneme - 1)
}

func (m *mainViewModel) selectDefaultPhoneme() {
	_, def, tags := m.getPhonemeOptionsForWord(m.currentWord)
	phonemes := slices.Sorted(maps.Keys(tags))

	for idx, p := range phonemes {
		if p == def {
			m.selectPhoneme(idx)
			return
		}
	}
	m.selectPhoneme(0)
}

func (m *mainViewModel) getPhonemeOptionsForWord(word_n int) (string, string, map[string][]string) {
	s_data := m.ctrl.sentenceData[m.currentSentence]
	if s_data == nil {
		return "", "", nil
	}
	origin := (*s_data.opts.WordOrigins)[word_n]
	selected := s_data.selectedPhonemes[word_n]
	tags := (*s_data.opts.Tags)[word_n]
	return origin, selected[1], tags
}

func (m *mainViewModel) togglePhonemeView() {
	m.phonemeView = !m.phonemeView
}

func (m mainViewModel) renderSentenceList() string {

	l := list.New().
		Enumerator(list.Arabic).
		ItemStyleFunc(func(_ list.Items, i int) lipgloss.Style {
			st := sentenceListBaseStyle
			if m.currentSentence == i {
				return st.Foreground(sentenceListHighlightColor)
			}
			return st.Foreground(sentenceListDimColor)
		}).
		EnumeratorStyleFunc(func(_ list.Items, i int) lipgloss.Style {
			if m.currentSentence == i {
				return lipgloss.NewStyle().Foreground(sentenceListHighlightColor)
			}
			return lipgloss.NewStyle().Foreground(sentenceListDimColor)
		})

	for _, d := range m.ctrl.source {
		l.Item(d)
	}

	return lipgloss.Sprint("\n", l, "\n")
}

func (m *mainViewModel) generateSentenceTranscription(use_phonemes bool) string {
	var builder strings.Builder
	m.selectSentence(m.currentSentence)
	sentence_data := m.ctrl.sentenceData[m.currentSentence]
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
		if idx == m.currentWord {
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

func (m mainViewModel) phonemePanelView() string {
	word, selected, tags := m.getPhonemeOptionsForWord(m.currentWord)
	var builder strings.Builder

	builder.WriteString(lipgloss.Sprintf(
		"%s / %s\n\n",
		phonemeViewHeaderWordStyle.Render(word),
		phonemeViewHeaderPhonemeStyle.Render(selected),
	))

	keys := slices.Sorted(maps.Keys(tags))

	for idx, phoneme := range keys {
		if idx == m.selectedPhoneme {
			builder.WriteString(lipgloss.Sprintf("%d) %s : \n", idx+1, phonemeViewSelectedPhoneme.Render(phoneme)))
		} else {
			builder.WriteString(lipgloss.Sprintf("%d) %s : \n", idx+1, phoneme))
		}
		// phonemes without tags are implied to be inferred
		if len(tags[phoneme]) == 0 {
			builder.WriteString(lipgloss.Sprint("  - INFERRED \n"))
		} else {
			for _, tag := range tags[phoneme] {
				builder.WriteString(lipgloss.Sprintf("  - %s \n", tag))
			}
		}
	}
	return phonemeViewBaseStyle.Render(builder.String())
}

func (m mainViewModel) sentencePanelView() string {

	output := sentencePanelStyle.Width(m.sentenceList.Width()).Render(
		m.generateSentenceTranscription(false),
		"\n\n",
		m.generateSentenceTranscription(true),
	)
	return lipgloss.Sprint(output)
}

func (m mainViewModel) headerView() string {
	title := titleStyle.Render("Narrative Debugger v0.1")
	line := strings.Repeat("─", max(0, m.sentenceList.Width()-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}
