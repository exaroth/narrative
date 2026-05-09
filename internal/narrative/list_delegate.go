package narrative

import (
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/exaroth/narrative/internal/common"
)

const (
	ellipsis = "…"
)

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
	var id string

	if i, ok := m.SelectedItem().(*common.TextSource); ok {
		id = i.Id
	} else {
		return nil
	}
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		k := msg.String()
		switch k {
		case "enter":
			cmds = append(cmds, SelectSource(id, -1))
		case "space":
			cmds = append(cmds, TogglePlayback())
		case "d":
			cmds = append(cmds, ShowPrompt("Delete source?", DeleteSource(id), nil))
		}
	}
	return tea.Batch(cmds...)
}

func (d SourceListDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	var (
		title, author, stype string
		title_s, author_s    lipgloss.Style
		matchedRunes         []int
		s                    = &d.Styles
		playing              bool
	)

	if m.Width() <= 0 {
		return
	}

	// fmt.Println(m.Width())
	textwidth := m.Width() - s.NormalTitle.GetPaddingLeft() - s.NormalTitle.GetPaddingRight()

	if i, ok := item.(*common.TextSource); ok {
		title = i.Title
		author = i.Author
		stype = i.SourceType.String()
		playing = i.Playing
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
	// author = "J.R.R Tokien"

	var base lipgloss.Style
	switch {
	case emptyFilter:
		title_s = s.DimmedTitle
		author_s = s.DimmedDesc
		base = s.Base
	case isSelected && m.FilterState() != list.Filtering:
		if isFiltered {
			unmatched := s.SelectedTitle.Inline(true)
			matched := unmatched.Inherit(s.FilterMatch)
			title = lipgloss.StyleRunes(title, matchedRunes, matched, unmatched)
		}
		if playing {
			base = s.BasePlayingSelected
		} else {
			base = s.BaseSelected
		}
		title_s = s.SelectedTitle
		author_s = s.SelectedDesc
	default:
		if isFiltered {
			unmatched := s.NormalTitle.Inline(true)
			matched := unmatched.Inherit(s.FilterMatch)
			title = lipgloss.StyleRunes(title, matchedRunes, matched, unmatched)
		}
		if playing {
			base = s.BasePlaying
		} else {
			base = s.Base
		}
		title_s = s.NormalTitle
		author_s = s.NormalDesc
	}

	base = base.Width(m.Width())

	stype = s.SourceTypeStyle.Render(strings.ToUpper(stype))
	// get width of the title - length of the stype

	title_row_l := textwidth - lipgloss.Width(stype) - 4 // padding
	title = ansi.Truncate(title, title_row_l, ellipsis)
	author = ansi.Truncate(author, textwidth, ellipsis)

	title = title_s.Width(title_row_l).Render(title)
	author = author_s.Render(author)
	title_row := lipgloss.JoinHorizontal(lipgloss.Top,
		title,
		stype,
	)
	result := base.Render(lipgloss.Sprintf("%s\n%s", title_row, author))
	fmt.Fprint(w, result)
}
