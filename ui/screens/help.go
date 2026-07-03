package screens

import (
	"ruffnut/commands"
	"ruffnut/ui/theme"

	"charm.land/lipgloss/v2"
)

func HelpView(styles theme.Styles) string {
	rowW := styles.HelpRow.GetWidth()
	stacked := styles.Help.GetWidth() < 82
	if stacked {
		rowW = max(1, styles.Help.GetWidth()-8)
	}

	keys := []string{styles.SectionTitle.Render("keys")}
	for _, binding := range commands.KeyBindings {
		keys = append(keys, renderHelpRow(styles, binding.Key, binding.Action, rowW))
	}

	commandRows := []string{styles.SectionTitle.Render("commands")}
	for _, name := range commands.CommandsOrder {
		spec := commands.Registry[name]
		commandRows = append(commandRows, renderHelpRow(styles, spec.Usage, spec.Summary, rowW))
	}

	keysBlock := lipgloss.JoinVertical(lipgloss.Left, keys...)
	commandsBlock := lipgloss.JoinVertical(lipgloss.Left, commandRows...)
	body := lipgloss.JoinHorizontal(lipgloss.Top, keysBlock, "    ", commandsBlock)
	if stacked {
		body = lipgloss.JoinVertical(lipgloss.Left, keysBlock, "", commandsBlock)
	}

	return lipgloss.JoinVertical(
		lipgloss.Center,
		RenderHeader(styles, "help"),
		"",
		styles.Help.Render(body),
	)
}

func renderHelpRow(styles theme.Styles, key string, summary string, rowW int) string {
	keyW := min(styles.HelpKey.GetWidth(), rowW)
	if rowW < styles.HelpKey.GetWidth()+12 {
		return styles.HelpRow.Width(rowW).Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				styles.HelpKey.Width(rowW).MaxWidth(rowW).Render(key),
				styles.HelpText.Width(rowW).MaxWidth(rowW).Render(summary),
			),
		)
	}

	textW := max(1, rowW-keyW)
	return styles.HelpRow.Width(rowW).Render(
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			styles.HelpKey.Width(keyW).Render(key),
			styles.HelpText.Width(textW).MaxWidth(textW).Render(summary),
		),
	)
}
