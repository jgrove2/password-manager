package passwordviewer

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"password-manager/internal/models"
	sqlite "password-manager/internal/sqlite"
)

type goToPasswordList func()

type PasswordViewerModel struct {
	title            string
	password         sqlite.Password
	sqliteManager    *sqlite.SQLiteManager
	err              error
	width            int
	height           int
	loaded           bool
	typeValue        string
	goToPasswordList goToPasswordList
}

func NewPasswordViewerModel(title string, sm *sqlite.SQLiteManager, id int, goToPasswordList goToPasswordList) PasswordViewerModel {
	model := &PasswordViewerModel{
		title:            title,
		password:         sqlite.Password{},
		sqliteManager:    sm,
		err:              nil,
		width:            0,
		height:           0,
		loaded:           false,
		typeValue:        "passwordviewer",
		goToPasswordList: goToPasswordList,
	}
	password, err := model.sqliteManager.GetPasswordByIndex(id)
	if err != nil {
		model.err = err
		return *model
	}
	model.password = password
	log.Printf("PasswordViewerModel: %v", model.password)
	model.loaded = true
	return *model
}

func (m PasswordViewerModel) Init() tea.Cmd {
	return func() tea.Msg {
		password, err := m.sqliteManager.GetPasswordByIndex(m.password.ID)
		if err != nil {
			return errMsg{err}
		}
		return &sqlite.Password{
			ID:       password.ID,
			Place:    password.Place,
			Username: password.Username,
			Password: password.Password,
		}
	}
}

func (m PasswordViewerModel) Update(msg tea.Msg) (models.DefaultModel, tea.Cmd) {
	switch msg := msg.(type) {
	case errMsg:
		m.err = msg.error
		return m, nil
	}
	return m, nil
}

func (m PasswordViewerModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("\nError: %s\n\nPress any key to continue...\n", m.err)
	}

	if !m.loaded {
		return "Loading..."
	}

	s := fmt.Sprintf("%s\n\n", m.password.Place)
	s += fmt.Sprintf("Username: %s\n", m.password.Username)
	s += fmt.Sprintf("Password: %s\n", m.password.Password)
	s += "Press 'q' to quit"

	return s
}

func (m PasswordViewerModel) Title() string {
	return m.title
}

func (m PasswordViewerModel) SetTitle(title string) {
	m.title = title
}

func (m PasswordViewerModel) Type() string {
	return m.typeValue
}

func (m PasswordViewerModel) SetType(t string) {
	m.typeValue = t
}

func (m PasswordViewerModel) Value() string {
	return m.password.Place
}

type errMsg struct{ error }

func (e errMsg) Error() string { return e.error.Error() }

type passwordMsg struct {
	ID    int
	Place string
}

