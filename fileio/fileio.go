// Package fileio provides utilities for file and directory operations.
//
// This package offers helper functions for validating file paths and ensuring directories exist,
// designed to be used alongside other packages in the devify-utils library, such as csv and encryption.
// It includes a Serializer interface for data serialization and file I/O operations, along with
// standardized error types for common failure cases.
package fileio

import (
	"errors"
	"os"
	"path/filepath"
)

// Serializer defines an interface for data serialization and file I/O operations.
//
// Implementations of this interface provide methods for marshaling and unmarshaling data,
// as well as reading and writing files. This is used by other packages in devify-utils
// to handle data in specific formats, such as CSV or encrypted data.
//
// Methods:
//   - Marshal: Converts data to a byte slice.
//   - Unmarshal: Parses a byte slice into a destination object.
//   - ReadFile: Reads data from a file into a destination object.
//   - WriteFile: Writes data to a file with optional permissions.
type Serializer interface {
	Marshal(data any) ([]byte, error)
	Unmarshal(data []byte, dest any) error
	ReadFile(path string, dest any) error
	WriteFile(data any, path string, perm ...os.FileMode) error
}

// Errors defined for common file operation failures.
var (
	// ErrEmptyPath is returned when a file path is empty or refers to the root directory.
	ErrEmptyPath = errors.New("path cannot be empty or root")
	// ErrPathTooLong is returned when a file path exceeds the maximum length of 4096 characters.
	ErrPathTooLong = errors.New("path too long")
	// ErrFileNotExist is returned when a file does not exist at the specified path.
	ErrFileNotExist = errors.New("file does not exist")
	// ErrIsDir is returned when the specified path is a directory instead of a file.
	ErrIsDir = errors.New("path is a directory, not a file")
)

// ValidatePath checks if a file path is valid and has the expected file extension.
//
// The function ensures the path is not empty, not too long, exists as a file (not a directory),
// and has the specified file extension (e.g., ".csv"). It returns an error if any validation fails.
//
// Example:
//
//	err := ValidatePath("data.csv", ".csv")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("Path is valid")
//
// Parameters:
//   - path: The file path to validate.
//   - ext: The expected file extension (e.g., ".csv").
//
// Returns:
//   - error: An error if the path is empty, too long, does not exist, is a directory,
//     or does not have the specified extension.
func ValidatePath(path string, ext string) error {
	if path == "" || path == "." {
		return ErrEmptyPath
	}
	if len(path) > 4096 {
		return ErrPathTooLong
	}
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrFileNotExist
		}
		return err
	}
	if info.IsDir() {
		return ErrIsDir
	}
	if filepath.Ext(path) != ext {
		return errors.New("file must have " + ext + " extension")
	}
	return nil
}

// EnsureDir creates all parent directories for a given file path if they do not exist.
//
// The function uses the specified permission mode for creating directories. If the path's
// parent directory is the current directory ("."), no action is taken, and nil is returned.
// This is useful for ensuring a file can be written to the specified path without directory-related errors.
//
// Example:
//
//	err := EnsureDir("data/output.csv", 0o755)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("Directories created")
//
// Parameters:
//   - path: The file path whose parent directories should be created.
//   - perm: The permission mode for created directories (e.g., 0o755).
//
// Returns:
//   - error: An error if directory creation fails.
func EnsureDir(path string, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if dir != "." {
		return os.MkdirAll(dir, perm)
	}
	return nil
}
