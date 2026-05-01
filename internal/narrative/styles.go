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
	ColorDim   color.Color
}

// Predefined colors.
var (
	colorError = lipgloss.Color("#F25D94")
)

var col = ColorPalette{
	Background: lipgloss.Color("#040C06"),
	Foreground: lipgloss.Color("#EEFFCC"),
	Color1:     lipgloss.Color("#BEDC7F"),
	Color2:     lipgloss.Color("#4D8061"),
	Color3:     lipgloss.Color("#305D42"),
	Color4:     lipgloss.Color("#112318"),
	ColorDim:   lipgloss.Color("#A49FA5"),
}

// Main list style
var docStyle = lipgloss.NewStyle().Margin(1, 2)

// Modal styles.
var (
	modalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFF7DB")).
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

// Style for app title.
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
		Foreground(col.ColorDim).
		Padding(0, 0, 0, 1)

	s.DimmedDesc = s.DimmedTitle.
		Foreground(col.ColorDim)

	s.SourceTypeStyle = lipgloss.NewStyle().Foreground(col.ColorDim)

	s.FilterMatch = lipgloss.NewStyle().Underline(true)
	return s
}
