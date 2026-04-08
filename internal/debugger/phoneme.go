package debugger

import (
	"maps"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type ClosePhonemePanelCmd struct{}

type phonemePanel struct {
	word            string
	selectedPhoneme string
	available       map[string][]string
	phonemesSorted  []string
	currentPhoneme  int
	sentenceNum     int
	input           *input
}

func OpenPhonemePanel(ctrl *Debugger, sentence_n, word_n int) *phonemePanel {

	s_data := ctrl.sentenceData[sentence_n]
	origin := (*s_data.opts.WordOrigins)[word_n]
	selected := s_data.selectedPhonemes[word_n]
	tags := (*s_data.opts.Tags)[word_n]

	p := &phonemePanel{
		word:            origin,
		selectedPhoneme: selected[1],
		available:       tags,
		phonemesSorted:  slices.Sorted(maps.Keys(tags)),
	}

	p.selectDefaultPhoneme()
	return p
}

func (p phonemePanel) Init() tea.Cmd {
	return nil
}

func (p phonemePanel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if k := msg.String(); k == "q" || k == "esc" {
			// p.togglePhonemeView()
			cmds = append(cmds, p.closePhonemePanel())
		}
		if k := msg.String(); k == "j" {
			p.selectPrevPhoneme()
		}
		if k := msg.String(); k == "k" || k == "tab" {
			p.selectNextPhoneme()
		}
		if k := msg.String(); k == "enter" {
			// p.updateExtDict()
			// p.togglePhonemeView()
		}
	}
	return p, tea.Batch(cmds...)
}

func (p phonemePanel) View() tea.View {
	var v tea.View
	v.SetContent(p.phonemePanelView())
	return v
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
