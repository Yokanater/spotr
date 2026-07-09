package app

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"ruffnut/data"
	"ruffnut/ui/screens"
	"sort"
	"strings"
)

const programTemplateVersion = 1
const defaultTemplateDir = "templates/programs"
const templateDirEnv = "SPOTR_TEMPLATE_DIR"

type programTemplate struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Version     int               `json:"version"`
	Workouts    []workoutTemplate `json:"workouts"`
}

type workoutTemplate struct {
	Name      string             `json:"name"`
	Exercises []exerciseTemplate `json:"exercises"`
}

type exerciseTemplate struct {
	Name string `json:"name"`
	Sets int    `json:"sets,omitempty"`
	Reps int    `json:"reps,omitempty"`
}

func loadProgramTemplate(path string) (programTemplate, error) {
	var tmpl programTemplate
	content, err := os.ReadFile(path)
	if err != nil {
		return tmpl, err
	}
	if err := json.Unmarshal(content, &tmpl); err != nil {
		return tmpl, err
	}
	if err := validateProgramTemplate(tmpl); err != nil {
		return tmpl, err
	}
	return tmpl, nil
}

func saveProgramTemplate(path string, tmpl programTemplate) error {
	if err := validateProgramTemplate(tmpl); err != nil {
		return err
	}
	content, err := json.MarshalIndent(tmpl, "", "  ")
	if err != nil {
		return err
	}
	content = append(content, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, content, 0o644)
}

func validateProgramTemplate(tmpl programTemplate) error {
	if strings.TrimSpace(tmpl.Name) == "" {
		return fmt.Errorf("template name is required")
	}
	if tmpl.Version != programTemplateVersion {
		return fmt.Errorf("template %s has unsupported version %d", tmpl.Name, tmpl.Version)
	}
	if len(tmpl.Workouts) == 0 {
		return fmt.Errorf("template %s needs at least one workout", tmpl.Name)
	}
	for _, workout := range tmpl.Workouts {
		if strings.TrimSpace(workout.Name) == "" {
			return fmt.Errorf("template %s has a workout with no name", tmpl.Name)
		}
		if len(workout.Exercises) == 0 {
			return fmt.Errorf("workout %s needs at least one exercise", workout.Name)
		}
		for _, exercise := range workout.Exercises {
			if strings.TrimSpace(exercise.Name) == "" {
				return fmt.Errorf("workout %s has an exercise with no name", workout.Name)
			}
			if exercise.Sets < 0 || exercise.Reps < 0 {
				return fmt.Errorf("exercise %s cannot have negative sets or reps", exercise.Name)
			}
		}
	}
	return nil
}

func listProgramTemplates(dir string) ([]programTemplateFile, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	files := []programTemplateFile{}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		tmpl, err := loadProgramTemplate(path)
		if err != nil {
			return nil, err
		}
		files = append(files, programTemplateFile{Path: path, Template: tmpl})
	}
	sort.Slice(files, func(i int, j int) bool {
		return strings.ToLower(files[i].Template.Name) < strings.ToLower(files[j].Template.Name)
	})
	return files, nil
}

func validateProgramTemplateDir(dir string) error {
	files, err := listProgramTemplates(dir)
	if err != nil {
		return err
	}
	seen := map[string]string{}
	for _, file := range files {
		base := filepath.Base(file.Path)
		expected := slugify(file.Template.Name) + ".json"
		if base != expected {
			return fmt.Errorf("template %s should be named %s, got %s", file.Template.Name, expected, base)
		}
		nameKey := strings.ToLower(file.Template.Name)
		if previous, ok := seen[nameKey]; ok {
			return fmt.Errorf("duplicate template name %s in %s and %s", file.Template.Name, previous, file.Path)
		}
		seen[nameKey] = file.Path
	}
	return nil
}

type programTemplateFile struct {
	Path     string
	Template programTemplate
}

func (m *model) openTemplates() {
	if err := m.loadTemplates(); err != nil {
		m.status = err.Error()
		return
	}
	m.screen = screenTemplates
	m.status = m.normalHelp()
}

func (m *model) loadTemplates() error {
	files, err := listProgramTemplates(programTemplateDir())
	if os.IsNotExist(err) {
		m.templateFiles = nil
		m.templateCursor = 0
		return nil
	}
	if err != nil {
		return err
	}
	m.templateFiles = files
	m.templateCursor = clampIndex(m.templateCursor, len(m.templateFiles))
	return nil
}

