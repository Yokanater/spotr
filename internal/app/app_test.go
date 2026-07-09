package app

import (
	"os"
	"path/filepath"
	"ruffnut/data"
	"ruffnut/store"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestParseLoggedExerciseValue(t *testing.T) {
	sets, reps, repsDetail, weight, notes, err := parseLoggedExerciseValue("4 12 42.5 last set hard")
	if err != nil {
		t.Fatalf("parseLoggedExerciseValue() error = %v", err)
	}
	if sets != 4 || reps != 12 || repsDetail != "" || weight != 42.5 || notes != "last set hard" {
		t.Fatalf("parseLoggedExerciseValue() = %d, %d, %q, %.1f, %q; want 4, 12, empty detail, 42.5, notes", sets, reps, repsDetail, weight, notes)
	}
}

func TestParseLoggedExerciseValueRejectsMissingReps(t *testing.T) {
	_, _, _, _, _, err := parseLoggedExerciseValue("4")
	if err == nil {
		t.Fatal("parseLoggedExerciseValue() error = nil; want usage error")
	}
}

func TestParseLoggedExerciseValueSupportsPerSetReps(t *testing.T) {
	sets, reps, repsDetail, weight, notes, err := parseLoggedExerciseValue("6/4 135 second set cooked")
	if err != nil {
		t.Fatalf("parseLoggedExerciseValue() error = %v", err)
	}
	if sets != 2 || reps != 4 || repsDetail != "6/4" || weight != 135 || notes != "second set cooked" {
		t.Fatalf("parseLoggedExerciseValue() = %d, %d, %q, %.1f, %q; want 2, 4, 6/4, 135, notes", sets, reps, repsDetail, weight, notes)
	}
}

func TestFormatSessionEntryShowsPerSetReps(t *testing.T) {
	entry := data.GymSessionEntry{Exercise: "bench", Sets: 2, Reps: 4, RepsDetail: "6/4", Weight: 135}
	got := formatSessionEntry(entry)
	if !strings.Contains(got, "2x6/4") {
		t.Fatalf("formatSessionEntry() = %q; want per-set reps", got)
	}
}

func TestLogEntryInputValuePrefillsEditableLog(t *testing.T) {
	entry := data.GymSessionEntry{Sets: 2, Reps: 4, RepsDetail: "6/4", Weight: 135, Notes: "second set cooked"}
	got := logEntryInputValue(entry)
	if got != "6/4 135 second set cooked" {
		t.Fatalf("logEntryInputValue() = %q; want editable per-set log value", got)
	}
}

func TestExerciseInputValuePrefillsEditableExercise(t *testing.T) {
	exercise := data.Exercise{Name: "Bench Press", Sets: 3, Reps: 8}
	got := exerciseInputValue(exercise)
	if got != "Bench Press 3 8" {
		t.Fatalf("exerciseInputValue() = %q; want editable exercise value", got)
	}
}

func TestParseExerciseValueSupportsNameWithDefaults(t *testing.T) {
	name, sets, reps, err := parseExerciseValue(strings.Fields("Incline Bench 4 10"))
	if err != nil {
		t.Fatalf("parseExerciseValue() error = %v", err)
	}
	if name != "Incline Bench" || sets != 4 || reps != 10 {
		t.Fatalf("parseExerciseValue() = %q, %d, %d; want Incline Bench, 4, 10", name, sets, reps)
	}
}

func TestHelperMessageUsesDotSeparator(t *testing.T) {
	got := helperMessage("j/k move", "enter open program", "a add program")
	if strings.Contains(got, ",") {
		t.Fatalf("helperMessage() = %q; want no commas", got)
	}
	if !strings.Contains(got, " · ") {
		t.Fatalf("helperMessage() = %q; want dot separators", got)
	}
}

func TestNormalHelpOffersTemplatesWhenNoProgramsExist(t *testing.T) {
	m := model{screen: screenProgram}
	got := m.normalHelp()

	for _, want := range []string{"a add program", "t templates"} {
		if !strings.Contains(got, want) {
			t.Fatalf("normalHelp() = %q; missing %q", got, want)
		}
	}
}

func TestIsHelperKeyRecognizesOnlyKeyTokens(t *testing.T) {
	for _, key := range []string{"a", ":", "?", "enter", "esc", "j/k", "v"} {
		if !isHelperKey(key) {
			t.Fatalf("isHelperKey(%q) = false; want true", key)
		}
	}

	for _, word := range []string{"type", "suggested", "edit", "command"} {
		if isHelperKey(word) {
			t.Fatalf("isHelperKey(%q) = true; want false", word)
		}
	}
}

func TestNormalKeySupportsVimHistoryScroll(t *testing.T) {
	m := model{
		mode:   modeNormal,
		screen: screenHistory,
		historyEntries: []data.GymSessionEntry{
			{EntryId: 1},
			{EntryId: 2},
		},
	}

	updated, _ := m.handleNormalKey(tea.KeyPressMsg{Code: 'j'})
	got := updated.(model)
	if got.historyCursor != 1 {
		t.Fatalf("historyCursor after j = %d; want 1", got.historyCursor)
	}

	updated, _ = got.handleNormalKey(tea.KeyPressMsg{Code: 'k'})
	got = updated.(model)
	if got.historyCursor != 0 {
		t.Fatalf("historyCursor after k = %d; want 0", got.historyCursor)
	}
}

func TestQuitConfirmationUsesHelperStatus(t *testing.T) {
	m := model{mode: modeNormal}
	m.requestQuit()
	if m.status != helperMessage("quit spotr?", "y confirm", "n cancel") {
		t.Fatalf("requestQuit() status = %q; want helper confirmation", m.status)
	}
}

func TestProgramTemplateValidateAndSummary(t *testing.T) {
	t.Setenv(templateDirEnv, filepath.Join("..", "..", "templates", "programs"))

	file, err := findProgramTemplate("Push Pull Legs")
	if err != nil {
		t.Fatalf("findProgramTemplate() error = %v", err)
	}
	tmpl := file.Template
	if tmpl.Name != "Push Pull Legs" || len(tmpl.Workouts) != 3 {
		t.Fatalf("template = %+v; want bundled PPL template", tmpl)
	}
	summary := formatTemplateSummary(tmpl)
	for _, want := range []string{"Push Pull Legs", "3 workouts", "12 exercises"} {
		if !strings.Contains(summary, want) {
			t.Fatalf("formatTemplateSummary() = %q; missing %q", summary, want)
		}
	}
}

func TestBundledProgramTemplatesAreValid(t *testing.T) {
	dir := filepath.Join("..", "..", "templates", "programs")
	if err := validateProgramTemplateDir(dir); err != nil {
		t.Fatalf("validateProgramTemplateDir(%q) error = %v", dir, err)
	}
	files, err := listProgramTemplates(dir)
	if err != nil {
		t.Fatalf("listProgramTemplates(%q) error = %v", dir, err)
	}
	if len(files) == 0 {
		t.Fatalf("listProgramTemplates(%q) returned no templates", dir)
	}
}

func TestValidateProgramTemplateDirRejectsFilenameMismatch(t *testing.T) {
	dir := t.TempDir()
	tmpl := programTemplate{
		Name:    "Name Mismatch",
		Version: programTemplateVersion,
		Workouts: []workoutTemplate{{
			Name: "Workout",
			Exercises: []exerciseTemplate{
				{Name: "Bench Press", Sets: 3, Reps: 8},
			},
		}},
	}
	if err := saveProgramTemplate(filepath.Join(dir, "not-the-name.json"), tmpl); err != nil {
		t.Fatalf("saveProgramTemplate() error = %v", err)
	}
	if err := validateProgramTemplateDir(dir); err == nil {
		t.Fatal("validateProgramTemplateDir() error = nil; want filename mismatch error")
	}
}

func TestCreateAndExportProgramTemplate(t *testing.T) {
	st, err := store.NewSQLite(filepath.Join(t.TempDir(), "ruffnut.db"))
	if err != nil {
		t.Fatalf("NewSQLite() error = %v", err)
	}
	defer st.Close()

	m := initialModel(st)
	tmpl := programTemplate{
		Name:    "Tiny Template",
		Version: programTemplateVersion,
		Workouts: []workoutTemplate{{
			Name: "Push",
			Exercises: []exerciseTemplate{
				{Name: "Bench Press", Sets: 3, Reps: 8},
			},
		}},
	}
	if err := m.createProgramFromTemplate(tmpl); err != nil {
		t.Fatalf("createProgramFromTemplate() error = %v", err)
	}
	if m.activeProgram.ProgramName != "Tiny Template" {
		t.Fatalf("activeProgram = %+v; want created template program", m.activeProgram)
	}
	workouts, err := st.ListWorkouts(m.activeProgram)
	if err != nil {
		t.Fatalf("ListWorkouts() error = %v", err)
	}
	if len(workouts) != 1 || workouts[0].Name != "Push" {
		t.Fatalf("workouts = %+v; want Push workout", workouts)
	}
	exercises, err := st.ListExercises(workouts[0])
	if err != nil {
		t.Fatalf("ListExercises() error = %v", err)
	}
	if len(exercises) != 1 || exercises[0].Name != "Bench Press" || exercises[0].Sets != 3 || exercises[0].Reps != 8 {
		t.Fatalf("exercises = %+v; want Bench Press 3x8", exercises)
	}

	exported, err := m.exportProgramTemplate(m.activeProgram)
	if err != nil {
		t.Fatalf("exportProgramTemplate() error = %v", err)
	}
	if exported.Name != "Tiny Template" || len(exported.Workouts) != 1 || len(exported.Workouts[0].Exercises) != 1 {
		t.Fatalf("exported = %+v; want matching template", exported)
	}
}

func TestProgramTemplateDirUsesEnvironmentOverride(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "program-templates")
	t.Setenv(templateDirEnv, dir)

	if got := programTemplateDir(); got != dir {
		t.Fatalf("programTemplateDir() = %q; want %q", got, dir)
	}
	if got := defaultProgramTemplatePath("My Program"); got != filepath.Join(dir, "my-program.json") {
		t.Fatalf("defaultProgramTemplatePath() = %q; want path inside env template dir", got)
	}
}

