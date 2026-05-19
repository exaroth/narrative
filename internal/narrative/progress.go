// Package progress provides a simple progress bar for Bubble Tea applications.
package narrative

import (
	"image/color"
	"maps"
	"math"
	"slices"
	"strings"
	"sync/atomic"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/harmonica"
)

type ProgressMarkType int

const (
	ProgressMarkTypeBookmark ProgressMarkType = iota
	ProgressMarkTypeChapter
)

type ProgressMark struct {
	perc float64
	t    ProgressMarkType
}

func NewProgressBookmark(perc float64) *ProgressMark {
	return &ProgressMark{
		perc: perc,
		t:    ProgressMarkTypeBookmark,
	}
}

func NewProgressChapter(perc float64) *ProgressMark {
	return &ProgressMark{
		perc: perc,
		t:    ProgressMarkTypeChapter,
	}
}

type ProgressMarks []*ProgressMark

// Return map of marks where keys are the integer offset
// based on the total weight of the bar.
func (c ProgressMarks) AsWidthMap(width int) map[int]*ProgressMark {
	result := make(map[int]*ProgressMark)
	var pm int
	for _, m := range c {
		pm = int(math.Round((float64(width) * m.perc)))
		result[pm] = m
	}
	return result
}

// This is generator function for retrieving progress marks used during rendering
// of the progress bar.
func updateMarkGen(marks []*ProgressMark, width int) func(int) (int, ProgressMarkType) {
	mark := marks[0]
	var mark_i int
	p_m := int(math.Round((float64(width) * mark.perc)))

	return func(offset int) (int, ProgressMarkType) {
		var n_p_m int
		for {
			if mark_i+1 < len(marks) {
				mark_i += 1
				mark = marks[mark_i]
				n_p_m = int(math.Round((float64(width) * mark.perc))) - offset
				if n_p_m == p_m {
					continue
				} else {
					p_m = n_p_m
					break
				}
			} else {
				p_m = -1
				break
			}
		}
		return p_m, mark.t
	}

}

// Internal ID management. Used during animating to assure that frame messages
// can only be received by progress components that sent them.
var lastID int64

func nextID() int {
	return int(atomic.AddInt64(&lastID, 1))
}

const (
	DefaultFullCharHalfBlock = '▌'
	DefaultEmptyCharBlock    = '░'

	fps              = 60
	defaultWidth     = 40
	defaultFrequency = 18.0
	defaultDamping   = 1.0
)

type Option func(*ProgressModel)

// FrameMsg indicates that an animation step should occur.
type FrameMsg struct {
	id  int
	tag int
}

// Customized bubbles.progress progress bar with support for marking
// bookmarks.
type ProgressModel struct {
	// An identifier to keep us from receiving messages intended for other
	// progress bars.
	id int

	// An identifier to keep us from receiving frame messages too quickly.
	tag int

	// Total width of the progress bar, including percentage, if set.
	width int

	// "Filled" sections of the progress bar.
	Full      rune
	FullColor color.Color

	// "Empty" sections of the progress bar.
	Empty      rune
	EmptyColor color.Color

	// Members for animated transitions.
	spring        harmonica.Spring
	percentShown  float64 // percent currently displaying
	targetPercent float64 // percent to which we're animating
	velocity      float64

	chapters  ProgressMarks // Chapters of the source, if available
	bookmarks ProgressMarks // Bookmark positions
}

// New returns a model with default values.
func NewProgress() ProgressModel {
	m := ProgressModel{
		id:        nextID(),
		width:     defaultWidth,
		Full:      DefaultFullCharHalfBlock,
		Empty:     DefaultEmptyCharBlock,
		chapters:  ProgressMarks{},
		bookmarks: ProgressMarks{},
	}

	return m
}

func (m ProgressModel) Init() tea.Cmd {
	return nil
}

func (m ProgressModel) Update(msg tea.Msg) (ProgressModel, tea.Cmd) {
	switch msg := msg.(type) {
	case FrameMsg:
		if msg.id != m.id || msg.tag != m.tag {
			return m, nil
		}

		if !m.IsAnimating() {
			return m, nil
		}

		m.percentShown, m.velocity = m.spring.Update(m.percentShown, m.velocity, m.targetPercent)
		return m, m.nextFrame()
	case SetBookmarksCmd:
		bs := []*ProgressMark{}
		for _, b := range msg {
			bs = append(bs, NewProgressBookmark(b))
		}
		m.bookmarks = bs
		return m, nil
	case SetChaptersCmd:
		bs := []*ProgressMark{}
		for _, b := range msg {
			bs = append(bs, NewProgressChapter(b))
		}
		m.chapters = bs
		return m, nil
	default:
		return m, nil
	}
}

func (m ProgressModel) Percent() float64 {
	return m.targetPercent
}

func (m *ProgressModel) SetPercent(p float64) tea.Cmd {
	m.targetPercent = math.Max(0, math.Min(1, p))
	m.tag++
	return m.nextFrame()
}