func (m model) templateItems() []screens.TemplateListItem {
	items := make([]screens.TemplateListItem, 0, len(m.templateFiles))
	for _, file := range m.templateFiles {
		workoutCount, exerciseCount := templateCounts(file.Template)
		items = append(items, screens.TemplateListItem{
			Name:        file.Template.Name,
			Description: file.Template.Description,
			Path:        file.Path,
			Workouts:    workoutCount,
			Exercises:   exerciseCount,
			Details:     templateDetailItems(file.Template),
		})
	}
	return items
}

func (m *model) importSelectedTemplate() {
	if len(m.templateFiles) == 0 {
		m.status = "no templates found"
		return
	}
	m.templateCursor = clampIndex(m.templateCursor, len(m.templateFiles))
	file := m.templateFiles[m.templateCursor]
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
}

func findProgramTemplate(nameOrPath string) (programTemplateFile, error) {
	if filepath.Ext(nameOrPath) == ".json" || strings.ContainsRune(nameOrPath, filepath.Separator) {
		tmpl, err := loadProgramTemplate(nameOrPath)
		return programTemplateFile{Path: nameOrPath, Template: tmpl}, err
	}
	files, err := listProgramTemplates(programTemplateDir())
	if err != nil {
		return programTemplateFile{}, err
	}
	needle := strings.ToLower(strings.TrimSpace(nameOrPath))
	for _, file := range files {
		if strings.ToLower(file.Template.Name) == needle || strings.TrimSuffix(filepath.Base(file.Path), ".json") == slugify(nameOrPath) {
			return file, nil
		}
	}
	return programTemplateFile{}, fmt.Errorf("template not found: %s", nameOrPath)
}

func (m *model) importProgramTemplate(tmpl programTemplate) (bool, error) {
	program, err := m.store.SelectProgram(tmpl.Name)
	if err == nil {
		m.activeProgram = program
		if err := m.loadPrograms(); err != nil {
			return false, err
		}
		for i := range m.programs {
			if m.programs[i].ProgramId == program.ProgramId {
				m.programCursor = i
				break
			}
		}
		if err := m.loadWorkouts(program); err != nil {
			return false, err
		}
		m.activeWorkout = data.Workout{}
		m.activeExercise = data.Exercise{}
		m.exerciseCursor = 0
		m.exercises = nil
		return false, nil
	}
	if err != sql.ErrNoRows {
		return false, err
	}
	if err := m.createProgramFromTemplate(tmpl); err != nil {
		return false, err
	}
	return true, nil
}

func (m *model) importWorkoutTemplate(tmpl programTemplate, workoutName string) (data.Workout, bool, error) {
	if m.activeProgram.ProgramId == 0 {
		return data.Workout{}, false, fmt.Errorf("select a program first: program select <id|name>")
	}
	workoutTmpl, err := findWorkoutTemplate(tmpl, workoutName)
	if err != nil {
		return data.Workout{}, false, err
	}
	if workout, err := m.store.SelectWorkout(workoutTmpl.Name, m.activeProgram); err == nil {
		if err := m.selectImportedWorkout(workout); err != nil {
			return data.Workout{}, false, err
		}
		return workout, false, nil
	} else if err != sql.ErrNoRows {
		return data.Workout{}, false, err
	}
	if err := m.store.CreateWorkout(workoutTmpl.Name, m.activeProgram); err != nil {
		return data.Workout{}, false, err
	}
	workout, err := m.store.SelectWorkout(workoutTmpl.Name, m.activeProgram)
	if err != nil {
		return data.Workout{}, false, err
	}
	for _, exerciseTmpl := range workoutTmpl.Exercises {
		if err := m.store.CreateExercise(exerciseTmpl.Name, exerciseTmpl.Sets, exerciseTmpl.Reps, workout); err != nil {
			return data.Workout{}, false, err
		}
	}
	if err := m.selectImportedWorkout(workout); err != nil {
		return data.Workout{}, false, err
	}
	return workout, true, nil
}

func (m *model) selectImportedWorkout(workout data.Workout) error {
	m.activeWorkout = workout
	m.activeExercise = data.Exercise{}
	m.exerciseCursor = 0
	if err := m.loadWorkouts(m.activeProgram); err != nil {
		return err
	}
	for i := range m.workouts {
		if m.workouts[i].WorkoutId == workout.WorkoutId {
			m.workoutCursor = i
			break
		}
	}
	if err := m.loadExercises(workout); err != nil {
		return err
	}
	return nil
}

func findWorkoutTemplate(tmpl programTemplate, workoutName string) (workoutTemplate, error) {
	needle := strings.ToLower(strings.TrimSpace(workoutName))
	for _, workout := range tmpl.Workouts {
		if strings.ToLower(workout.Name) == needle || slugify(workout.Name) == slugify(workoutName) {
			return workout, nil
		}
	}
	return workoutTemplate{}, fmt.Errorf("workout template not found: %s", workoutName)
}

