package json

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// Marshal serializes the given data to JSON format.
// It returns an error if the data cannot be marshaled or if the output is empty.
func Marshal(data any) ([]byte, error) {
	if data == nil {
		return nil, errors.New("data cannot be nil")
	}
	output, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	if len(output) <= 2 { // Check for "{}" or similar minimal output
		return nil, errors.New("marshaled JSON is empty")
	}
	return output, nil
}

// Unmarshal parses JSON data into the given destination struct.
// It returns an error if the input is empty or cannot be unmarshaled.
func Unmarshal(data []byte, dest any) error {
	if len(data) == 0 {
		return errors.New("JSON data cannot be empty")
	}
	if dest == nil {
		return errors.New("destination cannot be nil")
	}
	return json.Unmarshal(data, dest)
}

// ReadFile reads a JSON file from the given path and unmarshals it into the destination struct.
// It validates the file path and ensures it exists and is not a directory.
func ReadFile(path string, dest any) error {
	if path == "" || path == "." {
		return errors.New("path cannot be empty or root")
	}
	if len(path) > 4096 {
		return errors.New("path too long")
	}

	// Check if the path exists and is not a directory
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New("file does not exist")
		}
		return err
	}
	if info.IsDir() {
		return errors.New("path is a directory, not a file")
	}

	// Ensure the file has a .json extension
	ext := filepath.Ext(path)
	if ext != ".json" {
		return errors.New("file must have .json extension")
	}

	// Read the file content
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return errors.New("file is empty")
	}

	// Unmarshal the JSON content
	return Unmarshal(data, dest)
}

// WriteFile marshals the given data to JSON and writes it to the specified file path.
// It validates the directory path and ensures the file has a valid JSON extension.
func WriteFile(data any, path string, perm ...os.FileMode) error {
	if path == "" || path == "." {
		return errors.New("path cannot be empty or root")
	}
	if len(path) > 4096 {
		return errors.New("path too long")
	}

	// Ensure the file has a .json extension
	ext := filepath.Ext(path)
	if ext != ".json" {
		return errors.New("file must have .json extension")
	}

	// Marshal the data to JSON
	output, err := Marshal(data)
	if err != nil {
		return err
	}

	// Ensure the parent directory exists
	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	// Set default file permissions
	fileMode := os.FileMode(0o600)
	if len(perm) > 0 {
		fileMode = perm[0]
	}

	// Write the JSON content to the file
	return os.WriteFile(path, output, fileMode)
}
