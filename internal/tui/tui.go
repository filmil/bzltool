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

// selectModel provides a TUI list selection
type selectModel struct {
	choices []string
	cursor  int
	choice  string
	done    bool
	err     error
}

func (m selectModel) Init() tea.Cmd { return nil }

func (m selectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter", " ":
			m.choice = m.choices[m.cursor]
			m.done = true
			return m, tea.Quit
		}
	case error:
		m.err = msg
		return m, nil
	}
	return m, nil
}

func (m selectModel) View() string {
	if m.done {
		return ""
	}
	s := "Select a common configuration template:\n\n"
	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}
	s += "\n(press enter to select, esc to quit)\n"
	return s
}

// PromptTemplate uses bubbletea to prompt the user to select from a list of options.
func PromptTemplate(choices []string) (string, error) {
	p := tea.NewProgram(selectModel{choices: choices})
	m, err := p.Run()
	if err != nil {
		return "", err
	}
	if finalModel, ok := m.(selectModel); ok && finalModel.done {
		return finalModel.choice, nil
	}
	return "", fmt.Errorf("template selection was aborted")
}
