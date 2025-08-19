// Package filesystem provides utilities for file and directory operations, MIME type detection, and filename sanitization.
//
// This package offers helper functions for managing files and directories, checking file existence,
// appending content to files, creating files and directories, and sanitizing filenames for cross-platform compatibility.
// It also includes functions for determining MIME types based on file extensions or content.
// These utilities are designed to be used within the devify-utils library to support robust file operations.
package filesystem

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"unicode"
)

// CaseStyle defines the style of the filename case.
//
// This type is reserved for future use in functions that may adjust filename case formatting
// (e.g., camelCase, snake_case). Currently, it is unused in the package.
type CaseStyle string

// FileExists checks if a file or directory exists at the specified path.
//
// The function uses os.Stat to determine if the path exists, returning true if it does and false otherwise.
// It does not distinguish between files and directories.
//
// Example:
//
//	if FileExists("data.txt") {
//	    fmt.Println("File or directory exists")
//	} else {
//	    fmt.Println("File or directory does not exist")
//	}
//
// Parameters:
//   - path: The file or directory path to check.
//
// Returns:
//   - bool: True if the path exists, false otherwise.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// AppendToFile appends content to a file at the specified path.
//
// If the file does not exist, it is created with permissions 0644. The function opens the file in append mode
// and writes the provided content string. If any error occurs during file operations, it is returned.
//
// Example:
//
//	err := AppendToFile("log.txt", "Log entry\n")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Parameters:
//   - path: The file path to append content to.
//   - content: The string content to append to the file.
//
// Returns:
//   - error: An error if the file cannot be opened or written to.
func AppendToFile(path, content string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(content)
	return err
}

// CreateDirIfNotExist creates a directory at the specified path if it does not already exist.
//
// The function checks if the path is valid, not empty, and not too long (max 4096 characters).
// If the path exists and is a directory, no action is taken. If it exists as a file, an error is returned.
// Optional permissions can be provided; otherwise, the default permission is 0755.
//
// Example:
//
//	err := CreateDirIfNotExist("data/output", 0o755)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Parameters:
//   - path: The directory path to create.
//   - perm: Optional directory permission mode (os.FileMode). Defaults to 0755 if not provided.
//
// Returns:
//   - error: An error if the path is empty, too long, exists as a file, or directory creation fails.
func CreateDirIfNotExist(path string, perm ...os.FileMode) error {
	if path == "" || path == "." {
		return errors.New("path cannot be empty or root")
	}
	if len(path) > 4096 {
		return errors.New("path too long")
	}
	info, err := os.Stat(path)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("path %s is a file, not a directory", path)
		}
		return nil
	}
	if !os.IsNotExist(err) {
		return err
	}
	fileMode := os.FileMode(0o755)
	if len(perm) > 0 {
		fileMode = perm[0]
	}
	return os.MkdirAll(path, fileMode)
}

// CreateFileIfNotExist creates a file at the specified path if it does not already exist.
//
// The function checks if the path is valid, not empty, and not too long (max 4096 characters).
// If the path exists and is a file, no action is taken. If it exists as a directory, an error is returned.
// Optional permissions can be provided; otherwise, the default permission is 0600.
//
// Example:
//
//	err := CreateFileIfNotExist("data.txt", 0o644)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Parameters:
//   - path: The file path to create.
//   - perm: Optional file permission mode (os.FileMode). Defaults to 0600 if not provided.
//
// Returns:
//   - error: An error if the path is empty, too long, exists as a directory, or file creation fails.
func CreateFileIfNotExist(path string, perm ...os.FileMode) error {
	if path == "" || path == "." {
		return errors.New("path cannot be empty or root")
	}
	if len(path) > 4096 {
		return errors.New("path too long")
	}
	info, err := os.Stat(path)
	if err == nil {
		if info.IsDir() {
			return fmt.Errorf("path %s is a directory, not a file", path)
		}
		return nil
	}
	if !os.IsNotExist(err) {
		return err
	}
	fileMode := os.FileMode(0o600)
	if len(perm) > 0 {
		fileMode = perm[0]
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL, fileMode)
	if err != nil {
		return err
	}
	return file.Close()
}

