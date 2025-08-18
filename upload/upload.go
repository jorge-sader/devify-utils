// Package upload provides utilities for handling file uploads from HTTP requests.
//
// This package offers functions to process single or multiple file uploads, with support for sanitizing filenames,
// validating file types and sizes, and optionally renaming files with random hex names. It integrates with the
// filesystem package from devify-utils for safe filename handling and directory creation, and uses the go-playground/validator
// package for struct validation. All functions include robust error handling for invalid inputs or file operations.
package upload

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"

	"github.com/devify-me/devify-utils/filesystem"
	"github.com/go-playground/validator/v10"
)

// FileOperation manages configuration for file upload operations.
//
// It specifies the maximum file size, allowed file types, and a validator instance for validating uploaded files.
// The Validate field must be initialized with a validator.Validate instance that includes the "allowedfiletype"
// validation rule registered via IsAllowedFileType.
type FileOperation struct {
	// MaxFileSize is the maximum allowed file size in megabytes.
	MaxFileSize int64
	// AllowedFileTypes is a slice of allowed MIME types (e.g., "image/png", "application/pdf").
	AllowedFileTypes []string
	// Validate is the validator instance for validating UploadedFile structs.
	Validate *validator.Validate
}

// UploadedFile represents metadata for an uploaded file.
//
// It contains the original and encoded filenames, the full path, MIME type, extension, and file size,
// with validation tags for use with the go-playground/validator package.
type UploadedFile struct {
	// OriginalName is the original filename provided by the client.
	OriginalName string `validate:"required"`
	// EncodedName is the sanitized or randomly generated filename used for storage.
	EncodedName string `validate:"required"`
	// FullPath is the full filesystem path where the file is stored.
	FullPath string `validate:"required"`
	// FileMimeType is the MIME type of the file (e.g., "image/png").
	FileMimeType string `validate:"required,allowedfiletype"`
	// Extension is the file extension (e.g., ".png").
	Extension string `validate:"required"`
	// FileSize is the size of the file in bytes.
	FileSize int64 `validate:"gte=0"`
}

