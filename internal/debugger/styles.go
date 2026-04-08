package debugger

import "charm.land/lipgloss/v2"

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
