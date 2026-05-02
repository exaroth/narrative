// Package progress provides a simple progress bar for Bubble Tea applications.
package narrative

import (
	"image/color"
	"math"
	"strings"
	"sync/atomic"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/harmonica"
)

// Internal ID management. Used during animating to assure that frame messages
// can only be received by progress components that sent them.
var lastID int64

func nextID() int {
	return int(atomic.AddInt64(&lastID, 1))
}

const (
	DefaultFullCharHalfBlock = '▌'
	DefaultFullCharFullBlock = '█'
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

// Model stores values we'll use when rendering the progress bar.
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
	spring           harmonica.Spring
	springCustomized bool
	percentShown     float64 // percent currently displaying
	targetPercent    float64 // percent to which we're animating
	velocity         float64
}

// New returns a model with default values.
func NewProgress(col color.Color) ProgressModel {
	m := ProgressModel{
		id:         nextID(),
		width:      defaultWidth,
		Full:       DefaultFullCharHalfBlock,
		FullColor:  col,
		Empty:      DefaultEmptyCharBlock,
		EmptyColor: lipgloss.Color("#606060"),
	}

	if !m.springCustomized {
		m.SetSpringOptions(defaultFrequency, defaultDamping)
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

		// If we've more or less reached equilibrium, stop updating.
		if !m.IsAnimating() {
			return m, nil
		}

		m.percentShown, m.velocity = m.spring.Update(m.percentShown, m.velocity, m.targetPercent)
		return m, m.nextFrame()

	default:
		return m, nil
	}
}

// SetSpringOptions sets the frequency and damping for the current spring.
// Frequency corresponds to speed, and damping to bounciness. For details see:
//
// https://github.com/charmbracelet/harmonica
func (m *ProgressModel) SetSpringOptions(frequency, damping float64) {
	m.spring = harmonica.NewSpring(harmonica.FPS(fps), frequency, damping)
}

// Percent returns the current visible percentage on the model. This is only
// relevant when you're animating the progress bar.
//
// If you're rendering with ViewAs you won't need this.
func (m ProgressModel) Percent() float64 {
	return m.targetPercent
}

// SetPercent sets the percentage state of the model as well as a command
// necessary for animating the progress bar to this new percentage.
//
// If you're rendering with ViewAs you won't need this.
func (m *ProgressModel) SetPercent(p float64) tea.Cmd {
	m.targetPercent = math.Max(0, math.Min(1, p))
	m.tag++
	return m.nextFrame()
}

// IncrPercent increments the percentage by a given amount, returning a command
// necessary to animate the progress bar to the new percentage.
//
// If you're rendering with ViewAs you won't need this.
func (m *ProgressModel) IncrPercent(v float64) tea.Cmd {
	return m.SetPercent(m.Percent() + v)
}

// DecrPercent decrements the percentage by a given amount, returning a command
// necessary to animate the progress bar to the new percentage.
//
// If you're rendering with ViewAs you won't need this.
func (m *ProgressModel) DecrPercent(v float64) tea.Cmd {
	return m.SetPercent(m.Percent() - v)
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

func (m ProgressModel) barView(b *strings.Builder, percent float64, textWidth int) {
	var (
		tw = max(0, m.width-textWidth)                // total width
		fw = int(math.Round((float64(tw) * percent))) // filled width
	)

	fw = max(0, min(tw, fw))

	b.WriteString(lipgloss.NewStyle().
		Foreground(m.FullColor).
		Render(strings.Repeat(string(m.Full), fw)))

	// Empty fill.
	n := max(0, tw-fw)
	b.WriteString(lipgloss.NewStyle().
		Foreground(m.EmptyColor).
		Render(strings.Repeat(string(m.Empty), n)))
}

// IsAnimating returns false if the progress bar reached equilibrium and is no
// longer animating.
func (m *ProgressModel) IsAnimating() bool {
	dist := math.Abs(m.percentShown - m.targetPercent)
	return !(dist < 0.001 && m.velocity < 0.01)
}
