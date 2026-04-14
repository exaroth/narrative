package debugger

import "charm.land/lipgloss/v2"

func getStatusBar(contents string, r_contents string, width int) string {

	w := lipgloss.Width
	title := statusStyle.Render("Debugger")
	crumbs := statusBarCrumbsStyle.Render("v0.1")
	if len(r_contents) > 0 {
		r_contents = statusBarRContentsStyle.Render(r_contents + "  ")
	}
	help := statusBarStatusText.
		Width(width - w(title) - w(crumbs) - w(r_contents)).
		Render(contents)

	bar := lipgloss.JoinHorizontal(lipgloss.Top,
		title,
		help,
		r_contents,
		crumbs,
	)

	return lipgloss.Sprint(statusBarStyle.Width(width).Render(bar))
}
