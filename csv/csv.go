package csv

import (
	"encoding/csv"
	"errors"
	"os"
	"path/filepath"
)

// ReadFile reads a CSV file and returns its records.
func ReadFile(path string) ([][]string, error) {
	if path == "" || path == "." {
		return nil, errors.New("path cannot be empty or root")
	}
	if len(path) > 4096 {
		return nil, errors.New("path too long")
	}

	// Check file existence and type
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("file does not exist")
		}
		return nil, err
	}
	if info.IsDir() {
		return nil, errors.New("path is a directory, not a file")
	}

	// Ensure .csv extension
	ext := filepath.Ext(path)
	if ext != ".csv" {
		return nil, errors.New("file must have .csv extension")
	}

	// Read file
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, errors.New("file is empty")
	}

	return records, nil
}

// WriteFile writes records to a CSV file.
func WriteFile(records [][]string, path string, perm ...os.FileMode) error {
	if path == "" || path == "." {
		return errors.New("path cannot be empty or root")
	}
	if len(path) > 4096 {
		return errors.New("path too long")
	}

	// Ensure .csv extension
	ext := filepath.Ext(path)
	if ext != ".csv" {
		return errors.New("file must have .csv extension")
	}

	if len(records) == 0 {
		return errors.New("records cannot be empty")
	}

	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	// Set default permissions
	fileMode := os.FileMode(0o600)
	if len(perm) > 0 {
		fileMode = perm[0]
	}

	// Write file with specified permissions
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fileMode)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	err = writer.WriteAll(records)
	if err != nil {
		return err
	}
	writer.Flush()
	return writer.Error()
}