// generateRandomHex generates a random hexadecimal string of n characters.
//
// The number of characters (n) must be even, as each byte is encoded as two hexadecimal characters.
// The function uses crypto/rand for secure random number generation. It is unexported as it is intended for internal use.
//
// Parameters:
//   - n: The length of the hexadecimal string (must be even).
//
// Returns:
//   - string: A random hexadecimal string of length n.
//   - error: An error if n is odd or random generation fails.
func generateRandomHex(n int) (string, error) {
	if n%2 != 0 {
		return "", fmt.Errorf("n must be even for hex encoding")
	}
	bytes := make([]byte, n/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// UploadFiles handles uploading multiple files from an HTTP request to the specified directory.
//
// The function parses the multipart form data, validates file sizes and types, sanitizes filenames using
// filesystem.SanitizeFilename, and optionally renames files with a random 32-character hex string.
// Each uploaded file is validated using the FileOperation.Validate instance, which must have the
// "allowedfiletype" validation rule registered. The files are saved to the uploadDir, which is created
// if it does not exist. An error is returned if no files are uploaded or if any operation fails.
//
// Example:
//
//	fo := &FileOperation{
//	    MaxFileSize:     10, // 10 MB
//	    AllowedFileTypes: []string{"image/png", "image/jpeg"},
//	    Validate:         validator.New(),
//	}
//	fo.Validate.RegisterValidation("allowedfiletype", fo.IsAllowedFileType)
//	files, err := fo.UploadFiles(r, "uploads", true)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, f := range files {
//	    fmt.Println(f.FullPath) // Prints paths of uploaded files
//	}
//
// Parameters:
//   - r: The HTTP request containing the multipart form data with files.
//   - uploadDir: The directory where files will be saved (created if it does not exist).
//   - rename: If true, files are renamed with a random 32-character hex string plus their original extension.
//
// Returns:
//   - []UploadedFile: A slice of metadata for successfully uploaded files.
//   - error: An error if the upload directory cannot be created, form parsing fails, or any file operation or validation fails.
func (f *FileOperation) UploadFiles(r *http.Request, uploadDir string, rename bool) ([]UploadedFile, error) {
	if err := filesystem.CreateDirIfNotExist(uploadDir); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}
	if err := r.ParseMultipartForm(f.MaxFileSize << 20); err != nil {
		return nil, fmt.Errorf("failed to parse multipart form: %w", err)
	}
	var uploadedFiles []UploadedFile
	for _, fileHeaders := range r.MultipartForm.File {
		for _, header := range fileHeaders {
			uploadedFile, err := func() (*UploadedFile, error) {
				file, err := header.Open()
				if err != nil {
					return nil, fmt.Errorf("failed to open file: %w", err)
				}
				defer file.Close()
				if header.Filename == "" {
					return nil, errors.New("filename cannot be empty")
				}
				if header.Size > f.MaxFileSize {
					return nil, fmt.Errorf("file size %d exceeds maximum %d", header.Size, f.MaxFileSize)
				}
				sanitizedName, err := filesystem.SanitizeFilename(header.Filename)
				if err != nil {
					return nil, fmt.Errorf("failed to sanitize filename: %w", err)
				}
				var encodedName string
				if rename {
					hexStr, err := generateRandomHex(32)
					if err != nil {
						return nil, fmt.Errorf("failed to generate random name: %w", err)
					}
					encodedName = hexStr + filepath.Ext(sanitizedName)
				} else {
					encodedName = sanitizedName
				}
				fullPath := filepath.Join(uploadDir, encodedName)
				destFile, err := os.Create(fullPath)
				if err != nil {
					return nil, fmt.Errorf("failed to create destination file: %w", err)
				}
				defer destFile.Close()
				_, err = io.Copy(destFile, file)
				if err != nil {
					return nil, fmt.Errorf("failed to write file: %w", err)
				}
				uploadedFile := UploadedFile{
					OriginalName: header.Filename,
					EncodedName:  encodedName,
					FullPath:     fullPath,
					FileMimeType: header.Header.Get("Content-Type"),
					Extension:    filepath.Ext(encodedName),
					FileSize:     header.Size,
				}
				if err := f.Validate.Struct(uploadedFile); err != nil {
					return nil, fmt.Errorf("failed to validate uploaded file: %w", err)
				}
				return &uploadedFile, nil
			}()
			if err != nil {
				return uploadedFiles, fmt.Errorf("failed to save uploaded file: %w", err)
			}
			uploadedFiles = append(uploadedFiles, *uploadedFile)
		}
	}
	if len(uploadedFiles) == 0 {
		return nil, errors.New("no files uploaded")
	}
	return uploadedFiles, nil
}

// UploadOneFile handles uploading a single file from an HTTP request to the specified directory.
//
// The function wraps UploadFiles to process a single file, ensuring exactly one file is uploaded.
// It validates the file size, type, and filename, and optionally renames the file with a random 32-character
// hex string. The file is saved to the uploadDir, which is created if it does not exist. An error is returned
// if multiple files are uploaded or if any operation fails.
//
// Example:
//
//	fo := &FileOperation{
//	    MaxFileSize:     5, // 5 MB
//	    AllowedFileTypes: []string{"text/plain"},
//	    Validate:         validator.New(),
//	}
//	fo.Validate.RegisterValidation("allowedfiletype", fo.IsAllowedFileType)
//	file, err := fo.UploadOneFile(r, "uploads", false)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(file.FullPath) // Prints the path of the uploaded file
//
// Parameters:
//   - r: The HTTP request containing the multipart form data with a single file.
//   - uploadDir: The directory where the file will be saved (created if it does not exist).
//   - rename: If true, the file is renamed with a random 32-character hex string plus its original extension.
//
// Returns:
//   - *UploadedFile: A pointer to the metadata for the uploaded file.
//   - error: An error if exactly one file is not uploaded, or if any file operation or validation fails.
func (f *FileOperation) UploadOneFile(r *http.Request, uploadDir string, rename bool) (*UploadedFile, error) {
	files, err := f.UploadFiles(r, uploadDir, rename)
	if err != nil {
		return nil, err
	}
	if len(files) != 1 {
		return nil, fmt.Errorf("multiple files uploaded, expected one")
	}
	return &files[0], nil
}

// IsAllowedFileType validates if a file's MIME type is in the allowed list for an UploadedFile struct.
//
// This function is used as a custom validation rule for the go-playground/validator package, checking
// if the FileMimeType field of an UploadedFile is in the FileOperation.AllowedFileTypes slice.
// It must be registered with the validator instance in FileOperation.Validate using the "allowedfiletype" tag.
//
// Example:
//
//	fo := &FileOperation{
//	    AllowedFileTypes: []string{"image/png", "image/jpeg"},
//	    Validate:         validator.New(),
//	}
//	fo.Validate.RegisterValidation("allowedfiletype", fo.IsAllowedFileType)
//	// Use fo.Validate to validate an UploadedFile struct
//
// Parameters:
//   - fl: The validator.FieldLevel instance providing access to the field being validated.
//
// Returns:
//   - bool: True if the MIME type is in AllowedFileTypes, false otherwise.
func (f *FileOperation) IsAllowedFileType(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return slices.Contains(f.AllowedFileTypes, value)
}
