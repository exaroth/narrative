package debugger

import "charm.land/lipgloss/v2"

type helpPanel struct {
	width, height int
}

// Available commands to show in help panel.
func (h *helpPanel) commands() [][2]string {
	return [][2]string{
		{"j/k", "Next/previous sentence"},
		{"h/l", "Select word/phoneme"},
		{"Up/Down/Left/Right", "Viewport scroll"},
		{"Enter", "Show phoneme selection panel"},
		{"Space", "Play/Pause"},
		{"c", "Continuous mode on/off"},
		{"m", "Mark sentence for review"},
		{"<num>s", "Select sentence at <num>"},
		{"<num>w", "Select word at <num>"},
		{"Left/Right", "Previous/Next page"},
		{"Ctrl-a", "(Phoneme panel) Add new phoneme"},
		{"Ctrl-d", "(Phoneme panel) Delete override phoneme"},
		{"Enter", "(Phoneme panel) Set phoneme at cursor as override"},
		{"?", "Show help"},
		{"Q/Ctrl-C", "Exit debugger"},
	}
}

func (h *helpPanel) Render() string {
	logo := lipgloss.Place(
		h.width,
		1,
		lipgloss.Center,
		lipgloss.Center,
		"Narrative Debugger",
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
