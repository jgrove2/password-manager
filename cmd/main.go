package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"password-manager/internal/input"
	"password-manager/internal/models"
	"password-manager/internal/passwordcreator"
	"password-manager/internal/passwordlist"
	"password-manager/internal/passwordviewer"
	sqlite "password-manager/internal/sqlite"
)

type Styles struct {
	BorderColor lipgloss.Color
	InputField  lipgloss.Style
}

func DefaultStyles() *Styles {
	s := new(Styles)
	s.BorderColor = lipgloss.Color("36")
	s.InputField = lipgloss.NewStyle().
		BorderForeground(s.BorderColor).
		BorderStyle(
			lipgloss.NormalBorder(),
		).
		Padding(1).
		Width(40)
	return s
}

type model struct {
	states           []models.DefaultModel
	width            int
	height           int
	index            int
	styles           *Styles
	done             bool
	password         string
	validPassword    bool
	confirm          string
	newManager       bool
	sm               *sqlite.SQLiteManager
	selectedPassword int
	error            error
	passwords        []sqlite.PasswordPlace
}

type States struct {
	model models.DefaultModel
}

func New() *model {
	styles := DefaultStyles()
	return &model{
		states:           []models.DefaultModel{},
		styles:           styles,
		newManager:       false,
		validPassword:    false,
		selectedPassword: 0,
		passwords:        []sqlite.PasswordPlace{},
	}

}

func (m *model) Init() tea.Cmd {
	log.Println("Initializing Model")
	sm, err := sqlite.NewSQLiteManager("manager.db")
	if err != nil {
		log.Printf("Error initializing SqliteManager: %v", err)
		m.done = true
		m.error = err
		return nil
	}
	m.sm = sm

	// If database is brand new, Setup the application so the user can setup the master
	if m.sm.NewDb {
		log.Println("New DB is created, going to setup")
		m.states = []models.DefaultModel{passwordInput(), confirmPassword(), m.changetoPasswordList(), m.changetoPasswordCreator(), m.changetoPasswordViewer(m.selectedPassword)}
	} else {
		//If Master exists then just take it
		passwords, err := m.sm.GetAllPasswordPlaces()
		log.Printf("%v, %v", m.passwords, passwords)
		if err != nil {
			return nil
		}
		m.passwords = passwords
		log.Println("Master password found, taking login")
		m.states = []models.DefaultModel{passwordInput(), m.changetoPasswordList(), m.changetoPasswordCreator(), m.changetoPasswordViewer(m.selectedPassword)}
	}
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			log.Println("Quitting out of program")
			return m, tea.Quit
		}

		// Handle initial state for password input and confirmation
		if len(m.states) > 0 {
			current := m.states[m.index]
			switch msg.String() {
			case "enter":
				log.Println("Enter key pressed")
				if current.Type() == "passwordInput" {
					m.password = current.Value()
					if !m.sm.NewDb {
						correctPassword, err := m.sm.CheckMasterPassword(m.password)
						if err != nil {
							log.Printf("Error checking master password: %v", err)
							m.done = true
							m.error = err
							return m, nil
						}
						if !correctPassword {
							log.Println("Incorrect password")
							m.states[m.index].SetTitle("The password is incorrect!")
							m.password = ""
							m.validPassword = false
							return m, nil
						}
						log.Println("Correct password, showing password list")
						m.Next()
						return m, nil
					} else {
						log.Println("New DB, going to confirm password")
						m.states[m.index].SetTitle("Please confirm your password!")
						m.index = 1
						m.states[m.index] = confirmPassword()
						return m, nil
					}
					m.Next()
				} else if current.Type() == "confirmPassword" {
					log.Println("Confirm password step")
					m.confirm = current.Value()
					if m.password == "" {
						log.Println("Password is empty")
						m.states[0].SetTitle("Please enter a password first!")
						return m, nil
					}
					if m.confirm == "" {
						log.Println("Confirm password is empty")
						m.states[m.index].SetTitle("Please confirm your password!")
						return m, nil
					}
					if !m.validateConfirmPassword() {
						log.Println("Passwords do not match")
						m.index = 0
						m.password = ""
						m.confirm = ""
						m.states[m.index].SetTitle("Your password did not match!")
						return m, nil
					}
					log.Println("Passwords match, setting master password")
					p := sqlite.Password{
						Place:    "master",
						Username: "master",
						Password: m.password,
					}
					if err := m.sm.SetMasterPassword(p); err != nil {
						log.Printf("Error setting master password: %v", err)
						m.done = true
						m.error = err
						return m, nil
					}
					// After setting master password, show the password list
					m.Next()
					return m, nil
				} else if current.Type() == "passwordlist" {
					for _, password := range m.passwords {
						if password.Place == current.Value() {
							m.selectedPassword = password.ID
						}
					}
					log.Printf("Selected password: %v", m.selectedPassword)
					m.index = m.index + 2
					m.states[m.index] = m.changetoPasswordViewer(m.selectedPassword)
					log.Println("Going to password viewer")
					return m, nil
				}
			case "c":
				if current.Type() == "passwordlist" {
					log.Println("Creating new password")
					m.Next()
					return m, nil
				}
			case "q":
				if current.Type() == "passwordcreator" {
					m.index = m.index - 1
					passwords, err := m.sm.GetAllPasswordPlaces()
					log.Printf("%v, %v", m.passwords, passwords)
					if err != nil {
						return m, nil
					}
					m.passwords = passwords
					m.states[m.index] = m.changetoPasswordList()
					log.Println("Going back to password list")
					return m, nil
				} else if current.Type() == "passwordviewer" {
					m.index = m.index - 2
					passwords, err := m.sm.GetAllPasswordPlaces()
					log.Printf("%v, %v", m.passwords, passwords)
					if err != nil {
						return m, nil
					}
					m.passwords = passwords
					m.states[m.index] = m.changetoPasswordList()
					log.Println("Going back to password list")
					return m, nil
				}
			}

			// Update current input state
			var cmd tea.Cmd
			m.states[m.index], cmd = current.Update(msg)
			return m, cmd
		}
	}
	return m, nil
}

