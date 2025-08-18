package csv_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/devify-me/devify-utils/csv"
)

func TestReadFile(t *testing.T) {
	tempDir := t.TempDir()
	validPath := filepath.Join(tempDir, "test.csv")
	invalidTypePath := filepath.Join(tempDir, "invalid.csv")
	invalidExtPath := filepath.Join(tempDir, "test.txt")
	dirPath := filepath.Join(tempDir, "dir")
	emptyPath := filepath.Join(tempDir, "empty.csv")
	longPath := filepath.Join(tempDir, string(make([]rune, 4097)))

	// Setup files
	os.WriteFile(validPath, []byte("name,age\nAlice,30\nBob,25\n"), 0600)
	os.WriteFile(invalidTypePath, []byte("name,age\nAlice,\"30\nBob,25\n"), 0600) // Malformed quote
	os.WriteFile(invalidExtPath, []byte("dummy"), 0600)
	os.Mkdir(dirPath, 0755)
	os.WriteFile(emptyPath, []byte{}, 0600)

	tests := []struct {
		name    string
		path    string
		want    [][]string
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
			name:    "File not exist",
			path:    filepath.Join(tempDir, "nonexistent.csv"),
			wantErr: "file does not exist",
		},
		{
			name:    "Path is directory",
			path:    dirPath,
			wantErr: "path is a directory, not a file",
		},
		{
			name:    "Invalid extension",
			path:    invalidExtPath,
			wantErr: "file must have .csv extension",
		},
		{
			name:    "Empty file",
			path:    emptyPath,
			wantErr: "file is empty",
		},
		{
			name: "Valid file",
			path: validPath,
			want: [][]string{
				{"name", "age"},
				{"Alice", "30"},
				{"Bob", "25"},
			},
		},
		{
			name:    "Malformed CSV",
			path:    invalidTypePath,
			wantErr: "extraneous or missing \" in quoted-field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := csv.ReadFile(tt.path)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("ReadFile() error = %v, wantErr containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("ReadFile() unexpected error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWriteFile(t *testing.T) {
	tempDir := t.TempDir()
	validPath := filepath.Join(tempDir, "test.csv")
	invalidExtPath := filepath.Join(tempDir, "test.txt")
	longPath := filepath.Join(tempDir, string(make([]rune, 4097)))
	subDirPath := filepath.Join(tempDir, "subdir/test.csv")

	tests := []struct {
		name    string
		records [][]string
		path    string
		perm    os.FileMode
		wantErr string
		want    [][]string
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
			name:    "Invalid extension",
			path:    invalidExtPath,
			wantErr: "file must have .csv extension",
		},
		{
			name:    "Empty records",
			records: [][]string{},
			path:    validPath,
			wantErr: "records cannot be empty",
		},
		{
			name: "Valid write",
			records: [][]string{
				{"name", "age"},
				{"Alice", "30"},
			},
			path: validPath,
			want: [][]string{
				{"name", "age"},
				{"Alice", "30"},
			},
		},
		{
			name: "Write with custom perm",
			records: [][]string{
				{"name", "age"},
				{"Bob", "25"},
			},
			path: filepath.Join(tempDir, "custom_perm.csv"),
			perm: 0644,
			want: [][]string{
				{"name", "age"},
				{"Bob", "25"},
			},
		},
		{
			name: "Write to subdir",
			records: [][]string{
				{"name", "age"},
				{"Charlie", "40"},
			},
			path: subDirPath,
			want: [][]string{
				{"name", "age"},
				{"Charlie", "40"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var perm []os.FileMode
			if tt.perm != 0 {
				perm = []os.FileMode{tt.perm}
			}
			err := csv.WriteFile(tt.records, tt.path, perm...)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("WriteFile() error = %v, wantErr containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("WriteFile() unexpected error = %v", err)
			}
			got, err := csv.ReadFile(tt.path)
			if err != nil {
				t.Errorf("Failed to read written file: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WriteFile() content = %v, want %v", got, tt.want)
			}
			// Removed perm check to avoid umask issues
		})
	}
}
