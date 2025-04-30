package sqliteManager

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteManager struct {
	db         *sql.DB
	dbFileName string
	NewDb      bool
}

type Password struct {
	ID int
	Place    string
	Username string
	Password string
}

type PasswordPlace struct {
	ID    int
	Place string
}

func NewSQLiteManager(dbFileName string) (*SQLiteManager, error) {
	sqliteManager := &SQLiteManager{
		dbFileName: dbFileName,
		NewDb:      false,
	}

	if !SqliteFileExists(dbFileName) {
		err := sqliteManager.CreateSQLiteFile(dbFileName)
		if err != nil {
			return nil, fmt.Errorf("failed to create SQLite file: %w", err)
		}
		sqliteManager.NewDb = true
	}

	// Open a connection to Sqlite db
	db, err := sql.Open("sqlite3", dbFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	sqliteManager.db = db

	if err := sqliteManager.SetupDatabase(); err != nil {
		return nil, fmt.Errorf("failed to setup database: %w", err)
	}

	exists, err := sqliteManager.MasterPasswordExists()
	if err != nil {
		return nil, fmt.Errorf("failed to check if master password exists: %w", err)
	}
	if !exists {
		sqliteManager.NewDb = true
	}

	return sqliteManager, nil
}

func (sm *SQLiteManager) CreateSQLiteFile(dbFileName string) error {
	// Check if the file already exists and remove if it does
	if _, err := os.Stat(sm.dbFileName); err == nil {
		err := os.Remove(sm.dbFileName)
		if err != nil {
			return fmt.Errorf("failed to remove existing database file: %w", err)
		}
		fmt.Println("Existing database file removed.")
	}

	// Create the SQLite file.
	file, err := os.Create(dbFileName)
	if err != nil {
		return fmt.Errorf("failed to create database file: %w", err)
	}
	file.Close()

	fmt.Printf("SQLite file '%s' created successfully.\n", dbFileName)
	return nil
}

func (sm *SQLiteManager) SetupDatabase() error {
	// Create the 'passwords' table.
	_, err := sm.db.Exec(`
		CREATE TABLE IF NOT EXISTS passwords (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			place TEXT NOT NULL,
			username TEXT NOT NULL,
			password TEXT NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create passwords table: %w", err)
	}

	fmt.Println("Database tables created successfully.")
	return nil
}

func (sm *SQLiteManager) AddPassword(password Password) error {
	if password.Place == "master" {
		return fmt.Errorf("cannot add or override master password with this function")
	}

	_, err := sm.db.Exec(`
		Insert INTO passwords (place, username, password) VALUES (?, ?, ?)
		`, password.Place, password.Username, password.Password)
	if err != nil {
		return fmt.Errorf("Failed to insert password: %w", err)
	}
	fmt.Printf("Password for '%s' added successfully.\n", password.Place)
	return nil
}

func (sm *SQLiteManager) Close() error {
	return sm.db.Close()
}

// SetMasterPassword adds the master password only if one doesn't already exist.
func (sm *SQLiteManager) SetMasterPassword(password Password) error {
	// Check if the master password already exists.
	exists, err := sm.MasterPasswordExists()
	if err != nil {
		return err
	}

	if exists {
		return fmt.Errorf("master password already set; cannot override")
	}

	if len(password.Password) < 8 {
		return fmt.Errorf("Password must be at least 8 characters long")
	}

	if password.Place != "master" {
		return fmt.Errorf("you must add a master password with this function")
	}

	// Add the master password to the passwords table.
	_, err = sm.db.Exec(`
		INSERT INTO passwords (place, username, password) VALUES (?, ?, ?)
	`, "master", "master", password.Password)
	if err != nil {
		return fmt.Errorf("failed to set master password: %w", err)
	}
	fmt.Println("Master password set successfully.")
	return nil
}

// masterPasswordExists checks if a master password already exists.
func (sm *SQLiteManager) MasterPasswordExists() (bool, error) {
	var count int
	err := sm.db.QueryRow(`SELECT COUNT(*) FROM passwords WHERE place = ?`, "master").Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check if master password exists: %w", err)
	}
	return count > 0, nil
}

func (sm *SQLiteManager) CheckMasterPassword(password string) (bool, error) {
	var storedPassword string
	err := sm.db.QueryRow(`
		SELECT password FROM passwords WHERE place = ?
		`, "master").Scan(&storedPassword)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("Failed to query master password: %w", err)
	}

	return password == storedPassword, nil
}

func (sm *SQLiteManager) GetAllPasswordPlaces() ([]PasswordPlace, error) {
	rows, err := sm.db.Query(`SELECT id as ID, place as Place FROM passwords WHERE place != ?`, "master")
	if err != nil {
		return nil, fmt.Errorf("failed to query password places: %w", err)
	}
	defer rows.Close()

	var passwords []PasswordPlace
	for rows.Next() {
		var password struct {
			id int
			place string
		}
		if err := rows.Scan(&password.id, &password.place); err != nil {
			return nil, fmt.Errorf("failed to scan password place row: %w", err)
		}
		passwords = append(passwords, PasswordPlace{ID: password.id, Place: password.place})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over password place rows: %w", err)
	}

	return passwords, nil
}

func (sm *SQLiteManager) GetPasswordByIndex(index int) (Password, error) {
	var password Password
	err := sm.db.QueryRow(`SELECT place, username, password FROM passwords WHERE id = ?`, index).Scan(&password.Place, &password.Username, &password.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return Password{}, fmt.Errorf("no password found for the given index")
		}
		return Password{}, fmt.Errorf("failed to query password by index: %w", err)
	}

	return password, nil
}

