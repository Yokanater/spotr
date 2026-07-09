package app

import (
	"fmt"
	"path/filepath"
	"spotr/data"
	"strings"
)

func (m *model) handleTemplate(args []string) {
	if len(args) == 0 {
		m.status = "usage: template list | show <name|path> | import <name|path> | workout <template> <workout> | export [program] [path] | validate [name|path]"
		return
	}

	switch args[0] {
	case "list":
		m.openTemplates()

	case "show":
		if len(args) < 2 {
			m.status = "usage: template show <name|path>"
			return
		}
		file, err := findProgramTemplate(strings.Join(args[1:], " "))
		if err != nil {
			m.status = err.Error()
			return
		}
		m.status = formatTemplateSummary(file.Template)

	case "import", "use":
		if len(args) < 2 {
			m.status = "usage: template import <name|path>"
			return
		}
		file, err := findProgramTemplate(strings.Join(args[1:], " "))
		if err != nil {
			m.status = err.Error()
			return
		}
		created, err := m.importProgramTemplate(file.Template)
		if err != nil {
			m.status = err.Error()
			return
		}
		m.screen = screenProgram
		if created {
			m.status = "Created program from template " + file.Template.Name
			return
		}
		m.status = "Selected existing program " + file.Template.Name

	case "workout":
		if len(args) < 3 {
			m.status = "usage: template workout <template> <workout>"
			return
		}
		file, workoutName, err := findTemplateWorkoutArgs(args[1:])
		if err != nil {
			m.status = err.Error()
			return
		}
		workout, created, err := m.importWorkoutTemplate(file.Template, workoutName)
		if err != nil {
			m.status = err.Error()
			return
		}
		m.screen = screenProgram
		if created {
			m.status = "Created workout from template " + workout.Name
			return
		}
		m.status = "Selected existing workout " + workout.Name

	case "export":
		program, output, ok := m.templateExportArgs(args[1:])
		if !ok {
			return
		}
		tmpl, err := m.exportProgramTemplate(program)
		if err != nil {
			m.status = err.Error()
			return
		}
		if output == "" {
			output = defaultProgramTemplatePath(tmpl.Name)
		}
		if err := saveProgramTemplate(output, tmpl); err != nil {
			m.status = err.Error()
			return
		}
		m.status = "Exported template " + tmpl.Name + " to " + output

	case "validate":
		target := programTemplateDir()
		if len(args) >= 2 {
			target = strings.Join(args[1:], " ")
		}
		if err := validateTemplateTarget(target); err != nil {
			m.status = err.Error()
			return
		}
		m.status = "Templates valid"

	default:
		m.status = fmt.Sprintf("unknown template command: %s", args[0])
	}
}

func findTemplateWorkoutArgs(args []string) (programTemplateFile, string, error) {
	if len(args) < 2 {
		return programTemplateFile{}, "", fmt.Errorf("usage: template workout <template> <workout>")
	}
	for split := len(args) - 1; split >= 1; split-- {
		file, err := findProgramTemplate(strings.Join(args[:split], " "))
		if err == nil {
			return file, strings.Join(args[split:], " "), nil
		}
	}
	return programTemplateFile{}, "", fmt.Errorf("template not found: %s", strings.Join(args[:len(args)-1], " "))
}

func (m *model) templateExportArgs(args []string) (data.Program, string, bool) {
	if len(args) == 0 {
		if m.activeProgram.ProgramId == 0 {
			program, ok := m.selectedProgram()
			if !ok {
				m.status = "select a program first or pass one to export"
				return data.Program{}, "", false
			}
			return program, "", true
		}
		return m.activeProgram, "", true
	}

	output := ""
	programArgs := args
	last := args[len(args)-1]
	if strings.HasSuffix(last, ".json") || strings.Contains(last, string(filepath.Separator)) {
		output = last
		programArgs = args[:len(args)-1]
	}
	if len(programArgs) == 0 {
		if m.activeProgram.ProgramId == 0 {
			m.status = "select a program first or pass one to export"
			return data.Program{}, "", false
		}
		return m.activeProgram, output, true
	}
	program, err := m.store.SelectProgram(strings.Join(programArgs, " "))
	if err != nil {
		m.status = err.Error()
		return data.Program{}, "", false
	}
	return program, output, true
}
