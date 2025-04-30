package passwordlist

import (
	"fmt"
	"strings"
	"log"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"password-manager/internal/models"
	sqlite "password-manager/internal/sqlite"
)

type goToViewer func(int)
type goToCreator func()

type PasswordListModel struct {
	title         string
	passwords     []sqlite.PasswordPlace
	selected      int
	sqliteManager *sqlite.SQLiteManager
	textInput     textinput.Model
	typing        bool
	err           error
	width         int
	height        int
	loaded        bool
	typeValue     string
	goToViewer goToViewer
	goToCreator goToCreator
}


func NewPasswordListModel(title string, sm *sqlite.SQLiteManager, goToViewer goToViewer, goToCreator goToCreator, passwords []sqlite.PasswordPlace) PasswordListModel {
	ti := textinput.New()
	ti.Placeholder = "Search"
	ti.Focus()
	ti.Width = 20

	return PasswordListModel{
		title:         title,
		passwords:     passwords,
		selected:      0,
		sqliteManager: sm,
		textInput:     ti,
		typing:        false,
		err:           nil,
		width:         0,
		height:        0,
		loaded:        false,
		typeValue:     "passwordlist",
		goToViewer:    goToViewer,
		goToCreator:   goToCreator,
	}
}

func (m PasswordListModel) Init() tea.Cmd {
	return func() tea.Msg {
		passwords, err := m.sqliteManager.GetAllPasswordPlaces()
		log.Println("Loading passwords")
		if err != nil {
			return errMsg{err}
		}
		return passwords
	}
}

func (m PasswordListModel) Update(msg tea.Msg) (models.DefaultModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if m.selected > 0 {
				m.selected--
			}
		case "down":
			if m.selected < len(m.passwords)-1 {
				m.selected++
			}
		case "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m PasswordListModel) View() string {
	return m.createListView()
}

func (m PasswordListModel) Title() string {
	return m.title
}

func (m PasswordListModel) SetTitle(title string) {
	m.title = title
}

func (m PasswordListModel) Type() string {
	return m.typeValue
}

func (m PasswordListModel) SetType(t string) {
	m.typeValue = t
}

func (m PasswordListModel) Value() string {
	if len(m.passwords) > 0 {
		return m.passwords[m.selected].Place
	}
	return ""
}

func (m PasswordListModel) createListView() string {
	if len(m.passwords) == 0 {
		return fmt.Sprintf("%s\n\nNo passwords stored in the database.\n\nPress 'c' to create a new password\nPress 'q' to quit", m.title)
	}
	log.Printf("test: %v", m.passwords)

	var s strings.Builder
	s.WriteString(fmt.Sprintf("%s\n\n", m.title))

	// Show navigation instructions if we have passwords
	s.WriteString("Use up/down arrows to navigate\n")
	s.WriteString("Press Enter to view password details\n\n")

	for i, password := range m.passwords {
		cursor := " "
		if i == m.selected {
			cursor = ">"
		}
		s.WriteString(fmt.Sprintf("%s %s\n", cursor, password.Place))
	}

	s.WriteString("\nPress 'c' to create a new password\n")
	s.WriteString("Press 'q' to quit")

	return s.String()
}

type errMsg struct{ error }

func (e errMsg) Error() string { return e.error.Error() }

type passwordsMsg []sqlite.Password