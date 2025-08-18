package xml

import (
	"encoding/xml"
	"errors"
	"os"
	"path/filepath"
)

// Marshal serializes the given data to XML format.
// It returns an error if the data cannot be marshaled.
func Marshal(data any) ([]byte, error) {
	if data == nil {
		return nil, errors.New("data cannot be nil")
	}
	output, err := xml.Marshal(data)
	if err != nil {
		return nil, err
	}
	if len(output) == 0 {
		return nil, errors.New("marshaled XML is empty")
	}
	// Add XML header
	header := []byte(xml.Header)
	return append(header, output...), nil
}

// Unmarshal parses XML data into the given destination struct.
// It returns an error if the input is empty or cannot be unmarshaled.
func Unmarshal(data []byte, dest any) error {
	if len(data) == 0 {
		return errors.New("XML data cannot be empty")
	}
	if dest == nil {
		return errors.New("destination cannot be nil")
	}
	return xml.Unmarshal(data, dest)
}

// ReadFile reads an XML file from the given path and unmarshals it into the destination struct.
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

	// Ensure the file has a .xml extension
	ext := filepath.Ext(path)
	if ext != ".xml" {
		return errors.New("file must have .xml extension")
	}

	// Read the file content
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return errors.New("file is empty")
	}

	// Unmarshal the XML content
	return Unmarshal(data, dest)
}

// WriteFile marshals the given data to XML and writes it to the specified file path.
// It validates the directory path and ensures the file has a valid XML extension.
func WriteFile(data any, path string, perm ...os.FileMode) error {
	if path == "" || path == "." {
		return errors.New("path cannot be empty or root")
	}
	if len(path) > 4096 {
		return errors.New("path too long")
	}

	// Ensure the file has a .xml extension
	ext := filepath.Ext(path)
	if ext != ".xml" {
		return errors.New("file must have .xml extension")
	}

	// Marshal the data to XML
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

	// Write the XML content to the file
	return os.WriteFile(path, output, fileMode)
}
