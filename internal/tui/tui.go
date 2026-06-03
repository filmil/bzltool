package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	textInput textinput.Model
	err       error
	done      bool
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "my_project"
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 40

	return model{
		textInput: ti,
		err:       nil,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.textInput.Value() != "" {
				m.done = true
				return m, tea.Quit
			}
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}
	case error:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.done {
		return ""
	}
	return fmt.Sprintf(
		"Project name is required.\n\nWhat is the project name?\n\n%s\n\n(esc to quit)\n",
		m.textInput.View(),
	)
}

// PromptProjectName uses bubbletea to prompt the user for the project name.
func PromptProjectName() (string, error) {
	p := tea.NewProgram(initialModel())
	m, err := p.Run()
	if err != nil {
		return "", err
	}

	if finalModel, ok := m.(model); ok && finalModel.done {
		return finalModel.textInput.Value(), nil
	}

	return "", fmt.Errorf("project name input was aborted")
}
