package screens

import (
	"fmt"
	"strings"

	"github.com/Yokanater/spotr/ui/theme"

	"charm.land/lipgloss/v2"
)

type TemplateListItem struct {
	Name        string
	Description string
	Path        string
	Workouts    int
	Exercises   int
	Details     []TemplateWorkoutItem
}

type TemplateWorkoutItem struct {
	Name      string
	Exercises []TemplateExerciseItem
}

type TemplateExerciseItem struct {
	Name string
	Sets int
	Reps int
}

func TemplatesView(styles theme.Styles, templates []TemplateListItem, cursor int) string {
	title := styles.ProgramTitle.Render("templates")
	subtitle := styles.ProgramSubtitle.Render("community program templates")
	body := renderTemplateBrowser(styles, templates, cursor)

	return lipgloss.JoinVertical(lipgloss.Left, RenderHeader(styles, "templates"), "", title, subtitle, "", body)
}

func renderTemplateBrowser(styles theme.Styles, templates []TemplateListItem, cursor int) string {
	list := renderTemplateList(styles, templates, cursor)
	if len(templates) == 0 {
		return list
	}

	cursor = min(max(0, cursor), len(templates)-1)
	detail := renderTemplateDetail(styles, templates[cursor])
	if styles.ProgramTitle.GetWidth() < 82 {
		return lipgloss.JoinVertical(lipgloss.Left, list, detail)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, list, detail)
}

func renderTemplateList(styles theme.Styles, templates []TemplateListItem, cursor int) string {
	lines := []string{styles.ProgramPanelTitle.Render("program templates")}
	if styles.ProgramListRows > 1 {
		lines = append(lines, "")
	}
	if len(templates) == 0 {
		lines = append(lines, styles.ProgramEmpty.Render("no templates found"))
		return styles.ProgramPanel.Width(styles.Box.GetWidth()).Render(strings.Join(lines, "\n"))
	}

	start, end := visibleRange(len(templates), cursor, styles.ProgramListRows)
	if start > 0 {
		lines = append(lines, styles.ProgramEmpty.Render("  ..."))
	}
	for i := start; i < end; i++ {
		template := templates[i]
		marker := " "
		rowStyle := styles.ProgramItem
		if i == cursor {
			marker = ">"
			rowStyle = styles.ProgramSelected
		}
		summary := fmt.Sprintf("%s %s  %d workouts / %d exercises", marker, template.Name, template.Workouts, template.Exercises)
		lines = append(lines, rowStyle.Render(summary))
		if template.Description != "" && styles.ProgramListRows > 3 {
			lines = append(lines, styles.ProgramEmpty.Render("  "+template.Description))
		}
	}
	if end < len(templates) {
		lines = append(lines, styles.ProgramEmpty.Render("  ..."))
	}

	return templatePanelStyle(styles).Render(strings.Join(lines, "\n"))
}

func renderTemplateDetail(styles theme.Styles, template TemplateListItem) string {
	lines := []string{styles.ProgramPanelTitle.Render("preview")}
	if styles.ProgramListRows > 1 {
		lines = append(lines, "")
	}
	lines = append(lines, styles.ProgramSelected.Render(template.Name))
	if template.Description != "" {
		lines = append(lines, styles.ProgramEmpty.Render(template.Description))
	}
	if template.Path != "" && styles.ProgramListRows > 5 {
		lines = append(lines, styles.ProgramEmpty.Render(template.Path))
	}
	if len(template.Details) == 0 {
		lines = append(lines, "", styles.ProgramEmpty.Render("no workout details"))
		return templatePanelStyle(styles).Render(strings.Join(lines, "\n"))
	}

	rowsLeft := max(1, styles.ProgramListRows-3)
	for _, workout := range template.Details {
		if rowsLeft <= 0 {
			lines = append(lines, styles.ProgramEmpty.Render("  ..."))
			break
		}
		lines = append(lines, "", styles.SectionTitle.Render(workout.Name))
		rowsLeft--
		for _, exercise := range workout.Exercises {
			if rowsLeft <= 0 {
				lines = append(lines, styles.ProgramEmpty.Render("  ..."))
				break
			}
			lines = append(lines, styles.ProgramItem.Render("  "+templateExerciseLabel(exercise)))
			rowsLeft--
		}
	}

	return templatePanelStyle(styles).Render(strings.Join(lines, "\n"))
}

func templateExerciseLabel(exercise TemplateExerciseItem) string {
	if exercise.Sets > 0 || exercise.Reps > 0 {
		return fmt.Sprintf("%s  %dx%d", exercise.Name, exercise.Sets, exercise.Reps)
	}
	return exercise.Name
}

func templatePanelStyle(styles theme.Styles) lipgloss.Style {
	width := styles.ProgramPanel.GetWidth()
	if styles.ProgramTitle.GetWidth() < 82 {
		width = styles.Box.GetWidth()
	}
	return styles.ProgramPanel.Width(width)
}
