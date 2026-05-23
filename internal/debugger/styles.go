package debugger

import (
	"os"

	"charm.land/lipgloss/v2"
)

var (
	hasDarkBG = lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
	lightDark = lipgloss.LightDark(hasDarkBG)
)

// sentence list styles
var (
	sentenceListDimColor       = lipgloss.Color("250")
	sentenceListHighlightColor = lipgloss.Color("228")
	sentenceListMarkedColor    = lipgloss.Color("#186600")
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
			Height(20).
			BorderStyle(b).
			BorderForeground(lipgloss.Color("237")).
			PaddingLeft(2).PaddingTop(1).PaddingRight(1)
	}()
)

// Phoneme view styles

var (
	phonemeViewBaseStyle          = lipgloss.NewStyle().Padding(1)
	phonemeViewHeaderWordStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFF"))
	phonemeViewHeaderPhonemeStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFF"))
	phonemeViewSelectedPhoneme    = lipgloss.NewStyle().Foreground(lipgloss.Color("#F00"))
)

// status bar

var (
	statusBarStyle = lipgloss.NewStyle().
			Foreground(lightDark(lipgloss.Color("#343433"), lipgloss.Color("#C1C6B2"))).
			Background(lightDark(lipgloss.Color("#D9DCCF"), lipgloss.Color("#353533")))

	statusStyle = lipgloss.NewStyle().
			Inherit(statusBarStyle).
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#A550DF")).
			Padding(0, 1).
			MarginRight(1)

	statusBarStatusText  = lipgloss.NewStyle().Inherit(statusBarStyle)
	statusBarCrumbsStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFDF5")).
				Padding(0, 1).Background(lipgloss.Color("#6124DF"))
	statusBarRContentsStyle = lipgloss.NewStyle().Inherit(statusBarStyle).
				Foreground(lipgloss.Color("#FF0000"))
)

// Help panel
var (
	helpPanelLogoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("250")).Margin(1, 0)
	helpPanelTextStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("228"))
	helpPanelCommandStyle = lipgloss.NewStyle().Inherit(helpPanelTextStyle).Foreground(lipgloss.Color("230")).Bold(true)
	helpPanelListStyle    = lipgloss.NewStyle().MarginLeft(1).PaddingLeft(1).Foreground(lipgloss.Color("228"))
)
