// Package yaml provides utilities for YAML serialization and file operations.
//
// This package offers functions for marshaling and unmarshaling YAML data, as well as reading and writing YAML files.
// It integrates with the fileio package from devify-utils for path validation and directory creation, and uses
// gopkg.in/yaml.v3 for YAML processing. All functions include error handling for common cases, such as empty data,
// invalid file paths, or parsing errors. Both .yaml and .yml file extensions are supported.
package yaml

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/devify-me/devify-utils/fileio"
	yamlv3 "gopkg.in/yaml.v3"
)

// Marshal serializes the given data to YAML format as a byte slice.
//
// The function checks that the input data is not nil and marshals it to YAML using gopkg.in/yaml.v3.
// It handles potential panics during marshaling by recovering and converting them to errors.
// If serialization fails or the input is nil, an error is returned.
//
// Example:
//
//	data := map[string]string{"name": "Alice", "role": "admin"}
//	output, err := Marshal(data)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(string(output)) // Prints YAML, e.g., "name: Alice\nrole: admin\n"
//
// Parameters:
//   - data: The data to serialize to YAML (e.g., structs, maps, or other types supported by gopkg.in/yaml.v3).
//
// Returns:
//   - []byte: The YAML-encoded data as a byte slice.
//   - error: An error if the data is nil or cannot be marshaled.
func Marshal(data any) ([]byte, error) {
	if data == nil {
		return nil, errors.New("data cannot be nil")
	}
	var output []byte
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("%v", r)
			}
		}()
		output, err = yamlv3.Marshal(data)
	}()
	if err != nil {
		return nil, err
	}
	return output, nil
}

// Unmarshal parses YAML data into the provided destination.
//
// The destination must be a non-nil pointer to a struct, map, or other type supported by gopkg.in/yaml.v3.
// The function checks that the input data is not empty and that the destination is not nil.
// If parsing fails, an error is returned.
//
// Example:
//
//	var result map[string]string
//	data := []byte("name: Alice\nrole: admin")
//	err := Unmarshal(data, &result)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(result) // Prints map[name:Alice role:admin]
//
// Parameters:
//   - data: The YAML-encoded data as a byte slice.
//   - dest: A pointer to the destination where the parsed YAML data will be stored.
//
// Returns:
//   - error: An error if the data is empty, the destination is nil, or parsing fails.
func Unmarshal(data []byte, dest any) error {
	if len(data) == 0 {
		return errors.New("YAML data cannot be empty")
	}
	if dest == nil {
		return errors.New("destination cannot be nil")
	}
	return yamlv3.Unmarshal(data, dest)
}

// ReadFile reads a YAML file from the specified path and unmarshals it into the provided destination.
//
// The function validates that the file path has a ".yaml" or ".yml" extension, is not empty or root,
// and does not exceed 4096 characters. It uses fileio.ValidateReadPath to ensure the file exists and is not a directory.
// The file is checked for non-empty content before unmarshaling into the destination, which must be a non-nil
// pointer to a struct, map, or other type supported by gopkg.in/yaml.v3.
//
// Example:
//
//	var result map[string]string
//	err := ReadFile("config.yaml", &result)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(result) // Prints the parsed YAML data, e.g., map[name:Alice role:admin]
//
// Parameters:
//   - path: The file path of the YAML file to read (must have .yaml or .yml extension).
//   - dest: A pointer to the destination where the parsed YAML data will be stored.
//
// Returns:
//   - error: An error if the path is invalid, the file is empty, the destination is nil, or unmarshaling fails.
func ReadFile(path string, dest any) error {
	if path == "" || path == "." {
		return errors.New("path cannot be empty or root")
	}
	if len(path) > 4096 {
		return errors.New("path too long")
	}
	if dest == nil {
		return errors.New("destination cannot be nil")
	}
	ext := filepath.Ext(path)
	if err := fileio.ValidateReadPath(path, ext); err != nil {
		return err
	}
	if ext != ".yaml" && ext != ".yml" {
		return errors.New("file must have .yaml or .yml extension")
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

// WriteFile serializes the given data to YAML and writes it to a file at the specified path.
//
// The function validates that the file path has a ".yaml" or ".yml" extension, is not empty or root,
// and does not exceed 4096 characters. It uses fileio.ValidateWritePath to ensure the path is valid for writing.
// The data is marshaled to YAML, and parent directories are created using fileio.EnsureDir if needed.
// The resulting bytes are written to the file with the specified permissions (defaulting to 0600 if not provided).
// If the data cannot be marshaled or the file cannot be written, an error is returned.
//
// Example:
//
//	data := map[string]string{"name": "Alice", "role": "admin"}
//	err := WriteFile(data, "config.yaml", 0o644)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Parameters:
//   - data: The data to serialize to YAML (e.g., structs, maps, or other types supported by gopkg.in/yaml.v3).
//   - path: The file path where the YAML data will be written (must have .yaml or .yml extension).
//   - perm: Optional file permission mode (os.FileMode). Defaults to 0600 if not provided.
//
// Returns:
//   - error: An error if the path is invalid, data cannot be marshaled, directories cannot be created,
//     or the file cannot be written.
func WriteFile(data any, path string, perm ...os.FileMode) error {
	if path == "" || path == "." {
		return errors.New("path cannot be empty or root")
	}
	if len(path) > 4096 {
		return errors.New("path too long")
	}
	ext := filepath.Ext(path)
	if ext != ".yaml" && ext != ".yml" {
		return errors.New("file must have .yaml or .yml extension")
	}
	if err := fileio.ValidateWritePath(path, ext); err != nil {
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
