package filesystem

import (
	"crypto/rand"
	"encoding/hex"
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

	"github.com/go-playground/validator/v10"
)

// CaseStyle defines the style of the filename case.
type CaseStyle string

// FileOperation holds configuration for file operations.
type FileOperation struct {
	MaxFileSize      int64
	AllowedFileTypes []string
	Validate         *validator.Validate
}

// UploadedFile represents metadata for an uploaded file.
type UploadedFile struct {
	OriginalName string `validate:"required"`
	EncodedName  string `validate:"required"`
	FullPath     string `validate:"required"`
	FileMimeType string `validate:"required,allowedfiletype"`
	Extension    string `validate:"required"`
	FileSize     int64  `validate:"gte=0"`
}

// FileExists checks if a file exists using os.Stat.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// AppendToFile appends content to a file.
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

// GetMimeTypeFromContent is a placeholder for content-based MIME detection.
func GetMimeTypeFromContent(path string) (string, error) {
	// TODO: Read first 512 characters and determine mime type
	return "", nil
}

// SanitizeFilename sanitizes a filename by cleaning invalid characters, normalizing whitespace,
// and ensuring the filename is safe for use across Linux, macOS, and Windows.
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

// HasFileExtension checks if the component has a valid file extension.
func HasFileExtension(comp string) bool {
	ext := filepath.Ext(comp)
	return ext != "" && ext != comp
}

// generateRandomHex generates a random hexadecimal string of n characters (n must be even).
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

// UploadFiles handles uploading multiple files from an HTTP request.
func (f *FileOperation) UploadFiles(r *http.Request, uploadDir string, rename bool) ([]UploadedFile, error) {
	if err := CreateDirIfNotExist(uploadDir); err != nil {
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
				sanitizedName, err := SanitizeFilename(header.Filename)
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

// UploadOneFile handles uploading a single file from an HTTP request.
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

// IsAllowedFileType is a custom validation function for checking against a []string
func (f *FileOperation) IsAllowedFileType(fl validator.FieldLevel) bool {
	// Get the field value
	value := fl.Field().String()
	// Check if the value exists in the AllowedFileTypes slice
	return slices.Contains(f.AllowedFileTypes, value)
}