func (m *model) View() string {
	if m.done {
		if m.validPassword {
			return "Thank you for using this terminal password manager!"
		} else if m.error != nil {
			return fmt.Sprintf("Error: %v", m.error)
		} else {
			return "Exiting application"
		}
	}

	if m.width == 0 {
		return "loading..."
	}

	if len(m.states) == 0 {
		return "loading..."
	}

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		lipgloss.JoinVertical(
			lipgloss.Center,
			m.states[m.index].Title(),
			m.styles.InputField.Render(m.states[m.index].View()),
		),
	)
}

func (m *model) Next() {
	if m.index < len(m.states)-1 {
		m.index++
	} else {
		m.done = true
	}
}

func (m *model) validateConfirmPassword() bool {
	return m.password == m.confirm
}

func passwordInput() models.DefaultModel {
	passwordInput := input.NewInputField("Password")
	passwordInput.SetTitle("Your Password")
	passwordInput.SetType("passwordInput")
	return passwordInput
}

func confirmPassword() models.DefaultModel {
	passwordInput := input.NewInputField("Confirm Password")
	passwordInput.SetTitle("Please confirm your password")
	passwordInput.SetType("confirmPassword")
	return passwordInput
}

func passwordList(passwords []sqlite.Password) models.DefaultModel {
	passwordList := input.NewInputField("Password List")
	passwordList.SetTitle("Your Password List")
	passwordList.SetType("passwordList")
	return passwordList
}

func (m *model) changetoPasswordList() models.DefaultModel {
	log.Printf("%v", m.passwords)
	model := passwordlist.NewPasswordListModel("Password List", m.sm, m.goToPasswordViewer, m.goToPasswordChanger, m.passwords)
	return model
}

func (m *model) changetoPasswordViewer(id int) models.DefaultModel {
	model := passwordviewer.NewPasswordViewerModel("Password Viewer", m.sm, id, m.goToPasswordList)
	return model
}

func (m *model) changetoPasswordCreator() models.DefaultModel {
	model := passwordcreator.NewPasswordCreatorModel("Password Creator", m.sm, m.goToPasswordList)
	return model
}

func (m *model) goToPasswordList() {
	log.Println("Going to password list")
	m.index = len(m.states) - 3
}

func (m *model) goToPasswordViewer(id int) {
	log.Println("Going to password list")
	m.selectedPassword = id
	m.index = len(m.states) - 2
}

func (m *model) goToPasswordChanger() {
	log.Println("Going to password list")
	m.index = len(m.states) - 1
}

func main() {
	logFile, err := os.OpenFile("manager.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening log file: %v\n", err)
		os.Exit(1)
	}

	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	m := New()

	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
		logFile.Close()
	}
}
