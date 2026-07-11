package screens

import (
	"strings"

	"github.com/Yokanater/spotr/commands"
	"github.com/Yokanater/spotr/ui/theme"

	"charm.land/lipgloss/v2"
)

type helpRow struct {
	Label string
	Text  string
}

func HelpView(styles theme.Styles) string {
	contentW := max(20, styles.Help.GetWidth()-helpChrome(styles))
	title := styles.ProgramTitle.Width(contentW).Render("help")

	keys := make([]helpRow, 0, len(commands.KeyBindings))
	for _, binding := range commands.KeyBindings {
		keys = append(keys, helpRow{Label: binding.Key, Text: binding.Action})
	}

	commandRows := make([]helpRow, 0, len(commands.CommandsOrder))
	for _, name := range commands.CommandsOrder {
		commandRows = append(commandRows, helpRow{Label: ":" + name})
	}

	body := renderHelpBody(styles, contentW, keys, commandRows)
	panel := styles.Help.Render(lipgloss.JoinVertical(lipgloss.Left, title, "", body))

	return lipgloss.JoinVertical(
		lipgloss.Center,
		RenderHeader(styles, "help"),
		panel,
	)
}

func renderHelpBody(styles theme.Styles, width int, keys []helpRow, commandRows []helpRow) string {
	if width >= 48 {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			renderKeyGrid(styles, "keys", keys, width, 3),
			"",
			renderCommandBar(styles, commandRows, width),
		)
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		renderKeySection(styles, "keys", keys, width),
		"",
		renderCommandBar(styles, commandRows, width),
	)
}

func renderCommandBar(styles theme.Styles, rows []helpRow, width int) string {
	visible := []string{}
	for _, row := range rows {
		switch row.Label {
		case ":program", ":workout", ":exercise", ":log", ":history", ":template":
			visible = append(visible, row.Label)
		}
	}
	return lipgloss.JoinVertical(
		lipgloss.Left,
		styles.SectionTitle.Width(width).Render("commands"),
		styles.HelpKey.Width(width).Render(strings.Join(visible, "  ")),
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
	labelW := 11
	if width < 22 {
		labelW = max(6, width/2)
	}
	textW := max(4, width-labelW-1)
	for _, row := range rows {
		label := styles.HelpKey.Width(labelW).Render(compactHelpText(row.Label, labelW))
		text := styles.HelpText.Width(textW).Render(compactHelpText(row.Text, textW))
		lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Top, label, " ", text))
	}
	return lipgloss.NewStyle().Width(width).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func compactHelpText(value string, width int) string {
	runes := []rune(value)
	if len(runes) <= width {
		return value
	}
	if width <= 1 {
		return "…"
	}
	return string(runes[:width-1]) + "…"
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
