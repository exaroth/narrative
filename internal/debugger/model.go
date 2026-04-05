package debugger

import (
	"fmt"
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
	sentencePanelWordStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("110"))
	sentencePanelPunctStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("white"))
	sentencePanelNumStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
	sentencePanelPhonemeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("222"))

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
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()
)

type mainViewModel struct {
	ready           bool
	currentSentence int
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
	m.sentenceList.SetContent(m.renderList())

}

func (m mainViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if k := msg.String(); k == "ctrl+c" || k == "q" || k == "esc" {
			return m, tea.Quit
		}
		if k := msg.String(); k == "j" {
			m.selectNextSentence()
		}
		if k := msg.String(); k == "k" {
			m.selectPrevSentence()
		}
	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		sentencePanelHeight := lipgloss.Height(m.sentencePanelView())
		// get footer height
		verticalMarginHeight := headerHeight + sentencePanelHeight

		if !m.ready {
			m.sentenceList = viewport.New(
				viewport.WithWidth(msg.Width),
				viewport.WithHeight(msg.Height-verticalMarginHeight),
			)
			m.sentenceList.KeyMap = GetSentenceListKeymap()
			m.sentenceList.YPosition = headerHeight
			m.sentenceList.SetContent(m.renderList())
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
	} else {
		v.SetContent(fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.sentenceList.View(), m.sentencePanelView()))
		m.ready = true
	}

	return v
}

func (m mainViewModel) renderList() string {

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

func (m mainViewModel) generateSentenceTranscription(use_phonemes bool) string {
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
		func() {
			for _, e := range sentence_data.selectedPhonemes {
				words = append(words, e[1])
			}
		}()
	} else {
		style = sentencePanelWordStyle
		words = *sentence_data.opts.WordOrigins
	}
	punctuation := sentence_data.punctuation.AsArr(len(words))

	var w, pre_punct, post_punct, num string
	for idx, word := range words {
		num = lipgloss.Sprint(sentencePanelNumStyle.Render(fmt.Sprintf("(%d)", idx+1)))
		w = lipgloss.Sprint(style.Render(word))
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
		return fmt.Sprintf("Phonemes: %s", builder.String())
	}
	return fmt.Sprintf(" Input: %s", builder.String())
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
