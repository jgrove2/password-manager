package choice

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"strings"

	"password-manager/internal/models"
)

type ChoiceModel struct {
	t        string
	cursor   int
	choice   string
	title    string
	choices  []string
	question string
}

func NewChoiceInput(question string, choices []string) *ChoiceModel {
	c := &ChoiceModel{
		t:        "choiceInput",
		cursor:   0,
		question: question,
		choices:  choices,
	}
	return c
}

func (m *ChoiceModel) Update(msg tea.Msg) (models.DefaultModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit

		//case "enter":
		// Send the choice on the channel and exit.
		//	m.choice = m.choices[m.cursor]
		//	return m, tea.Quit

		case "down", "j":
			m.cursor++
			if m.cursor >= len(m.choices) {
				m.cursor = 0
			}
			m.choice = m.choices[m.cursor]

		case "up", "k":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.choices) - 1
			}
			m.choice = m.choices[m.cursor]
		}
	}

	return m, nil
}

func (f *ChoiceModel) View() string {
	s := strings.Builder{}
	question := fmt.Sprintf("%s\n\n", f.question)
	s.WriteString(question)

	for i := 0; i < len(f.choices); i++ {
		if f.cursor == i {
			s.WriteString("(•) ")
		} else {
			s.WriteString("( ) ")
		}
		s.WriteString(f.choices[i])
		s.WriteString("\n")
	}
	s.WriteString("\n(press q to quit)\n")

	return s.String()
}

func (f *ChoiceModel) Value() string {
	return f.choice
}

func (f *ChoiceModel) Type() string {
	return f.t
}

func (f *ChoiceModel) SetType(t string) {
	f.t = t
}

func (f *ChoiceModel) Title() string {
	return f.title
}

func (f *ChoiceModel) SetTitle(title string) {
	f.title = title
}
