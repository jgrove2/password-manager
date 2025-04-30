package input

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"password-manager/internal/models"
)

type InputField struct {
	textinput textinput.Model
	t         string
	title     string
}

func NewInputField(p string) *InputField {
	f := textinput.New()
	f.Placeholder = p
	f.Focus()
	f.EchoMode = textinput.EchoNormal
	f.EchoCharacter = '•'
	return &InputField{textinput: f, t: "textInput"}
}

func (f *InputField) Value() string {
	return f.textinput.Value()
}

func (f *InputField) Focus() {
	f.textinput.Focus()
}

func (f *InputField) Blur() {
	f.textinput.Blur()
}

func (f *InputField) View() string {
	return f.textinput.View()
}

func (f *InputField) Update(msg tea.Msg) (models.DefaultModel, tea.Cmd) {
	var cmd tea.Cmd
	f.textinput, cmd = f.textinput.Update(msg)
	return f, cmd
}

func (f *InputField) SetType(t string) {
	f.t = t
	if t == "password" || t == "passwordInput" || t == "confirmPassword" {
		f.textinput.EchoMode = textinput.EchoPassword
		f.textinput.EchoCharacter = '•'
	}
}

func (f *InputField) Type() string {
	return f.t
}

func (f *InputField) Title() string {
	return f.title
}

func (f *InputField) SetTitle(title string) {
	f.title = title
}
