package sqliteManager

import (
	"fmt"
	"os"
	"path/filepath"
)

func SqliteFileExists(filename string) bool {
	// Get the current working directory
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current working directory:", err)
		return false
	}

	filePath := filepath.Join(dir, filename)

	// Check if the file exists
	_, err = os.Stat(filePath)
	if err == nil {
		// File exists
		return true
	}
	if os.IsNotExist(err) {
		// File does not exist
		return false
	}

	fmt.Println("Error checking file existence:", err)
	return false
}
