package debugger

import (
	"maps"
	"slices"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type ClosePhonemePanelCmd struct{}

type UpdatePhonemeCmd struct {
	word, phoneme string
	wordNum       int
	sentenceNum   int
	reopen        bool
}

type phonemePanel struct {
	word            string
	wordNum         int
	selectedPhoneme string
	available       map[string][]string
	phonemesSorted  []string
	currentPhoneme  int
	sentenceNum     int
	input           textinput.Model
	showInput       bool
	inputQuitting   bool

	sentenceData *[]map[string][]string
}

func OpenPhonemePanel(ctrl *Debugger, sentence_n, word_n int) *phonemePanel {

	s_data := ctrl.sentenceData[sentence_n]
	origin := (*s_data.opts.WordOrigins)[word_n]
	selected := s_data.selectedPhonemes[word_n]
	tags := (*s_data.opts.Tags)[word_n]

	p := &phonemePanel{
		word:            origin,
		wordNum:         word_n,
		selectedPhoneme: selected[1],
		available:       tags,
		sentenceNum:     sentence_n,
		phonemesSorted:  slices.Sorted(maps.Keys(tags)),
		sentenceData:    s_data.opts.Tags,
	}

	p.selectDefaultPhoneme()
	return p
}

func (p phonemePanel) Init() tea.Cmd {
	return nil
}

func (p phonemePanel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if p.showInput {
			if k := msg.String(); k == "enter" {
				cmds = append(cmds, p.updatePhonemeInput())
				p.closeInput()
			}
			if k := msg.String(); k == "esc" || k == "ctrl-c" {
				p.closeInput()
			}
		} else {
			if k := msg.String(); k == "q" || k == "esc" {
				cmds = append(cmds, p.closePhonemePanel())
			}
			if k := msg.String(); k == "j" {
				p.selectPrevPhoneme()
			}
			if k := msg.String(); k == "k" || k == "tab" {
				p.selectNextPhoneme()
			}
			if k := msg.String(); k == "enter" {
				cmds = append(cmds, p.updatePhonemeSelected())
			}
			if k := msg.String(); k == "a" {
				p.startInput()
			}

		}
	}

	p.input, cmd = p.input.Update(msg)
	cmds = append(cmds, cmd)

	return p, tea.Batch(cmds...)
}

func (p phonemePanel) View() tea.View {
	var v tea.View
	mainPanel := p.phonemePanelView()
	if !p.showInput {
		v.SetContent(mainPanel)
		// v.SetContent(pp.Sprintf("\n%s", p.sentenceData))
		return v
	}

	var c *tea.Cursor
	if !p.input.VirtualCursor() {
		c = p.input.Cursor()
		c.Y += lipgloss.Height(mainPanel)
	}

	str := lipgloss.JoinVertical(lipgloss.Top, mainPanel, p.input.View())
	if p.inputQuitting {
		str += "\n"
	}

	v.SetContent(str)
	v.Cursor = c
	return v

}

func (p *phonemePanel) startInput() {

	ti := textinput.New()
	ti.Placeholder = p.selectedPhoneme
	ti.SetValue(p.selectedPhoneme)
	ti.SetVirtualCursor(false)
	ti.Focus()
	ti.CharLimit = 156
	ti.SetWidth(200)

	p.input = ti
	p.showInput = true

}

func (p *phonemePanel) closeInput() {
	p.inputQuitting = true
	p.showInput = false
}

func (p phonemePanel) closePhonemePanel() tea.Cmd {
	return func() tea.Msg {
		return ClosePhonemePanelCmd{}
	}
}

func (p *phonemePanel) selectPhoneme(n int) {
	if n < 0 {
		n = len(p.phonemesSorted) - 1
	}
	if n >= len(p.phonemesSorted) {
		n = 0
	}
	p.currentPhoneme = n
}

func (p *phonemePanel) selectNextPhoneme() {
	p.selectPhoneme(p.currentPhoneme + 1)
}

func (p *phonemePanel) selectPrevPhoneme() {
	p.selectPhoneme(p.currentPhoneme - 1)
}

func (p *phonemePanel) updatePhonemeInput() tea.Cmd {
	return p.updatePhoneme(
		p.word,
		p.input.Value(),
		true,
	)
}

func (p *phonemePanel) updatePhonemeSelected() tea.Cmd {
	return p.updatePhoneme(
		p.word,
		p.phonemesSorted[p.currentPhoneme],
		false,
	)
}

func (p *phonemePanel) updatePhoneme(word, phoneme string, reopen bool) tea.Cmd {

	return func() tea.Msg {
		return UpdatePhonemeCmd{
			word:        word,
			wordNum:     p.wordNum,
			phoneme:     phoneme,
			sentenceNum: p.sentenceNum,
			reopen:      reopen,
		}
	}
}

func (p *phonemePanel) selectDefaultPhoneme() {

	for idx, ph := range p.phonemesSorted {
		if ph == p.selectedPhoneme {
			p.selectPhoneme(idx)
			return
		}
	}
	p.selectPhoneme(0)
}

func (p phonemePanel) phonemePanelView() string {
	var builder strings.Builder

	builder.WriteString(lipgloss.Sprintf(
		"%s / %s\n\n",
		phonemeViewHeaderWordStyle.Render(p.word),
		phonemeViewHeaderPhonemeStyle.Render(p.selectedPhoneme),
	))

	for idx, phoneme := range p.phonemesSorted {
		if idx == p.currentPhoneme {
			builder.WriteString(lipgloss.Sprintf("%d) %s : \n", idx+1, phonemeViewSelectedPhoneme.Render(phoneme)))
		} else {
			builder.WriteString(lipgloss.Sprintf("%d) %s : \n", idx+1, phoneme))
		}
		// phonemes without tags are implied to be inferred
		if len(p.available[phoneme]) == 0 {
			builder.WriteString(lipgloss.Sprint("  - INFERRED \n"))
		} else {
			for _, tag := range p.available[phoneme] {
				builder.WriteString(lipgloss.Sprintf("  - %s \n", tag))
			}
		}
	}
	return phonemeViewBaseStyle.Render(builder.String())
}
