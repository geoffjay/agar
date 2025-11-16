package commands

import (
	"os"
)

// readFile reads a file and returns its contents
func readFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

// writeFile writes data to a file
func writeFile(filename string, data []byte) error {
	return os.WriteFile(filename, data, 0644)
}

// fileExists checks if a file exists
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}
