package app

import (
	tea "charm.land/bubbletea/v2"
	"github.com/Yokanater/spotr/ui/theme"
	"strings"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.termW = msg.Width
		m.termH = msg.Height
		m.appW = min(m.termW, min(m.maxW, m.theme.InputMax))
		m.appH = min(m.termH, m.maxH)
		m.styles = theme.NewStyles(m.theme, m.appW, m.appH)
		m.input.SetWidth(max(1, min(m.theme.InputMax, m.appW-m.theme.PadX)))
	case tea.MouseWheelMsg:
		return m.handleMouseWheel(msg)
	case tea.KeyPressMsg:
		switch m.mode {
		case modeQuit:
			return m.handleQuitKey(msg)

		case modeDelete:
			return m.handleDeleteKey(msg)

		case modeCmd:
			return m.handleCommandKey(msg)

		case modeInput:
			return m.handleInputKey(msg)

		case modeNormal:
			return m.handleNormalKey(msg)
		}
	}
	return m, cmd
}

func (m model) handleCommandKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.String() {
	case "enter":
		line := m.input.Value()
		m.input.SetValue("")
		m.mode = modeNormal
		m.inputPurpose = inputNone
		m.resetInputPrompt()
		if line == "" {
			m.status = "command cancelled"
			return m, cmd
		}
		return m.runCommandLine(line)
	case "ctrl+c":
		m.requestQuit()
		return m, cmd
	case "esc":
		m.input.SetValue("")
		m.mode = modeNormal
		m.inputPurpose = inputNone
		m.resetInputPrompt()
		m.status = "command cancelled"
		return m, cmd
	default:
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
}

func (m model) handleInputKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.String() {
	case "enter":
		value := strings.TrimSpace(m.input.Value())
		m.input.SetValue("")
		purpose := m.inputPurpose
		m.mode = modeNormal
		m.inputPurpose = inputNone
		m.resetInputPrompt()
		if value == "" {
			m.status = inputCancelledStatus(purpose)
			return m, cmd
		}
		m.submitInput(purpose, value)
		return m, cmd
	case "ctrl+c":
		m.requestQuit()
		return m, cmd
	case "esc":
		purpose := m.inputPurpose
		m.input.SetValue("")
		m.mode = modeNormal
		m.inputPurpose = inputNone
		m.resetInputPrompt()
		m.status = inputCancelledStatus(purpose)
		return m, cmd
	default:
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
}

func (m model) handleNormalKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.String() {
	case ":":
		m.mode = modeCmd
		m.input.Focus()
		m.inputPurpose = inputNone
		m.input.SetValue("")
		m.input.Placeholder = "program list"
		m.input.Prompt = "spotr $ "
		m.status = "Type a command"
	case "a":
		m.startAdd()
	case "s":
		m.startLogSession()
	case "l":
		m.startLogExerciseInput()
	case "e":
		m.startEditSelectedInput()
	case "d":
		m.requestDeleteSelected()
	case "v":
		m.viewRecentLogs()
	case "t":
		m.openTemplates()
	case "p":
		m.openProgramPicker()
	case "f":
		m.finishLogSession()
	case "down", "j":
		m.moveCursor(1)
	case "up", "k":
		m.moveCursor(-1)
	case "pgdown", "ctrl+f":
		m.moveCursor(5)
	case "pgup", "ctrl+b":
		m.moveCursor(-5)
	case "enter":
		m.openSelected()
	case "?":
		m.helpReturnScreen = m.screen
		m.screen = screenHelp
		m.status = ""
	case "home":
		m.goHome()
	case "b", "esc":
		m.goBack()
	case "q", "ctrl+c":
		m.requestQuit()
	default:
		m.status = ""
	}
	return m, cmd
}

func (m model) handleMouseWheel(msg tea.MouseWheelMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.Mouse().Button {
	case tea.MouseWheelDown:
		m.moveCursor(1)
	case tea.MouseWheelUp:
		m.moveCursor(-1)
	}
	return m, cmd
}

func (m model) handleQuitKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.String() {
	case "y", "Y", "enter":
		return m, tea.Quit
	case "n", "N", "esc", "q":
		m.mode = modeNormal
		m.status = "Quit cancelled"
		return m, cmd
	default:
		m.status = quitConfirmStatus()
		return m, cmd
	}
}

func (m model) handleDeleteKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.String() {
	case "y", "Y", "enter":
		m.confirmDeleteSelected()
		return m, cmd
	case "n", "N", "esc", "q":
		m.clearDeleteState()
		m.status = "Delete cancelled"
		return m, cmd
	default:
		m.status = m.deleteConfirmStatus()
		return m, cmd
	}
}

func (m *model) requestQuit() {
	m.mode = modeQuit
	m.inputPurpose = inputNone
	m.input.SetValue("")
	m.resetInputPrompt()
	m.status = quitConfirmStatus()
}

func quitConfirmStatus() string {
	return "Quit Spotr?"
}