func (m *model) createProgramFromTemplate(tmpl programTemplate) error {
	programID, err := m.store.CreateProgram(tmpl.Name)
	if err != nil {
		return err
	}
	program := data.Program{ProgramId: programID, ProgramName: tmpl.Name}
	for _, workoutTmpl := range tmpl.Workouts {
		if err := m.store.CreateWorkout(workoutTmpl.Name, program); err != nil {
			return err
		}
		workout, err := m.store.SelectWorkout(workoutTmpl.Name, program)
		if err != nil {
			return err
		}
		for _, exerciseTmpl := range workoutTmpl.Exercises {
			if err := m.store.CreateExercise(exerciseTmpl.Name, exerciseTmpl.Sets, exerciseTmpl.Reps, workout); err != nil {
				return err
			}
		}
	}
	m.activeProgram = program
	if err := m.loadPrograms(); err != nil {
		return err
	}
	for i := range m.programs {
		if m.programs[i].ProgramId == programID {
			m.programCursor = i
			break
		}
	}
	return m.loadWorkouts(program)
}

func (m *model) exportProgramTemplate(program data.Program) (programTemplate, error) {
	workouts, err := m.store.ListWorkouts(program)
	if err != nil {
		return programTemplate{}, err
	}
	tmpl := programTemplate{
		Name:    program.ProgramName,
		Version: programTemplateVersion,
	}
	for _, workout := range workouts {
		exercises, err := m.store.ListExercises(workout)
		if err != nil {
			return programTemplate{}, err
		}
		workoutTmpl := workoutTemplate{Name: workout.Name}
		for _, exercise := range exercises {
			workoutTmpl.Exercises = append(workoutTmpl.Exercises, exerciseTemplate{
				Name: exercise.Name,
				Sets: exercise.Sets,
				Reps: exercise.Reps,
			})
		}
		tmpl.Workouts = append(tmpl.Workouts, workoutTmpl)
	}
	if err := validateProgramTemplate(tmpl); err != nil {
		return programTemplate{}, err
	}
	return tmpl, nil
}

func defaultProgramTemplatePath(name string) string {
	return filepath.Join(programTemplateDir(), slugify(name)+".json")
}

func formatTemplateSummary(tmpl programTemplate) string {
	workoutCount, exerciseCount := templateCounts(tmpl)
	summary := fmt.Sprintf("%s: %d workouts, %d exercises", tmpl.Name, workoutCount, exerciseCount)
	if tmpl.Description != "" {
		summary += " - " + tmpl.Description
	}
	return summary
}

func templateCounts(tmpl programTemplate) (int, int) {
	exerciseCount := 0
	for _, workout := range tmpl.Workouts {
		exerciseCount += len(workout.Exercises)
	}
	return len(tmpl.Workouts), exerciseCount
}

func templateDetailItems(tmpl programTemplate) []screens.TemplateWorkoutItem {
	workouts := make([]screens.TemplateWorkoutItem, 0, len(tmpl.Workouts))
	for _, workout := range tmpl.Workouts {
		item := screens.TemplateWorkoutItem{Name: workout.Name}
		for _, exercise := range workout.Exercises {
			item.Exercises = append(item.Exercises, screens.TemplateExerciseItem{
				Name: exercise.Name,
				Sets: exercise.Sets,
				Reps: exercise.Reps,
			})
		}
		workouts = append(workouts, item)
	}
	return workouts
}

func validateTemplateTarget(target string) error {
	info, err := os.Stat(target)
	if err == nil && info.IsDir() {
		return validateProgramTemplateDir(target)
	}
	if err == nil {
		_, err := loadProgramTemplate(target)
		return err
	}
	if os.IsNotExist(err) {
		_, err = findProgramTemplate(target)
		return err
	}
	return err
}

func programTemplateDir() string {
	if dir := strings.TrimSpace(os.Getenv(templateDirEnv)); dir != "" {
		return dir
	}
	if dir, ok := firstExistingDir(defaultTemplateDir, filepath.Join("..", defaultTemplateDir), filepath.Join("..", "..", defaultTemplateDir)); ok {
		return dir
	}
	return defaultTemplateDir
}

func firstExistingDir(paths ...string) (string, bool) {
	for _, path := range paths {
		info, err := os.Stat(path)
		if err == nil && info.IsDir() {
			return path, true
		}
	}
	return "", false
}

var slugCleanup = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = slugCleanup.ReplaceAllString(value, "-")
	value = strings.Trim(value, "-")
	if value == "" {
		return "template"
	}
	return value
}
