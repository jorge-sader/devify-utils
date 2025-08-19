package fileio_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/devify-me/devify-utils/fileio"
)

func TestValidateReadPath(t *testing.T) {
	tempDir := t.TempDir()
	validPath := filepath.Join(tempDir, "test.csv")
	os.WriteFile(validPath, []byte("test"), 0600)
	dirPath := filepath.Join(tempDir, "dir")
	os.Mkdir(dirPath, 0755)
	invalidExtPath := filepath.Join(tempDir, "test.txt")
	os.WriteFile(invalidExtPath, []byte("test"), 0600)
	emptyPath := filepath.Join(tempDir, "empty.csv")
	os.WriteFile(emptyPath, []byte{}, 0600)
	nonexistentPath := filepath.Join(tempDir, "nonexistent.csv")
	longPath := filepath.Join(tempDir, string(make([]rune, 4097)))

	tests := []struct {
		name    string
		path    string
		ext     string
		wantErr string
	}{
		{
			name:    "Valid path",
			path:    validPath,
			ext:     ".csv",
			wantErr: "",
		},
		{
			name:    "Empty path",
			path:    "",
			ext:     ".csv",
			wantErr: "path cannot be empty or root",
		},
		{
			name:    "Root path",
			path:    ".",
			ext:     ".csv",
			wantErr: "path cannot be empty or root",
		},
		{
			name:    "Path too long",
			path:    longPath,
			ext:     ".csv",
			wantErr: "path too long",
		},
		{
			name:    "File not exist",
			path:    nonexistentPath,
			ext:     ".csv",
			wantErr: "file does not exist",
		},
		{
			name:    "Path is directory",
			path:    dirPath,
			ext:     ".csv",
			wantErr: "path is a directory, not a file",
		},
		{
			name:    "Invalid extension",
			path:    invalidExtPath,
			ext:     ".csv",
			wantErr: "file must have .csv extension",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fileio.ValidateReadPath(tt.path, tt.ext)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("ValidateReadPath() error = %v, wantErr containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("ValidateReadPath() unexpected error = %v", err)
			}
		})
	}
}

func TestValidateWritePath(t *testing.T) {
	tempDir := t.TempDir()
	validPath := filepath.Join(tempDir, "test.csv")
	invalidExtPath := filepath.Join(tempDir, "test.txt")
	longPath := filepath.Join(tempDir, string(make([]rune, 4097)))

	tests := []struct {
		name    string
		path    string
		ext     string
		wantErr string
	}{
		{
			name:    "Valid path",
			path:    validPath,
			ext:     ".csv",
			wantErr: "",
		},
		{
			name:    "Empty path",
			path:    "",
			ext:     ".csv",
			wantErr: "path cannot be empty or root",
		},
		{
			name:    "Root path",
			path:    ".",
			ext:     ".csv",
			wantErr: "path cannot be empty or root",
		},
		{
			name:    "Path too long",
			path:    longPath,
			ext:     ".csv",
			wantErr: "path too long",
		},
		{
			name:    "Invalid extension",
			path:    invalidExtPath,
			ext:     ".csv",
			wantErr: "file must have .csv extension",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fileio.ValidateWritePath(tt.path, tt.ext)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("ValidateWritePath() error = %v, wantErr containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("ValidateWritePath() unexpected error = %v", err)
			}
		})
	}
}

func TestEnsureDir(t *testing.T) {
	tempDir := t.TempDir()
	validPath := filepath.Join(tempDir, "subdir/test.csv")
	existingDir := filepath.Join(tempDir, "existing")
	os.Mkdir(existingDir, 0755)

	tests := []struct {
		name    string
		path    string
		perm    os.FileMode
		wantErr bool
	}{
		{
			name: "Create new directory",
			path: validPath,
			perm: 0755,
		},
		{
			name: "Existing directory",
			path: filepath.Join(existingDir, "test.csv"),
			perm: 0755,
		},
		{
			name: "Current directory",
			path: "test.csv",
			perm: 0755,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fileio.EnsureDir(tt.path, tt.perm)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnsureDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				dir := filepath.Dir(tt.path)
				if dir != "." {
					info, err := os.Stat(dir)
					if err != nil || !info.IsDir() {
						t.Errorf("Directory %s was not created", dir)
					}
				}
			}
		})
	}
}
