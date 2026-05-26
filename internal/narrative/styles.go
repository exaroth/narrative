package narrative

import (
	"image/color"

	"charm.land/lipgloss/v2"
	"github.com/exaroth/narrative/pkg/reader"
)

type ColorPalette struct {
	Background    color.Color
	Foreground    color.Color
	Color1        color.Color
	Color2        color.Color
	Color3        color.Color
	Color4        color.Color
	ColorError    color.Color
	ColorDim      color.Color
	ColorBookmark color.Color
	ColorChapter  color.Color
}

var col = ColorPalette{
	Background:    lipgloss.Color("#040C06"),
	Foreground:    lipgloss.Color("#EEFFCC"),
	Color1:        lipgloss.Color("#BEDC7F"),
	Color2:        lipgloss.Color("#4D8061"),
	Color3:        lipgloss.Color("#305D42"),
	Color4:        lipgloss.Color("#112318"),
	ColorError:    lipgloss.Color("#F25D94"),
	ColorDim:      lipgloss.Color("#5C5C5C"),
	ColorBookmark: lipgloss.Color("#F25D94"),
	ColorChapter:  lipgloss.Color("#0057E3"),
}

// Main list style
// ===============
var (
	docStyle = lipgloss.NewStyle().Margin(1, 2)
	errStyle = lipgloss.NewStyle().Foreground(col.ColorError)
)

var sourceTypeStyleMap = map[reader.SourceType]color.Color{
	reader.SourceTypeText:        col.ColorDim,
	reader.SourceTypeEpub:        lipgloss.Color("#FFC400"),
	reader.SourceTypeMobi:        lipgloss.Color("#11A30F"),
	reader.SourceTypeAzw3:        lipgloss.Color("#FF8400"),
	reader.SourceTypeHTML:        lipgloss.Color("#0099FF"),
	reader.SourceTypeMarkdown:    lipgloss.Color("#0FA374"),
	reader.SourceTypeUnsupported: col.ColorError,
}

// Modal styles.
// ===============
var (
	modalStyle = lipgloss.NewStyle().
			Padding(1, 1).
			Border(lipgloss.NormalBorder(), true, true, true, true).
			Align(lipgloss.Center)

	modalInfoStyle = lipgloss.NewStyle().
			Inherit(modalStyle).
			Foreground(col.Foreground).
			BorderForeground(col.Foreground)

	modalErrorStyle = lipgloss.NewStyle().
			Inherit(modalStyle).
			Foreground(col.ColorError).
			BorderForeground(col.ColorError)
)

// Transcript styles
// =================

var (
	transcriptStyle = lipgloss.NewStyle().
			Padding(1, 1).Margin(0, 0, 1, 0)
	transcriptSeparatorStyle = lipgloss.NewStyle().Foreground(col.Color2)
)

// Style for % read text.
// ======================
var percReadStyle = lipgloss.NewStyle().
	Foreground(col.Color1)

// Style for app title.
// ====================
var TitleStyle = lipgloss.NewStyle().
	Foreground(col.Foreground).
	Background(col.Color3).
	Padding(0, 1)

// Contains styles for primary sources list.
type ListStyles struct {
	Base                lipgloss.Style
	BasePlaying         lipgloss.Style
	BaseSelected        lipgloss.Style
	BasePlayingSelected lipgloss.Style

	// The Normal state.
	NormalTitle lipgloss.Style
	NormalDesc  lipgloss.Style

	// The selected item state.
	SelectedTitle lipgloss.Style
	SelectedDesc  lipgloss.Style

	// The dimmed state, for when the filter input is initially activated.
	DimmedTitle lipgloss.Style
	DimmedDesc  lipgloss.Style

	// Characters matching the current filter, if any.
	FilterMatch     lipgloss.Style
	SourceTypeStyle lipgloss.Style
}