// GetMimeTypeFromExtension returns the MIME type for a given file extension.
//
// If the extension does not start with a dot, it is added automatically. If no MIME type is found,
// the default "application/octet-stream" is returned. This function uses the standard library's mime package.
//
// Example:
//
//	mimeType := GetMimeTypeFromExtension(".pdf")
//	fmt.Println(mimeType) // Prints "application/pdf"
//
// Parameters:
//   - ext: The file extension (e.g., ".pdf" or "pdf").
//
// Returns:
//   - string: The MIME type for the extension, or "application/octet-stream" if unknown.
func GetMimeTypeFromExtension(ext string) string {
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		return "application/octet-stream"
	}
	return mimeType
}

// GetMimeTypeFromContent determines the MIME type of a file based on its content.
// It reads the first 512 bytes of the file and uses http.DetectContentType to identify the MIME type.
// If the file cannot be opened or read, an error is returned.
// For empty files, it returns "application/octet-stream" with no error.
//
// Example:
// mimeType, err := GetMimeTypeFromContent("image.png")
//
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// fmt.Println(mimeType) // Prints "image/png"
//
// Parameters:
// - path: The file path to analyze.
//
// Returns:
// - string: The detected MIME type (e.g., "image/png", "text/plain; charset=utf-8").
// - error: An error if the file cannot be opened or read.
func GetMimeTypeFromContent(path string) (string, error) {
	if path == "" || path == "." {
		return "", errors.New("path cannot be empty or root")
	}
	if len(path) > 4096 {
		return "", errors.New("path too long")
	}
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", err
	}
	if n == 0 {
		return "application/octet-stream", nil
	}
	mimeType := http.DetectContentType(buffer[:n])
	return mimeType, nil
}

// SanitizeFilename sanitizes a filename to ensure it is safe for use across Linux, macOS, and Windows.
//
// The function removes or replaces invalid characters, non-printable characters, and control characters,
// trims leading/trailing spaces and dots, and checks for reserved filenames (e.g., "CON", "."). It also
// ensures the filename length does not exceed 255 bytes, a common filesystem limit. If the filename is
// empty or invalid after sanitization, an error is returned.
//
// Example:
//
//	safeName, err := SanitizeFilename("my<file>.txt")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(safeName) // Prints "my_file_.txt"
//
// Parameters:
//   - filename: The filename to sanitize.
//
// Returns:
//   - string: The sanitized filename.
//   - error: An error if the filename is empty, a reserved name, or empty after sanitization.
func SanitizeFilename(filename string) (string, error) {
	if filename == "" {
		return "", errors.New("filename cannot be empty")
	}
	// Replace invalid filename characters with an underscore
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		filename = strings.ReplaceAll(filename, char, "_")
	}
	// Replace non-printable or control characters with an underscore
	cleaned := strings.Map(func(r rune) rune {
		if !unicode.IsPrint(r) || unicode.IsControl(r) {
			return '_'
		}
		return r
	}, filename)
	// Trim leading/trailing spaces and dots
	cleaned = strings.Trim(cleaned, " .")
	// Check for reserved filenames
	reservedNames := []string{
		".", "..", // Linux/macOS reserved
		"CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
	}
	baseWithoutExt := strings.TrimSuffix(cleaned, filepath.Ext(cleaned))
	if slices.ContainsFunc(reservedNames, func(s string) bool { return strings.EqualFold(baseWithoutExt, s) }) {
		return "", errors.New("filename is a reserved name: " + cleaned)
	}
	// Ensure the filename is not empty after cleaning
	if cleaned == "" {
		return "", errors.New("sanitized filename is empty")
	}
	// Limit filename length to 255 bytes (common filesystem limit)
	if len(cleaned) > 255 {
		ext := filepath.Ext(cleaned)
		maxBaseLen := 255 - len(ext)
		if maxBaseLen <= 0 {
			return "", errors.New("sanitized filename is empty after truncation")
		}
		cleaned = cleaned[:maxBaseLen] + ext
	}
	return cleaned, nil
}

// HasFileExtension checks if the provided string has a valid file extension.
//
// A valid extension is a non-empty suffix starting with a dot (e.g., ".txt").
// The function returns true if the string has an extension and false otherwise.
//
// Example:
//
//	if HasFileExtension("document.txt") {
//	    fmt.Println("Has valid extension")
//	} else {
//	    fmt.Println("No valid extension")
//	}
//
// Parameters:
//   - comp: The string to check for a file extension.
//
// Returns:
//   - bool: True if the string has a valid file extension, false otherwise.
func HasFileExtension(comp string) bool {
	ext := filepath.Ext(comp)
	return ext != "" && ext != comp
}
