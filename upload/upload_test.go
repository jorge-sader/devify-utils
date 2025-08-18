package upload

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/devify-me/devify-utils/filesystem"
	"github.com/go-playground/validator/v10"
)

func setupValidator(f *FileOperation) *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())
	v.RegisterValidation("allowedfiletype", f.IsAllowedFileType)
	v.RegisterValidation("mime", func(fl validator.FieldLevel) bool {
		return slices.Contains(f.AllowedFileTypes, fl.Field().String())
	})
	return v
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

	f := &FileOperation{
		MaxFileSize:      10 << 20,
		AllowedFileTypes: []string{"text/plain", "application/octet-stream"},
		Validate:         setupValidator(&FileOperation{AllowedFileTypes: []string{"text/plain", "application/octet-stream"}}),
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

	f := &FileOperation{
		MaxFileSize:      10 << 20,
		AllowedFileTypes: []string{"text/plain", "application/octet-stream"},
		Validate:         setupValidator(&FileOperation{AllowedFileTypes: []string{"text/plain", "application/octet-stream"}}),
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
	f := &FileOperation{
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
