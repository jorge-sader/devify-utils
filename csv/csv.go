package csv

import (
	"bytes"
	"encoding/csv"
	"errors"
	"os"

	"github.com/devify-me/devify-utils/fileio"
)

// ReadFile reads a CSV file from the specified path and stores the records in the provided destination.
//
// The destination must be a pointer to a slice of string slices (*[][]string). The function validates the file path,
// ensures it has a .csv extension, and checks that the file is not empty. If any errors occur during reading or if the
// destination type is incorrect, an error is returned.
//
// Example:
//
//	var records [][]string
//	err := ReadFile("data.csv", &records)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(records) // Prints the CSV records
//
// Parameters:
//   - path: The file path of the CSV file to read.
//   - dest: A pointer to a slice of string slices (*[][]string) where the CSV records will be stored.
//
// Returns:
//   - error: An error if the file cannot be read, the path is invalid, the file is empty, or the destination type is incorrect.
func ReadFile(path string, dest any) error {
	if err := fileio.ValidatePath(path, ".csv"); err != nil {
		return err
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}
	if len(records) == 0 {
		return errors.New("file is empty")
	}
	recordsPtr, ok := dest.(*[][]string)
	if !ok {
		return errors.New("destination must be *[][]string")
	}
	*recordsPtr = records
	return nil
}

// WriteFile writes a slice of string slices to a CSV file at the specified path.
//
// The data must be a slice of string slices ([][]string) and must not be empty. The function validates the file path,
// ensures it has a .csv extension, and creates any necessary parent directories. A file permission mode can be optionally
// provided; otherwise, a default mode of 0600 is used. If any errors occur during writing, an error is returned.
//
// Example:
//
//	records := [][]string{{"a", "b"}, {"c", "d"}}
//	err := WriteFile(records, "output.csv", 0o644)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Parameters:
//   - data: The CSV data to write, as a slice of string slices ([][]string).
//   - path: The file path where the CSV file will be written.
//   - perm: Optional file permission mode (os.FileMode). Defaults to 0600 if not provided.
//
// Returns:
//   - error: An error if the path is invalid, data is empty or of incorrect type, directory creation fails, or writing fails.
func WriteFile(data any, path string, perm ...os.FileMode) error {
	if err := fileio.ValidatePath(path, ".csv"); err != nil {
		return err
	}
	records, ok := data.([][]string)
	if !ok {
		return errors.New("data must be [][]string")
	}
	if len(records) == 0 {
		return errors.New("records cannot be empty")
	}
	if err := fileio.EnsureDir(path, 0o755); err != nil {
		return err
	}
	fileMode := os.FileMode(0o600)
	if len(perm) > 0 {
		fileMode = perm[0]
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fileMode)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	if err := writer.WriteAll(records); err != nil {
		return err
	}
	writer.Flush()
	return writer.Error()
}

// Marshal converts a slice of string slices to CSV-encoded bytes.
//
// The input data must be a slice of string slices ([][]string) and must not be empty. The function serializes the data
// into CSV format and returns the resulting bytes. If any errors occur during serialization, an error is returned.
//
// Example:
//
//	records := [][]string{{"a", "b"}, {"c", "d"}}
//	data, err := Marshal(records)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(string(data)) // Prints CSV-encoded string
//
// Parameters:
//   - data: The CSV data to marshal, as a slice of string slices ([][]string).
//
// Returns:
//   - []byte: The CSV-encoded data as bytes.
//   - error: An error if the data is empty, of incorrect type, or if serialization fails.
func Marshal(data any) ([]byte, error) {
	records, ok := data.([][]string)
	if !ok {
		return nil, errors.New("data must be [][]string")
	}
	if len(records) == 0 {
		return nil, errors.New("records cannot be empty")
	}
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	if err := writer.WriteAll(records); err != nil {
		return nil, err
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Unmarshal parses CSV-encoded bytes into a slice of string slices.
//
// The destination must be a pointer to a slice of string slices (*[][]string). The function parses the input bytes as CSV
// data and stores the records in the provided destination. If the input data is empty, the destination type is incorrect,
// or parsing fails, an error is returned.
//
// Example:
//
//	var records [][]string
//	data := []byte("a,b\nc,d")
//	err := Unmarshal(data, &records)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(records) // Prints [["a", "b"], ["c", "d"]]
//
// Parameters:
//   - data: The CSV-encoded data as bytes.
//   - dest: A pointer to a slice of string slices (*[][]string) where the parsed records will be stored.
//
// Returns:
//   - error: An error if the data is empty, the destination type is incorrect, no records are found, or parsing fails.
func Unmarshal(data []byte, dest any) error {
	if len(data) == 0 {
		return errors.New("CSV data cannot be empty")
	}
	recordsPtr, ok := dest.(*[][]string)
	if !ok {
		return errors.New("destination must be *[][]string")
	}
	reader := csv.NewReader(bytes.NewReader(data))
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}
	if len(records) == 0 {
		return errors.New("no records found")
	}
	*recordsPtr = records
	return nil
}
