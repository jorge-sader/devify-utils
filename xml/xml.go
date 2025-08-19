// Package xml provides utilities for XML serialization and file operations.
//
// This package offers functions for marshaling and unmarshaling XML data, as well as reading and writing XML files.
// It integrates with the fileio package from devify-utils for path validation and directory creation.
// All functions include error handling for common cases, such as empty data or invalid file paths,
// and the Marshal function prepends the standard XML header (<?xml version="1.0" encoding="UTF-8"?>).
package xml

import (
	"encoding/xml"
	"errors"
	"os"

	"github.com/devify-me/devify-utils/fileio"
)

// Marshal serializes the given data to XML format as a byte slice with an XML header.
//
// The function checks that the input data is not nil and marshals it to XML, prepending the standard XML header
// (<?xml version="1.0" encoding="UTF-8"?>). If serialization fails or the output is empty, an error is returned.
//
// Example:
//
//	type Person struct {
//	    Name string `xml:"name"`
//	    Age  int    `xml:"age"`
//	}
//	data := Person{Name: "Alice", Age: 30}
//	output, err := Marshal(data)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(string(output)) // Prints `<?xml version="1.0" encoding="UTF-8"?><Person><name>Alice</name><age>30</age></Person>`
//
// Parameters:
//   - data: The data to serialize to XML (must be compatible with encoding/xml, e.g., structs with XML tags).
//
// Returns:
//   - []byte: The XML-encoded data with the XML header as a byte slice.
//   - error: An error if the data is nil, cannot be marshaled, or results in empty XML.
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

// Unmarshal parses XML data into the provided destination.
//
// The destination must be a non-nil pointer to a struct or other type compatible with encoding/xml.
// The function checks that the input data is not empty and that the destination is not nil.
// If parsing fails, an error is returned.
//
// Example:
//
//	type Person struct {
//	    Name string `xml:"name"`
//	    Age  int    `xml:"age"`
//	}
//	var result Person
//	data := []byte(`<Person><name>Alice</name><age>30</age></Person>`)
//	err := Unmarshal(data, &result)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(result) // Prints {Alice 30}
//
// Parameters:
//   - data: The XML-encoded data as a byte slice.
//   - dest: A pointer to the destination where the parsed XML data will be stored.
//
// Returns:
//   - error: An error if the data is empty, the destination is nil, or parsing fails.
func Unmarshal(data []byte, dest any) error {
	if len(data) == 0 {
		return errors.New("XML data cannot be empty")
	}
	if dest == nil {
		return errors.New("destination cannot be nil")
	}
	return xml.Unmarshal(data, dest)
}

// ReadFile reads an XML file from the specified path and unmarshals it into the provided destination.
//
// The function validates that the file path has a ".xml" extension and exists using fileio.ValidatePath.
// It checks that the file is not empty before attempting to unmarshal the data into the destination,
// which must be a non-nil pointer to a struct or other type compatible with encoding/xml.
//
// Example:
//
//	type Person struct {
//	    Name string `xml:"name"`
//	    Age  int    `xml:"age"`
//	}
//	var result Person
//	err := ReadFile("person.xml", &result)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(result) // Prints the parsed XML data, e.g., {Alice 30}
//
// Parameters:
//   - path: The file path of the XML file to read.
//   - dest: A pointer to the destination where the parsed XML data will be stored.
//
// Returns:
//   - error: An error if the path is invalid, the file is empty, or unmarshaling fails.
func ReadFile(path string, dest any) error {
	if err := fileio.ValidateReadPath(path, ".xml"); err != nil {
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

// WriteFile serializes the given data to XML and writes it to a file at the specified path.
//
// The function validates that the file path has a ".xml" extension using fileio.ValidatePath and ensures
// the parent directories exist using fileio.EnsureDir. The data is marshaled to XML with the standard XML header,
// and the resulting bytes are written to the file with the specified permissions (defaulting to 0600 if not provided).
// If the data cannot be marshaled or the file cannot be written, an error is returned.
//
// Example:
//
//	type Person struct {
//	    Name string `xml:"name"`
//	    Age  int    `xml:"age"`
//	}
//	data := Person{Name: "Alice", Age: 30}
//	err := WriteFile(data, "person.xml", 0o644)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Parameters:
//   - data: The data to serialize to XML (must be compatible with encoding/xml, e.g., structs with XML tags).
//   - path: The file path where the XML data will be written.
//   - perm: Optional file permission mode (os.FileMode). Defaults to 0600 if not provided.
//
// Returns:
//   - error: An error if the path is invalid, data cannot be marshaled, directories cannot be created,
//     or the file cannot be written.
func WriteFile(data any, path string, perm ...os.FileMode) error {
	if err := fileio.ValidateWritePath(path, ".xml"); err != nil {
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