func (m *ProgressModel) SetMarks(bmarks []float64, chapters []float64) tea.Cmd {
	b := []*ProgressMark{}
	c := []*ProgressMark{}
	for _, bmark := range bmarks {
		b = append(b, NewProgressBookmark(bmark))
	}
	for _, ch := range chapters {
		c = append(c, NewProgressChapter(ch))
	}
	m.chapters = c
	m.bookmarks = b
	m.tag++
	return m.nextFrame()
}

func (m *ProgressModel) AddBookmark(mark float64) tea.Cmd {
	for _, b := range m.bookmarks {
		if b.perc == mark {
			return nil
		}
	}
	m.bookmarks = append(m.bookmarks, NewProgressBookmark(mark))
	m.tag++
	return m.nextFrame()
}

// View renders an animated progress bar in its current state. To render
// a static progress bar based on your own calculations use ViewAs instead.
func (m ProgressModel) View() string {
	return m.ViewAs(m.percentShown)
}

// ViewAs renders the progress bar with a given percentage.
func (m ProgressModel) ViewAs(percent float64) string {
	b := strings.Builder{}
	m.barView(&b, percent, 0)
	return b.String()
}

// SetWidth sets the width of the progress bar.
func (m *ProgressModel) SetWidth(w int) {
	m.width = w
}

// Width returns the width of the progress bar.
func (m ProgressModel) Width() int {
	return m.width
}

func (m *ProgressModel) nextFrame() tea.Cmd {
	return tea.Tick(time.Second/time.Duration(fps), func(time.Time) tea.Msg {
		return FrameMsg{id: m.id, tag: m.tag}
	})
}

// Retrieve ordered marks - chapters and bookmarks for bar. Pass
// total width of progress bar in argument. Marks width same position
// will be coalesced.
func (m *ProgressModel) getMarksForBar(tw int) []*ProgressMark {
	mark_map := m.chapters.AsWidthMap(tw)
	maps.Copy(mark_map, m.bookmarks.AsWidthMap(tw))
	marks := slices.SortedFunc(maps.Values(mark_map), func(a, b *ProgressMark) int {
		if a.perc < b.perc {
			return -1
		}
		if a.perc > b.perc {
			return 1
		}
		return 0
	})
	return marks
}

// Render progress bar including bookmarks/chapters.
func (m ProgressModel) barView(b *strings.Builder, percent float64, textWidth int) {
	var (
		tw = max(0, m.width-textWidth)                // total width
		fw = int(math.Round((float64(tw) * percent))) // filled width
	)

	fw = max(0, min(tw, fw))
	marks := m.getMarksForBar(tw)

	if len(marks) == 0 {
		b.WriteString(progressBarFilledStyle.
			Render(strings.Repeat(string(m.Full), fw)))

		n := max(0, tw-fw)
		b.WriteString(progressBarEmptyStyle.
			Render(strings.Repeat(string(m.Empty), n)))
	} else {
		var barB strings.Builder
		temp_s := []rune{}
		update_mark := updateMarkGen(marks, tw)
		p_m, mark_t := update_mark(0)
		for fw_i := 0; fw_i < fw; fw_i++ {
			if fw_i == p_m {
				barB.WriteString(progressBarFilledStyle.Render(string(temp_s)))
				if mark_t == ProgressMarkTypeBookmark {
					barB.WriteString(progressBarBookmarkStyle.Render(string(m.Full)))
				} else {
					barB.WriteString(progressBarChapterStyle.Render(string(m.Full)))
				}
				temp_s = []rune{}
				p_m, mark_t = update_mark(0)
			} else {
				temp_s = append(temp_s, m.Full)
			}
		}
		if len(temp_s) > 0 {
			barB.WriteString(progressBarFilledStyle.Render(string(temp_s)))
			temp_s = []rune{}
		}

		empty_n := max(0, tw-fw)
		p_m -= fw

		for e_i := 0; e_i < empty_n; e_i++ {
			if e_i == p_m {
				barB.WriteString(progressBarEmptyStyle.Render(string(temp_s)))
				if mark_t == ProgressMarkTypeBookmark {
					barB.WriteString(progressBarBookmarkStyle.Render(string(m.Empty)))
				} else {
					barB.WriteString(progressBarChapterStyle.Render(string(m.Empty)))
				}
				temp_s = []rune{}
				p_m, mark_t = update_mark(fw)
			} else {
				temp_s = append(temp_s, m.Empty)
			}
		}
		if len(temp_s) > 0 {
			barB.WriteString(progressBarEmptyStyle.Render(string(temp_s)))
		}
		b.WriteString(barB.String())
	}
}

// IsAnimating returns false if the progress bar reached equilibrium and is no
// longer animating.
func (m *ProgressModel) IsAnimating() bool {
	dist := math.Abs(m.percentShown - m.targetPercent)
	return !(dist < 0.001 && m.velocity < 0.01)
}