func TestOpenTemplatesHandlesMissingDirectory(t *testing.T) {
	missingDir := filepath.Join(t.TempDir(), "missing-templates")
	t.Setenv(templateDirEnv, missingDir)

	m := model{screen: screenProgram}
	m.openTemplates()

	if m.screen != screenTemplates {
		t.Fatalf("screen = %q; want templates", m.screen)
	}
	if len(m.templateFiles) != 0 {
		t.Fatalf("templateFiles = %+v; want none for missing directory", m.templateFiles)
	}
	if m.status != helperMessage("b back", ": command") {
		t.Fatalf("status = %q; want empty template browser help", m.status)
	}
}

func TestTemplateCommandImportAndExport(t *testing.T) {
	t.Chdir(filepath.Join("..", ".."))

	st, err := store.NewSQLite(filepath.Join(t.TempDir(), "ruffnut.db"))
	if err != nil {
		t.Fatalf("NewSQLite() error = %v", err)
	}
	defer st.Close()

	m := initialModel(st)
	updated, _ := m.runCommandLine("template import Push Pull Legs")
	m = updated.(model)
	if m.status != "Created program from template Push Pull Legs" {
		t.Fatalf("status after import = %q; want created template status", m.status)
	}
	if m.activeProgram.ProgramName != "Push Pull Legs" {
		t.Fatalf("activeProgram = %+v; want Push Pull Legs", m.activeProgram)
	}

	programs, err := st.ListPrograms()
	if err != nil {
		t.Fatalf("ListPrograms() error = %v", err)
	}
	if len(programs) != 1 || programs[0].ProgramName != "Push Pull Legs" {
		t.Fatalf("programs = %+v; want imported PPL program", programs)
	}

	output := filepath.Join(t.TempDir(), "exported-ppl.json")
	updated, _ = m.runCommandLine("template export Push Pull Legs " + output)
	m = updated.(model)
	if m.status != "Exported template Push Pull Legs to "+output {
		t.Fatalf("status after export = %q; want exported status", m.status)
	}
	if _, err := os.Stat(output); err != nil {
		t.Fatalf("exported template missing: %v", err)
	}
	exported, err := loadProgramTemplate(output)
	if err != nil {
		t.Fatalf("loadProgramTemplate(exported) error = %v", err)
	}
	if exported.Name != "Push Pull Legs" || len(exported.Workouts) != 3 {
		t.Fatalf("exported = %+v; want PPL template", exported)
	}
}

