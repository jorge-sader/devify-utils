// Package json provides utilities for JSON serialization and file operations.
//
// This package offers functions for marshaling and unmarshaling JSON data, as well as reading and writing JSON files.
// It integrates with the fileio package from devify-utils for path validation and directory creation.
// All functions include error handling for common cases, such as empty data or invalid file paths.
package json

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/devify-me/devify-utils/fileio"
)

// Marshal serializes the given data to JSON format as a byte slice.
//
// The function checks that the input data is not nil and that the marshaled output is not empty
// (i.e., not "{}" or "[]"). If serialization fails or the output is empty, an error is returned.
//
// Example:
//
//	data := map[string]string{"key": "value"}
//	output, err := Marshal(data)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(string(output)) // Prints `{"key":"value"}`
//
// Parameters:
//   - data: The data to serialize to JSON (can be any type supported by encoding/json).
//
// Returns:
//   - []byte: The JSON-encoded data as a byte slice.
//   - error: An error if the data is nil, cannot be marshaled, or results in empty JSON.
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

// Unmarshal parses JSON data into the provided destination.
//
// The destination must be a non-nil pointer to a struct, map, or other type supported by encoding/json.
// The function checks that the input data is not empty and that the destination is not nil.
// If parsing fails, an error is returned.
//
// Example:
//
//	var result map[string]string
//	data := []byte(`{"key":"value"}`)
//	err := Unmarshal(data, &result)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(result) // Prints map[key:value]
//
// Parameters:
//   - data: The JSON-encoded data as a byte slice.
//   - dest: A pointer to the destination where the parsed data will be stored.
//
// Returns:
//   - error: An error if the data is empty, the destination is nil, or parsing fails.
func Unmarshal(data []byte, dest any) error {
	if len(data) == 0 {
		return errors.New("JSON data cannot be empty")
	}
	if dest == nil {
		return errors.New("destination cannot be nil")
	}
	return json.Unmarshal(data, dest)
}

// ReadFile reads a JSON file from the specified path and unmarshals it into the provided destination.
//
// The function validates that the file path has a ".json" extension and exists using fileio.ValidatePath.
// It also checks that the file is not empty before attempting to unmarshal the data into the destination,
// which must be a non-nil pointer to a struct, map, or other type supported by encoding/json.
//
// Example:
//
//	var result map[string]string
//	err := ReadFile("config.json", &result)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(result) // Prints the parsed JSON data
//
// Parameters:
//   - path: The file path of the JSON file to read.
//   - dest: A pointer to the destination where the parsed JSON data will be stored.
//
// Returns:
//   - error: An error if the path is invalid, the file is empty, or unmarshaling fails.
func ReadFile(path string, dest any) error {
	if err := fileio.ValidateReadPath(path, ".json"); err != nil {
		return err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return errors.New("file is empty")
	}
	return Unmarshal(data, dest)
}

// WriteFile serializes the given data to JSON and writes it to a file at the specified path.
//
// The function validates that the file path has a ".json" extension using fileio.ValidatePath and ensures
// the parent directories exist using fileio.EnsureDir. The data is marshaled to JSON, and the resulting
// bytes are written to the file with the specified permissions (defaulting to 0600 if not provided).
// If the data cannot be marshaled or the file cannot be written, an error is returned.
//
// Example:
//
//	data := map[string]string{"key": "value"}
//	err := WriteFile(data, "config.json", 0o644)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Parameters:
//   - data: The data to serialize and write to the file (can be any type supported by encoding/json).
//   - path: The file path where the JSON data will be written.
//   - perm: Optional file permission mode (os.FileMode). Defaults to 0600 if not provided.
//
// Returns:
//   - error: An error if the path is invalid, data cannot be marshaled, directories cannot be created,
//     or the file cannot be written.
func WriteFile(data any, path string, perm ...os.FileMode) error {
	if err := fileio.ValidateWritePath(path, ".json"); err != nil {
		return err
	}
	output, err := Marshal(data)
	if err != nil {
		return err
	}
	if err := fileio.EnsureDir(path, 0o755); err != nil {
		return err
	}
	fileMode := os.FileMode(0o600)
	if len(perm) > 0 {
		fileMode = perm[0]
	}
	return os.WriteFile(path, output, fileMode)
}
