package utils

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
)

func CreateSha256Checksum(input []byte) (hash string) {
	h := sha256.New()
	h.Write(input)
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

func EnsureDBPathExists(dbPath string) error {
	// Get the directory of the database path
	dir := filepath.Dir(dbPath)

	// Check if the directory exists
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		// Directory does not exist, create it
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}
	} else if err != nil {
		// An error occurred while checking the directory
		return fmt.Errorf("failed to check directory: %v", err)
	}

	// Create the file if it doesn't exist
	_, err = os.Stat(dbPath)
	if os.IsNotExist(err) {
		// File does not exist, create it
		file, err := os.Create(dbPath)
		if err != nil {
			return fmt.Errorf("failed to create file: %v", err)
		}
		defer file.Close()
	} else if err != nil {
		// An error occurred while checking the file
		return fmt.Errorf("failed to check file: %v", err)
	}

	return nil
}