func TestTemplateWorkoutCommandImportsWorkoutIntoActiveProgram(t *testing.T) {
	t.Chdir(filepath.Join("..", ".."))

	st, err := store.NewSQLite(filepath.Join(t.TempDir(), "ruffnut.db"))
	if err != nil {
		t.Fatalf("NewSQLite() error = %v", err)
	}
	defer st.Close()

	m := initialModel(st)
	updated, _ := m.runCommandLine("program add Custom")
	m = updated.(model)
	m.activeProgram = m.programs[m.programCursor]

	updated, _ = m.runCommandLine("template workout Push Pull Legs Push")
	m = updated.(model)
	if m.status != "Created workout from template Push" {
		t.Fatalf("status = %q; want workout import status", m.status)
	}
	if m.activeWorkout.Name != "Push" {
		t.Fatalf("activeWorkout = %+v; want Push", m.activeWorkout)
	}
	exercises, err := st.ListExercises(m.activeWorkout)
	if err != nil {
		t.Fatalf("ListExercises() error = %v", err)
	}
	if len(exercises) != 4 || exercises[0].Name == "" {
		t.Fatalf("exercises = %+v; want imported template exercises", exercises)
	}
}

func TestTemplateListCommandOpensTemplateBrowser(t *testing.T) {
	t.Chdir(filepath.Join("..", ".."))

	st, err := store.NewSQLite(filepath.Join(t.TempDir(), "ruffnut.db"))
	if err != nil {
		t.Fatalf("NewSQLite() error = %v", err)
	}
	defer st.Close()

	m := initialModel(st)
	updated, _ := m.runCommandLine("template list")
	m = updated.(model)
	if m.screen != screenTemplates {
		t.Fatalf("screen = %q; want templates", m.screen)
	}
	if len(m.templateFiles) == 0 {
		t.Fatal("templateFiles empty; want bundled templates")
	}
	if m.status != helperMessage("j/k move", "enter import", "b back", ": command") {
		t.Fatalf("status = %q; want template browser help", m.status)
	}
}

