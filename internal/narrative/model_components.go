package narrative

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
)

// Displays percent of source read.
type percRead struct {
	width int
	perc  float64
}

func (p *percRead) SetWidth(w int) *percRead {
	p.width = w
	return p
}

func (p *percRead) SetPerc(perc float64) *percRead {
	p.perc = perc
	return p
}

func (p *percRead) Render() string {
	perc := percReadStyle.Render(fmt.Sprintf(" %3.0f%% Read", p.perc*100))
	return lipgloss.Place(p.width, 1, lipgloss.Center, lipgloss.Center, perc)
}

// Status bar for the main app.
type statusBar struct {
	width  int
	status playbackStatus
}

func (s *statusBar) SetWidth(w int) *statusBar {
	s.width = w
	return s
}

func (s *statusBar) getContentWidth() int {
	return s.width - 9 // 9 is status
}

// Render playback status.
func (s *statusBar) getStatus() string {
	s_col := GetStatusBarStatusColor(PSM.Status())
	return statusBarStatusStyle.Background(s_col).Render(" " + PSM.Status().StatusString())
}

// Internal render compiles all status bar components together.
func (s *statusBar) render(contents string) string {

	contents = lipgloss.NewStyle().
		Width(s.getContentWidth()).
		Background(col.Color4).Render(contents)

	bar := lipgloss.JoinHorizontal(lipgloss.Top,
		contents,
		s.getStatus(),
	)
	return statusBarStyle.Width(s.width).Render(bar)
}

// Render status bar help contents.
func (s *statusBar) Render() string {
	w := s.getContentWidth()
	var builder strings.Builder
	var help_t, temp_t string
	for _, h := range statusHelp {
		builder.WriteString(
			lipgloss.Sprintf("%s%s",
				statusBarStatusTextCommand.Render(h[0]),
				statusBarStatusText.Render(": "+h[1]),
			),
		)
		temp_t = builder.String()
		if lipgloss.Width(temp_t) > w-6 {
			break
		} else {
			help_t = temp_t
		}
		builder.WriteString(statusBarStatusText.Render(" "))
	}
	return s.render(help_t)
}

// Render prompt on the status bar.
func (s *statusBar) Prompt(text string) string {
	contents := lipgloss.Sprintf("%s %s / %s",
		statusBarPromptTextStyle.Render(text),
		statusBarPromptYStyle.Render("y"),
		statusBarPromptNStyle.Render("n"),
	)
	return s.render(contents)
}

// Controller for grabbing fast forward (and backward)
// data.
type fastForwarder struct {
	sMin, sMax, cur int
	start           time.Time
	lastUpd         time.Time
}

func newFastForwarder(sMin, sMax, cur int) *fastForwarder {
	return &fastForwarder{
		sMin:    sMin,
		sMax:    sMax,
		cur:     cur,
		start:   time.Now(),
		lastUpd: time.Now(),
	}
}

// Increase or decrease sentence number.
func (f *fastForwarder) Update(forward bool) {
	now := time.Now()
	diff := now.Sub(f.start).Seconds()
	var change int
	switch {
	case diff < 2:
		change = 1
	case diff < 4:
		change = 2
	case diff < 6:
		change = 5
	default:
		change = 10
	}
	var cur int
	if forward {
		cur = f.cur + change
	} else {
		cur = f.cur - change
	}
	if cur < f.sMin {
		cur = f.sMin
	}
	if cur > f.sMax {
		cur = f.sMax
	}
	f.cur = cur
	MessageCh <- fmt.Sprintf("%d/%d", f.cur, f.sMax)
	f.lastUpd = now
}

// Tick handler.
func (f *fastForwarder) Tick(t time.Time) (int, bool) {
	// Wait 1 second after last upd.
	if int(t.Sub(f.lastUpd).Seconds()) > 0 {
		return f.cur, true
	}
	return f.cur, false
}

type transcript struct {
	width, height int
}

func (t *transcript) SetWidth(w int) {
	t.width = w
}

func (t *transcript) Render(source *Source) string {
	var builder strings.Builder
	sep := lipgloss.Place(
		t.width,
		1,
		lipgloss.Center,
		lipgloss.Center,
		transcriptSeparatorStyle.Render("-------"))
	builder.WriteString(sep)
	builder.WriteString("\n")
	contents := lipgloss.Place(
		t.width, t.height,
		lipgloss.Center,
		lipgloss.Center,
		source.getCurrentRawSentence(),
	)
	builder.WriteString(transcriptStyle.Width(t.width).Height(t.height).Render(contents))
	return builder.String()
}

// Component for rendering help panel.
type helpPanel struct {
	width, height int
}

// Available commands to show in help panel.
func (h *helpPanel) commands() [][2]string {
	return [][2]string{
		{"j/k/Up/Down", "Navigation"},
		{"Enter", "Select text source"},
		{"Space", "Play/Pause"},
		{"h/l", "Fast-Forward/Rewind text source"},
		{"Left/Right", "Previous/Next page"},
		{"/", "Fuzzy Search"},
		{"d", "Delete text source"},
		{"ctrl-b", "Create new bookmark"},
		{"b", "Go to bookmark"},
		{"v", "Change voice"},
		{"t", "Change theme"},
		{"q", "Quit application"},
		{"F1", "Show help"},
	}
}

func (h *helpPanel) Render() string {
	logo := lipgloss.Place(
		h.width,
		lipgloss.Height(LOGO),
		lipgloss.Center,
		lipgloss.Center,
		LOGO,
	)
	logo = helpPanelLogoStyle.Render(logo)

	var elem string
	var elems []string
	for _, c := range h.commands() {
		elem = helpPanelListStyle.Render(
			lipgloss.Sprintf("• %s - %s",
				helpPanelCommandStyle.Render(c[0]),
				helpPanelTextStyle.Render(c[1]),
			),
		)
		elems = append(elems, elem)
	}
	contents := lipgloss.JoinVertical(
		lipgloss.Left,
		logo,
		lipgloss.JoinVertical(lipgloss.Left, elems...))
	return contents
}
