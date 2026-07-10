package screens

import (
	"strings"

	"spotr/commands"
	"spotr/ui/theme"

	"charm.land/lipgloss/v2"
)

type helpRow struct {
	Label string
	Text  string
}

func HelpView(styles theme.Styles) string {
	contentW := max(20, styles.Help.GetWidth()-helpChrome(styles))
	title := styles.ProgramTitle.Width(contentW).Render("help")
	subtitle := styles.ProgramSubtitle.Width(contentW).Render("shortcuts and command groups")

	keys := make([]helpRow, 0, len(commands.KeyBindings))
	for _, binding := range commands.KeyBindings {
		keys = append(keys, helpRow{Label: binding.Key, Text: binding.Action})
	}

	commandRows := make([]helpRow, 0, len(commands.CommandsOrder))
	for _, name := range commands.CommandsOrder {
		commandRows = append(commandRows, helpRow{Label: ":" + name, Text: commandGroupSummary(name)})
	}

	body := renderHelpBody(styles, contentW, keys, commandRows)
	panel := styles.Help.Render(lipgloss.JoinVertical(lipgloss.Left, title, subtitle, "", body))

	return lipgloss.JoinVertical(
		lipgloss.Center,
		RenderHeader(styles, "help"),
		"",
		panel,
	)
}

func commandGroupSummary(name string) string {
	switch name {
	case "help":
		return "show this screen"
	case "home":
		return "go home"
	case "program":
		return "manage programs"
	case "workout":
		return "manage workouts"
	case "exercise":
		return "manage exercises"
	case "log":
		return "record training"
	case "history":
		return "browse sessions"
	case "template":
		return "manage templates"
	case "quit":
		return "exit spotr"
	default:
		return name
	}
}

func renderHelpBody(styles theme.Styles, width int, keys []helpRow, commandRows []helpRow) string {
	if width >= 76 {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			renderKeyGrid(styles, "keys", keys, width, 3),
			"",
			renderKeyGrid(styles, "commands", commandRows, width, 3),
		)
	}

	if width >= 48 {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			renderKeyGrid(styles, "keys", keys, width, 2),
			"",
			renderKeyGrid(styles, "commands", commandRows, width, 2),
		)
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		renderKeySection(styles, "keys", keys, width),
		"",
		renderKeySection(styles, "commands", commandRows, width),
	)
}

func renderKeyGrid(styles theme.Styles, title string, rows []helpRow, width int, columns int) string {
	gap := 3
	columnW := max(14, (width-(columns-1)*gap)/columns)
	columnBlocks := make([]string, 0, columns)
	for column := 0; column < columns; column++ {
		start := column * len(rows) / columns
		end := (column + 1) * len(rows) / columns
		columnBlocks = append(columnBlocks, renderKeySection(styles, "", rows[start:end], columnW))
	}

	return lipgloss.NewStyle().Width(width).Render(lipgloss.JoinVertical(
		lipgloss.Left,
		styles.SectionTitle.Width(width).Render(title),
		lipgloss.JoinHorizontal(lipgloss.Top, joinWithGap(columnBlocks, gap)...),
	))
}

func renderKeySection(styles theme.Styles, title string, rows []helpRow, width int) string {
	if width < 36 {
		return renderCompactKeySection(styles, title, rows, width)
	}

	labelW := 14
	if width < 62 {
		labelW = 11
	}
	textW := max(8, width-labelW-2)

	lines := []string{}
	if title != "" {
		lines = append(lines, styles.SectionTitle.Width(width).Render(title))
	}
	for _, row := range rows {
		label := styles.HelpKey.Width(labelW).MaxWidth(labelW).Render(row.Label)
		text := styles.HelpText.Width(textW).MaxWidth(textW).Render(row.Text)
		lines = append(lines, styles.HelpRow.Width(width).Render(lipgloss.JoinHorizontal(lipgloss.Top, label, "  ", text)))
	}
	return lipgloss.NewStyle().Width(width).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func renderCompactKeySection(styles theme.Styles, title string, rows []helpRow, width int) string {
	lines := []string{}
	if title != "" {
		lines = append(lines, styles.SectionTitle.Width(width).Render(title))
	}
	for _, row := range rows {
		lines = append(lines,
			styles.HelpKey.Width(width).MaxWidth(width).Render(row.Label),
			styles.HelpText.Width(width).MaxWidth(width).Render(row.Text),
		)
	}
	return lipgloss.NewStyle().Width(width).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func joinWithGap(blocks []string, gap int) []string {
	if len(blocks) == 0 {
		return nil
	}
	spaced := make([]string, 0, len(blocks)*2-1)
	for i, block := range blocks {
		if i > 0 {
			spaced = append(spaced, strings.Repeat(" ", gap))
		}
		spaced = append(spaced, block)
	}
	return spaced
}

func helpChrome(styles theme.Styles) int {
	return styles.Help.GetHorizontalBorderSize() + styles.Help.GetHorizontalPadding()
}