func TestTemplateBrowserImportsSelectedTemplate(t *testing.T) {
	t.Chdir(filepath.Join("..", ".."))

	st, err := store.NewSQLite(filepath.Join(t.TempDir(), "ruffnut.db"))
	if err != nil {
		t.Fatalf("NewSQLite() error = %v", err)
	}
	defer st.Close()

	m := initialModel(st)
	m.openTemplates()
	if len(m.templateFiles) < 2 {
		t.Fatalf("templateFiles = %d; want at least two bundled templates", len(m.templateFiles))
	}

	updated, _ := m.handleNormalKey(tea.KeyPressMsg{Code: 'j'})
	m = updated.(model)
	if m.templateCursor != 1 {
		t.Fatalf("templateCursor after j = %d; want 1", m.templateCursor)
	}

	selectedName := m.templateFiles[m.templateCursor].Template.Name
	updated, _ = m.handleNormalKey(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = updated.(model)
	if m.screen != screenProgram {
		t.Fatalf("screen after import = %q; want program", m.screen)
	}
	if m.activeProgram.ProgramName != selectedName {
		t.Fatalf("activeProgram = %+v; want imported %q", m.activeProgram, selectedName)
	}
}

func TestTemplateImportSelectsExistingProgram(t *testing.T) {
	st, err := store.NewSQLite(filepath.Join(t.TempDir(), "ruffnut.db"))
	if err != nil {
		t.Fatalf("NewSQLite() error = %v", err)
	}
	defer st.Close()

	m := initialModel(st)
	tmpl := programTemplate{
		Name:    "Existing Template",
		Version: programTemplateVersion,
		Workouts: []workoutTemplate{{
			Name: "Push",
			Exercises: []exerciseTemplate{
				{Name: "Bench Press", Sets: 3, Reps: 8},
			},
		}},
	}
	created, err := m.importProgramTemplate(tmpl)
	if err != nil {
		t.Fatalf("first importProgramTemplate() error = %v", err)
	}
	if !created {
		t.Fatal("first importProgramTemplate() created = false; want true")
	}
	created, err = m.importProgramTemplate(tmpl)
	if err != nil {
		t.Fatalf("second importProgramTemplate() error = %v", err)
	}
	if created {
		t.Fatal("second importProgramTemplate() created = true; want false")
	}
	programs, err := st.ListPrograms()
	if err != nil {
		t.Fatalf("ListPrograms() error = %v", err)
	}
	if len(programs) != 1 || m.activeProgram.ProgramName != "Existing Template" {
		t.Fatalf("programs = %+v, activeProgram = %+v; want one selected existing program", programs, m.activeProgram)
	}
}
