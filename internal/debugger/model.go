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
	sentenceListBaseStyle      = lipgloss.NewStyle().MarginBottom(1).MarginLeft(1).Inline(true)
	sentenceListDimColor       = lipgloss.Color("250")
	sentenceListHighlightColor = lipgloss.Color("#EE6FF8")
)

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

func NewDebuggerModel(controller *Debugger) tea.Model {
	return &mainViewModel{
		ctrl: controller,
	}
}

func (m mainViewModel) Init() tea.Cmd {
	// return tick()
	return nil
}

func (m mainViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if k := msg.String(); k == "ctrl+c" || k == "q" || k == "esc" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		fmt.Println(headerHeight)
		// get footer height
		verticalMarginHeight := headerHeight
		if !m.ready {
			m.sentenceList = viewport.New(
				viewport.WithWidth(msg.Width),
				viewport.WithHeight(msg.Height-verticalMarginHeight),
			)
			m.sentenceList.SoftWrap = true
			m.sentenceList.SetContent(m.renderList())
			m.ready = true
		} else {
			m.sentenceList.SetWidth(msg.Width)
			m.sentenceList.SetHeight(msg.Height - verticalMarginHeight)
		}
	case tickMsg:
		fmt.Println("tick")
		// return m, tick()
	}
	return m, nil
}

func (m mainViewModel) View() tea.View {
	var v tea.View
	v.AltScreen = true
	if !m.ready {
		v.SetContent("\n  Initializing...")
	} else {
		v.SetContent(fmt.Sprintf("%s\n%s", m.headerView(), m.sentenceList.View()))
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

// func (m model) renderSentenceParts() string {

// }

// func (m model) generatePreProcParts() string {

// }

// func (m model) generatePhonemeParts() string {

// }

func (m mainViewModel) headerView() string {
	title := titleStyle.Render("Narrative Debugger v0.1")
	line := strings.Repeat("─", max(0, m.sentenceList.Width()-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}
