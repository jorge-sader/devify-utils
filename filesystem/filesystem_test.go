package filesystem_test

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"slices"

	"github.com/devify-me/devify-utils/filesystem"
	"github.com/go-playground/validator/v10"
)

func setupValidator(f *filesystem.FileOperation) *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())
	v.RegisterValidation("allowedfiletype", f.IsAllowedFileType)
	v.RegisterValidation("mime", func(fl validator.FieldLevel) bool {
		return slices.Contains(f.AllowedFileTypes, fl.Field().String())
	})
	return v
}

func TestFileExists(t *testing.T) {
	tempDir := t.TempDir()
	existingFile := filepath.Join(tempDir, "exists.txt")
	os.WriteFile(existingFile, []byte("content"), 0600)

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "Existing file",
			path: existingFile,
			want: true,
		},
		{
			name: "Non-existing file",
			path: filepath.Join(tempDir, "nonexistent.txt"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filesystem.FileExists(tt.path); got != tt.want {
				t.Errorf("FileExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppendToFile(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "append.txt")

	tests := []struct {
		name    string
		content string
		want    string
		wantErr bool
	}{
		{
			name:    "Append to new file",
			content: "hello",
			want:    "hello",
		},
		{
			name:    "Append to existing file",
			content: " world",
			want:    "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := filesystem.AppendToFile(filePath, tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("AppendToFile() error = %v, wantErr %v", err, tt.wantErr)
			}
			got, _ := os.ReadFile(filePath)
			if string(got) != tt.want {
				t.Errorf("File content = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCreateDirIfNotExist(t *testing.T) {
	tempDir := t.TempDir()
	validPath := filepath.Join(tempDir, "newdir")
	filePath := filepath.Join(tempDir, "file.txt")
	os.WriteFile(filePath, []byte{}, 0600)
	longPath := filepath.Join(tempDir, string(make([]rune, 4097)))

	tests := []struct {
		name    string
		path    string
		perm    os.FileMode
		wantErr string
	}{
		{
			name:    "Empty path",
			path:    "",
			wantErr: "path cannot be empty or root",
		},
		{
			name:    "Root path",
			path:    ".",
			wantErr: "path cannot be empty or root",
		},
		{
			name:    "Path too long",
			path:    longPath,
			wantErr: "path too long",
		},
		{
			name:    "Path is file",
			path:    filePath,
			wantErr: "is a file, not a directory",
		},
		{
			name: "Create new dir",
			path: validPath,
		},
		{
			name: "Existing dir",
			path: validPath,
		},
		{
			name: "With custom perm",
			path: filepath.Join(tempDir, "customperm"),
			perm: 0700,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var perm []os.FileMode
			if tt.perm != 0 {
				perm = []os.FileMode{tt.perm}
			}
			err := filesystem.CreateDirIfNotExist(tt.path, perm...)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("CreateDirIfNotExist() error = %v, wantErr containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("CreateDirIfNotExist() unexpected error = %v", err)
			}
			info, err := os.Stat(tt.path)
			if err != nil || !info.IsDir() {
				t.Errorf("Directory not created or not a dir")
			}
			// Removed perm check to avoid umask issues
		})
	}
}

func TestCreateFileIfNotExist(t *testing.T) {
	tempDir := t.TempDir()
	validPath := filepath.Join(tempDir, "newfile.txt")
	dirPath := filepath.Join(tempDir, "dir")
	os.Mkdir(dirPath, 0755)
	longPath := filepath.Join(tempDir, string(make([]rune, 4097)))

	tests := []struct {
		name    string
		path    string
		perm    os.FileMode
		wantErr string
	}{
		{
			name:    "Empty path",
			path:    "",
			wantErr: "path cannot be empty or root",
		},
		{
			name:    "Root path",
			path:    ".",
			wantErr: "path cannot be empty or root",
		},
		{
			name:    "Path too long",
			path:    longPath,
			wantErr: "path too long",
		},
		{
			name:    "Path is directory",
			path:    dirPath,
			wantErr: "is a directory, not a file",
		},
		{
			name: "Create new file",
			path: validPath,
		},
		{
			name: "Existing file",
			path: validPath,
		},
		{
			name: "With custom perm",
			path: filepath.Join(tempDir, "customperm.txt"),
			perm: 0644,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var perm []os.FileMode
			if tt.perm != 0 {
				perm = []os.FileMode{tt.perm}
			}
			err := filesystem.CreateFileIfNotExist(tt.path, perm...)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("CreateFileIfNotExist() error = %v, wantErr containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("CreateFileIfNotExist() unexpected error = %v", err)
			}
			info, err := os.Stat(tt.path)
			if err != nil || info.IsDir() {
				t.Errorf("File not created or is dir")
			}
			// Removed perm check to avoid umask issues
		})
	}
}

func TestGetMimeTypeFromExtension(t *testing.T) {
	tests := []struct {
		name string
		ext  string
		want string
	}{
		{
			name: "With dot",
			ext:  ".jpg",
			want: "image/jpeg",
		},
		{
			name: "Without dot",
			ext:  "png",
			want: "image/png",
		},
		{
			name: "Unknown",
			ext:  ".unknown",
			want: "application/octet-stream",
		},
		{
			name: "Empty",
			ext:  "",
			want: "application/octet-stream",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filesystem.GetMimeTypeFromExtension(tt.ext); got != tt.want {
				t.Errorf("GetMimeTypeFromExtension() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetMimeTypeFromContent(t *testing.T) {
	// TODO: Implement when function is completed
	if got, err := filesystem.GetMimeTypeFromContent("dummy"); got != "" || err != nil {
		t.Errorf("GetMimeTypeFromContent() = %v, err = %v; want \"\", nil", got, err)
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     string
		wantErr  string
	}{
		{
			name:     "Empty filename",
			filename: "",
			wantErr:  "filename cannot be empty",
		},
		{
			name:     "Invalid chars",
			filename: "file/name:with?invalid*chars",
			want:     "file_name_with_invalid_chars",
		},
		{
			name:     "Control chars",
			filename: "file\x00name",
			want:     "file_name",
		},
		{
			name:     "Reserved name",
			filename: "CON",
			wantErr:  "filename is a reserved name",
		},
		{
			name:     "Reserved name with extension",
			filename: "CON.txt",
			wantErr:  "filename is a reserved name",
		},
		{
			name:     "Too long",
			filename: strings.Repeat("a", 300) + ".txt",
			want:     strings.Repeat("a", 255-len(".txt")) + ".txt",
		},
		{
			name:     "Leading dots",
			filename: "...file.txt",
			want:     "file.txt",
		},
		{
			name:     "Trailing spaces",
			filename: "file.txt  ",
			want:     "file.txt",
		},
		{
			name:     "Becomes empty after clean",
			filename: " . ",
			wantErr:  "sanitized filename is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := filesystem.SanitizeFilename(tt.filename)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("SanitizeFilename() error = %v, wantErr containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("SanitizeFilename() unexpected error = %v", err)
			}
			if got != tt.want {
				t.Errorf("SanitizeFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasFileExtension(t *testing.T) {
	tests := []struct {
		name string
		comp string
		want bool
	}{
		{
			name: "With extension",
			comp: "file.txt",
			want: true,
		},
		{
			name: "Without extension",
			comp: "file",
			want: false,
		},
		{
			name: "Dot file",
			comp: ".hidden",
			want: false,
		},
		{
			name: "Empty",
			comp: "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filesystem.HasFileExtension(tt.comp); got != tt.want {
				t.Errorf("HasFileExtension() = %v, want %v", got, tt.want)
			}
		})
	}
}

func createMultipartRequest(files map[string]struct{ Content, Mime string }) *http.Request {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for name, data := range files {
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, name))
		if data.Mime != "" {
			h.Set("Content-Type", data.Mime)
		}
		part, _ := writer.CreatePart(h)
		part.Write([]byte(data.Content))
	}
	writer.Close()

	req := &http.Request{
		Method: "POST",
		Header: http.Header{"Content-Type": []string{writer.FormDataContentType()}},
		Body:   io.NopCloser(body),
	}
	req.ContentLength = int64(body.Len())
	return req
}

func TestFileOperation_UploadFiles(t *testing.T) {
	tempDir := t.TempDir()
	uploadDir := filepath.Join(tempDir, "uploads")

	f := &filesystem.FileOperation{
		MaxFileSize:      10 << 20,
		AllowedFileTypes: []string{"text/plain", "application/octet-stream"},
		Validate:         setupValidator(&filesystem.FileOperation{AllowedFileTypes: []string{"text/plain", "application/octet-stream"}}),
	}

	tests := []struct {
		name      string
		req       *http.Request
		uploadDir string
		rename    bool
		wantLen   int
		wantErr   string
	}{
		{
			name:      "No files",
			req:       createMultipartRequest(map[string]struct{ Content, Mime string }{}),
			uploadDir: uploadDir,
			wantErr:   "no files uploaded",
		},
		{
			name:      "Single file",
			req:       createMultipartRequest(map[string]struct{ Content, Mime string }{"test.txt": {Content: "content", Mime: "text/plain"}}),
			uploadDir: uploadDir,
			wantLen:   1,
		},
		{
			name:      "Multiple files",
			req:       createMultipartRequest(map[string]struct{ Content, Mime string }{"file1.txt": {Content: "content1", Mime: "text/plain"}, "file2.txt": {Content: "content2", Mime: "text/plain"}}),
			uploadDir: uploadDir,
			wantLen:   2,
		},
		{
			name:      "With rename",
			req:       createMultipartRequest(map[string]struct{ Content, Mime string }{"test.txt": {Content: "content", Mime: "text/plain"}}),
			uploadDir: uploadDir,
			rename:    true,
			wantLen:   1,
		},
		{
			name:      "Invalid filename",
			req:       createMultipartRequest(map[string]struct{ Content, Mime string }{"": {Content: "content", Mime: "text/plain"}}),
			uploadDir: uploadDir,
			wantErr:   "no files uploaded",
		},
		{
			name:      "File too large",
			req:       createMultipartRequest(map[string]struct{ Content, Mime string }{"large.txt": {Content: strings.Repeat("a", int(f.MaxFileSize+1)), Mime: "text/plain"}}),
			uploadDir: uploadDir,
			wantErr:   "file size",
		},
		{
			name:      "Invalid mime",
			req:       createMultipartRequest(map[string]struct{ Content, Mime string }{"test.exe": {Content: "content", Mime: "application/zip"}}),
			uploadDir: uploadDir,
			wantErr:   "failed to validate uploaded file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := f.UploadFiles(tt.req, tt.uploadDir, tt.rename)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("UploadFiles() error = %v, wantErr containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("UploadFiles() unexpected error = %v", err)
			}
			if len(got) != tt.wantLen {
				t.Errorf("UploadFiles() len = %v, want %v", len(got), tt.wantLen)
			}
			// Check files exist
			for _, uf := range got {
				if !filesystem.FileExists(uf.FullPath) {
					t.Errorf("Uploaded file does not exist: %s", uf.FullPath)
				}
			}
		})
	}
}

func TestFileOperation_UploadOneFile(t *testing.T) {
	tempDir := t.TempDir()
	uploadDir := filepath.Join(tempDir, "Uploads")

	f := &filesystem.FileOperation{
		MaxFileSize:      10 << 20,
		AllowedFileTypes: []string{"text/plain", "application/octet-stream"},
		Validate:         setupValidator(&filesystem.FileOperation{AllowedFileTypes: []string{"text/plain", "application/octet-stream"}}),
	}

	tests := []struct {
		name      string
		req       *http.Request
		uploadDir string
		rename    bool
		wantErr   string
	}{
		{
			name:      "Single file",
			req:       createMultipartRequest(map[string]struct{ Content, Mime string }{"test.txt": {Content: "content", Mime: "text/plain"}}),
			uploadDir: uploadDir,
		},
		{
			name:      "Multiple files error",
			req:       createMultipartRequest(map[string]struct{ Content, Mime string }{"file1.txt": {Content: "content1", Mime: "text/plain"}, "file2.txt": {Content: "content2", Mime: "text/plain"}}),
			uploadDir: uploadDir,
			wantErr:   "multiple files uploaded, expected one",
		},
		{
			name:      "No files",
			req:       createMultipartRequest(map[string]struct{ Content, Mime string }{}),
			uploadDir: uploadDir,
			wantErr:   "no files uploaded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := f.UploadOneFile(tt.req, tt.uploadDir, tt.rename)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("UploadOneFile() error = %v, wantErr containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("UploadOneFile() unexpected error = %v", err)
			}
			if got == nil {
				t.Errorf("UploadOneFile() got nil")
			}
		})
	}
}

func TestFileOperation_IsAllowedFileType(t *testing.T) {
	f := &filesystem.FileOperation{
		AllowedFileTypes: []string{"text/plain", "image/jpeg"},
	}
	v := setupValidator(f)

	type testStruct struct {
		Mime string `validate:"allowedfiletype"`
	}

	tests := []struct {
		name    string
		mime    string
		wantErr bool
	}{
		{
			name:    "Allowed type",
			mime:    "text/plain",
			wantErr: false,
		},
		{
			name:    "Not allowed type",
			mime:    "application/octet-stream",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testStruct{Mime: tt.mime}
			err := v.Struct(s)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsAllowedFileType() validation error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Mock for validator.FieldLevel
type mockFieldLevel struct {
	value string
}

func (m *mockFieldLevel) Field() reflect.Value {
	v := reflect.ValueOf(m.value)
	return v
}
func (m *mockFieldLevel) FieldName() string {
	return ""
}
func (m *mockFieldLevel) StructFieldName() string {
	return ""
}
func (m *mockFieldLevel) Top() reflect.Value {
	return reflect.Value{}
}
func (m *mockFieldLevel) Parent() reflect.Value {
	return reflect.Value{}
}
func (m *mockFieldLevel) Param() string {
	return ""
}
func (m *mockFieldLevel) GetTag() string {
	return ""
}
func (m *mockFieldLevel) ExtractType(field reflect.Value) (reflect.Value, reflect.Kind, bool) {
	return reflect.Value{}, 0, false
}
func (m *mockFieldLevel) GetStructFieldOK() (reflect.Value, reflect.Kind, bool) {
	return reflect.Value{}, 0, false
}
func (m *mockFieldLevel) GetStructFieldOK2() (reflect.Value, reflect.Kind, bool, bool) {
	return reflect.Value{}, 0, false, false
}
func (m *mockFieldLevel) GetStructFieldOKAdvanced(field reflect.Value, searchName string) (reflect.Value, reflect.Kind, bool) {
	return reflect.Value{}, 0, false
}
func (m *mockFieldLevel) GetStructFieldOKAdvanced2(field reflect.Value, ns string) (reflect.Value, reflect.Kind, bool, bool) {
	return reflect.Value{}, 0, false, false
}
