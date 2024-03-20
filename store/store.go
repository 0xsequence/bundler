package store

import (
	"fmt"
	"os"
	"path/filepath"
)

type Store string

func CreateInstanceStore(instance string) (Store, error) {
	topLevelDir := ".bundler_5189"

	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting home directory: %w", err)
	}

	// Construct the full path for the instance directory
	instanceDirPath := filepath.Join(homeDir, topLevelDir, instance)

	// Ensure the directory exists (create it along with any necessary parents)
	err = os.MkdirAll(instanceDirPath, 0700) // Using 0700 permissions for user privacy
	if err != nil {
		return "", fmt.Errorf("error creating instance directory: %w", err)
	}

	// Return the full path to the instance directory
	return Store(instanceDirPath), nil
}

func (s Store) ReadFile(name string) (string, error) {
	if s == "" {
		return "", fmt.Errorf("store not initialized")
	}

	filePath := filepath.Join(string(s), name)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	return string(data), nil
}

func (s Store) WriteFile(name string, data string) error {
	if s == "" {
		return fmt.Errorf("store not initialized")
	}

	filePath := filepath.Join(string(s), name)
	err := os.WriteFile(filePath, []byte(data), 0600)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}
	return nil
}

func (s Store) String() string {
	return string(s)
}
