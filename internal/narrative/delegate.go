package narrative

import (
	"fmt"
	"image/color"
	"io"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

const (
	ellipsis = "…"
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

var col = ColorPalette{
	Background: lipgloss.Color("#040C06"),
	Foreground: lipgloss.Color("#EEFFCC"),
	Color1:     lipgloss.Color("#BEDC7F"),
	Color2:     lipgloss.Color("#4D8061"),
	Color3:     lipgloss.Color("#305D42"),
	Color4:     lipgloss.Color("#112318"),
	ColorDim:   lipgloss.Color("#A49FA5"),
}

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
	FilterMatch lipgloss.Style
}

func NewListStyles(isDark bool) (s ListStyles) {
	// TODO
	// lightDark := lipgloss.LightDark(isDark)

	s.Base = lipgloss.NewStyle().
		Background(col.Background).
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
		Foreground(col.Foreground)

	s.DimmedTitle = lipgloss.NewStyle().
		Foreground(col.ColorDim).
		Padding(0, 0, 0, 1)

	s.DimmedDesc = s.DimmedTitle.
		Foreground(col.ColorDim)

	s.FilterMatch = lipgloss.NewStyle().Underline(true)

	return s
}

type DefaultItem interface {
	list.Item
	Title() string
	Description() string
}

type SourceListDelegate struct {
	Styles        ListStyles
	UpdateFunc    func(tea.Msg, *list.Model) tea.Cmd
	ShortHelpFunc func() []key.Binding
	FullHelpFunc  func() [][]key.Binding
	height        int
	spacing       int
}

// Create new source list delegate.
func NewDelegate() SourceListDelegate {
	const defaultHeight = 2
	const defaultSpacing = 1
	return SourceListDelegate{
		Styles:  NewListStyles(true),
		height:  defaultHeight,
		spacing: defaultSpacing,
	}
}

// SetHeight sets delegate's preferred height.
func (d *SourceListDelegate) SetHeight(i int) {
	d.height = i
}

func (d SourceListDelegate) Height() int {
	return d.height
}

func (d *SourceListDelegate) SetSpacing(i int) {
	d.spacing = i
}

func (d SourceListDelegate) Spacing() int {
	return d.spacing
}

func (d SourceListDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	if d.UpdateFunc == nil {
		return nil
	}
	return d.UpdateFunc(msg, m)
}

// Render prints an item.
func (d SourceListDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	var (
		title, author, stype string
		matchedRunes         []int
		s                    = &d.Styles
	)

	if m.Width() <= 0 {
		return
	}

	// fmt.Println(m.Width())
	textwidth := m.Width() - s.NormalTitle.GetPaddingLeft() - s.NormalTitle.GetPaddingRight()

	if i, ok := item.(*TextSource); ok {
		title = ansi.Truncate(i.Title, textwidth, ellipsis)
		author = ansi.Truncate(i.Author, textwidth, ellipsis)
		stype = i.SourceType.String()
	} else {
		// should not ever happen
		panic("Invalid item type passed to delegate")
	}

	var (
		isSelected  = index == m.Index()
		emptyFilter = m.FilterState() == list.Filtering && m.FilterValue() == ""
		isFiltered  = m.FilterState() == list.Filtering || m.FilterState() == list.FilterApplied
	)

	// if isFiltered && index < len(m.filteredItems) {
	if isFiltered {
		matchedRunes = m.MatchesForItem(index)
	}

	var base lipgloss.Style
	if emptyFilter {
		title = s.DimmedTitle.Render(title)
		author = s.DimmedDesc.Render(author)
		base = s.Base
	} else if isSelected && m.FilterState() != list.Filtering {
		if isFiltered {
			unmatched := s.SelectedTitle.Inline(true)
			matched := unmatched.Inherit(s.FilterMatch)
			title = lipgloss.StyleRunes(title, matchedRunes, matched, unmatched)
		}
		base = s.BaseSelected
		title = s.SelectedTitle.Render(title)
		author = s.SelectedDesc.Render(author)
	} else {
		if isFiltered {
			unmatched := s.NormalTitle.Inline(true)
			matched := unmatched.Inherit(s.FilterMatch)
			title = lipgloss.StyleRunes(title, matchedRunes, matched, unmatched)
		}
		base = s.Base
		title = s.NormalTitle.Render(title)
		author = s.NormalDesc.Render(author)
	}

	base = base.Width(m.Width())
	result := base.Render(lipgloss.Sprintf("%s\n%s%s", title, author, stype))
	fmt.Fprint(w, result)
}

// ShortHelp returns the delegate's short help.
func (d SourceListDelegate) ShortHelp() []key.Binding {
	if d.ShortHelpFunc != nil {
		return d.ShortHelpFunc()
	}
	return nil
}

// FullHelp returns the delegate's full help.
func (d SourceListDelegate) FullHelp() [][]key.Binding {
	if d.FullHelpFunc != nil {
		return d.FullHelpFunc()
	}
	return nil
}
