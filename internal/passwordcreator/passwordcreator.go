package passwordcreator

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"password-manager/internal/models"
	sqlite "password-manager/internal/sqlite"
)

type goToPasswordList func()

type PasswordCreatorModel struct {
	title            string
	accountName      textinput.Model
	username         textinput.Model
	password         textinput.Model
	err              error
	focusIndex       int
	sqliteMan        *sqlite.SQLiteManager
	typeValue        string
	goToPasswordList goToPasswordList
}

func NewPasswordCreatorModel(title string, sm *sqlite.SQLiteManager, goToPasswordList goToPasswordList) PasswordCreatorModel {
	accountName := textinput.New()
	accountName.Placeholder = "Account Name"
	accountName.Focus()
	accountName.Prompt = "Account Name: "
	accountName.CharLimit = 256

	username := textinput.New()
	username.Placeholder = "Username"
	username.Prompt = "Username: "
	username.CharLimit = 256

	password := textinput.New()
	password.Placeholder = "Password"
	password.Prompt = "Password: "
	password.CharLimit = 256
	password.EchoMode = textinput.EchoPassword
	return PasswordCreatorModel{
		title:            title,
		accountName:      accountName,
		username:         username,
		password:         password,
		err:              nil,
		focusIndex:       0,
		sqliteMan:        sm,
		typeValue:        "passwordcreator",
		goToPasswordList: goToPasswordList,
	}
}

func (m PasswordCreatorModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m PasswordCreatorModel) Update(msg tea.Msg) (models.DefaultModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab", "enter":
			if msg.String() == "tab" {
				m.focusIndex++
				if m.focusIndex > 2 {
					m.focusIndex = 0
				}
			} else {
				m.focusIndex--
				if m.focusIndex < 0 {
					m.focusIndex = 2
				}
			}

			cmds := make([]tea.Cmd, 3)
			cmds[0] = nil
			cmds[1] = nil
			cmds[2] = nil

			if msg.String() == "shift+tab" {
				m.focusIndex--

			}

			switch m.focusIndex {
			case 0:
				cmds[0] = m.accountName.Focus()
				m.password.Blur()
				m.username.Blur()
			case 1:
				cmds[1] = m.username.Focus()
				m.accountName.Blur()
				m.password.Blur()
			case 2:
				cmds[2] = m.password.Focus()
				m.accountName.Blur()
				m.username.Blur()
			}

			return m, tea.Batch(cmds...)
		case "ctrl+s":
			p := sqlite.Password{
				Place:    m.accountName.Value(),
				Username: m.username.Value(),
				Password: m.password.Value(),
			}
			err := m.sqliteMan.AddPassword(p)
			if err != nil {
				m.err = err
				return m, nil
			}
			return m, nil
		}
	case errorMsg:
		m.err = msg.error
		return m, nil

	// We handle errors just like any other message
	case errMsg:
		m.err = msg.error
		return m, nil
	}

	// Finally, handle the messages we pass to the text input models
	var cmd tea.Cmd
	cmds := make([]tea.Cmd, 4)
	m.accountName, cmd = m.accountName.Update(msg)
	cmds[0] = cmd
	m.username, cmd = m.username.Update(msg)
	cmds[1] = cmd
	m.password, cmd = m.password.Update(msg)
	cmds[2] = cmd

	return m, tea.Batch(cmds...)
}

func (m PasswordCreatorModel) View() string {
	return fmt.Sprintf(
		`%s

Account Name: %s
Username: %s
Password: %s

Press tab to move to the next field
Press ctrl+s to save
Press q to quit

%s
`,
		m.title,
		m.accountName.View(),
		m.username.View(),
		m.password.View(),
		func() string {
			if m.err != nil {
				return fmt.Sprintf("\nError: %s\n", m.err)
			}
			return ""
		}(),
	)
}

func (m PasswordCreatorModel) Title() string {
	return m.title
}

func (m PasswordCreatorModel) SetTitle(title string) {
	m.title = title
}

func (m PasswordCreatorModel) Type() string {
	return m.typeValue
}

func (m PasswordCreatorModel) SetType(t string) {
	m.typeValue = t
}

func (m PasswordCreatorModel) Value() string {
	return m.title
}

type errMsg struct{ error }

func (e errMsg) Error() string { return e.error.Error() }

type errorMsg struct{ error }
