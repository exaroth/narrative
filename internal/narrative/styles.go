package narrative

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

type ColorPalette struct {
	Background color.Color
	Foreground color.Color
	Color1     color.Color
	Color2     color.Color
	Color3     color.Color
	Color4     color.Color
}

// Predefined colors.
var (
	colorError = lipgloss.Color("#F25D94")
	colorDim   = lipgloss.Color("#A49FA5")
)

var col = ColorPalette{
	Background: lipgloss.Color("#040C06"),
	Foreground: lipgloss.Color("#EEFFCC"),
	Color1:     lipgloss.Color("#BEDC7F"),
	Color2:     lipgloss.Color("#4D8061"),
	Color3:     lipgloss.Color("#305D42"),
	Color4:     lipgloss.Color("#112318"),
}

// Main list style
// ===============
var docStyle = lipgloss.NewStyle().Margin(1, 2)

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
			Foreground(colorError).
			BorderForeground(colorError)
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
	Base         lipgloss.Style
	BaseSelected lipgloss.Style

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
		Foreground(colorDim).
		Padding(0, 0, 0, 1)

	s.DimmedDesc = s.DimmedTitle.
		Foreground(colorDim)

	s.SourceTypeStyle = lipgloss.NewStyle().Foreground(colorDim)

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
	statusBarStyle = lipgloss.NewStyle().Background(col.Color4)

	statusBarStatusStyle = lipgloss.NewStyle().Width(9).Foreground(col.Color4)
	statusBarStatusText  = lipgloss.NewStyle().Foreground(colorDim)
)

// Progress bar styles
// ==================a

func ProgressBarColorFunc(total, current float64) color.Color {
	if total <= 0.25 {
		return col.Color1
	}
	if total <= 0.50 {
		return col.Color2
	}
	if total <= 0.50 {
		return col.Color3
	}
	return col.Color4
}