// Return styles for primary sources list.
func NewListStyles(isDark bool) (s ListStyles) {
	// TODO
	// lightDark := lipgloss.LightDark(isDark)

	s.Base = lipgloss.NewStyle().
		// Background(col.Background).
		Border(lipgloss.NormalBorder(), true, true, true, true).
		BorderForeground(col.Color3)

	s.BaseSelected = s.Base.
		BorderForeground(col.Color1)

	s.BasePlaying = lipgloss.NewStyle().
		BorderForeground(col.Color2).
		Border(lipgloss.ThickBorder(), true, true, true, true)

	s.BasePlayingSelected = s.BasePlaying.
		BorderForeground(col.Color1)

	s.NormalTitle = lipgloss.NewStyle().
		Padding(0, 0, 0, 1)

	s.NormalDesc = s.NormalTitle.
		Foreground(col.Color3)

	s.SelectedTitle = lipgloss.NewStyle().
		Foreground(col.Foreground).
		Padding(0, 0, 0, 1)

	s.SelectedDesc = s.SelectedTitle.
		Foreground(col.Color2)

	s.DimmedTitle = lipgloss.NewStyle().
		Foreground(col.ColorDim).
		Padding(0, 0, 0, 1)

	s.DimmedDesc = s.DimmedTitle.
		Foreground(col.ColorDim)

	s.SourceTypeStyle = lipgloss.NewStyle().Foreground(col.ColorDim).Bold(true)

	s.FilterMatch = lipgloss.NewStyle().Underline(true)
	return s
}

// Main status bar styles
// ======================

// Get color for the status part of the status bar
func GetStatusBarStatusColor(status playbackStatus) color.Color {
	var c color.Color
	switch status {
	case playbackBuffering:
		c = lipgloss.Color("#0099DB")
	case playbackPaused:
		c = lipgloss.Color("#FEAE34")
	case playbackPlaying:
		c = lipgloss.Color("#E43B44")
	case playbackIdle:
		c = lipgloss.Color("#8b9bb4")
	}
	return c
}

var (
	statusBarStyle = lipgloss.NewStyle()

	statusBarStatusStyle       = lipgloss.NewStyle().Width(9).Foreground(col.Color4).Bold(true)
	statusBarStatusText        = lipgloss.NewStyle().Foreground(col.ColorDim).Background(col.Color4)
	statusBarStatusTextCommand = lipgloss.NewStyle().Inherit(statusBarStatusText).Bold(true)
	statusBarPromptTextStyle   = lipgloss.NewStyle().Foreground(col.Color2).Background(col.Color4)
	statusBarPromptYStyle      = lipgloss.NewStyle().Foreground(col.Color1).Background(col.Color4)
	statusBarPromptAccStyle    = lipgloss.NewStyle().Foreground(col.ColorError).Background(col.Color4)
)

// Progress bar styles
// ==================a

var progressBarFilledStyle = lipgloss.NewStyle().Foreground(col.Color2)
var progressBarEmptyStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#606060"))
var progressBarBookmarkStyle = lipgloss.NewStyle().Foreground(col.ColorBookmark)
var progressBarChapterStyle = lipgloss.NewStyle().Foreground(col.ColorChapter)

// Help panel
// ==========

var (
	helpPanelLogoStyle    = lipgloss.NewStyle().Foreground(col.ColorDim)
	helpPanelTextStyle    = lipgloss.NewStyle().Foreground(col.Color2)
	helpPanelCommandStyle = lipgloss.NewStyle().Inherit(helpPanelTextStyle).Foreground(col.Color1).Bold(true)
	helpPanelListStyle    = lipgloss.NewStyle().MarginLeft(1).PaddingLeft(1).Foreground(col.Color2)
)

// Welcome screen
// ===============

var welcomeScreenLogoStyle = lipgloss.NewStyle().Foreground(col.Color1)
var welcomeScreenMessageStyle = lipgloss.NewStyle().Margin(1, 0)
var welcomeScreenWarningStyleStyle = lipgloss.NewStyle().Foreground(col.ColorError)
var welcomeScreenModelDescriptionStyle = lipgloss.NewStyle().Foreground(col.Color3)
var welcomeScreenModelNameStyle = lipgloss.NewStyle().Foreground(col.Color2)
var welcomeScreenModelNameSelectedStyle = lipgloss.NewStyle().
	Inherit(welcomeScreenModelNameStyle).
	Foreground(col.Color1)

// Info panel
// ==========
var infoPanelStyle = lipgloss.NewStyle().Background(col.Background).Padding(2)
