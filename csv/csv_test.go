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
	unicodePath := filepath.Join(tempDir, "unicode.csv")
	invalidTypePath := filepath.Join(tempDir, "invalid.csv")
	invalidExtPath := filepath.Join(tempDir, "test.txt")
	dirPath := filepath.Join(tempDir, "dir")
	emptyPath := filepath.Join(tempDir, "empty.csv")
	largePath := filepath.Join(tempDir, "large.csv")
	nonexistentPath := filepath.Join(tempDir, "nonexistent.csv")
	longPath := filepath.Join(tempDir, string(make([]rune, 4097)))

	// Setup test files
	os.WriteFile(validPath, []byte("name,age\nAlice,30\nBob,25\n"), 0600)
	os.WriteFile(unicodePath, []byte("名字,年龄\n张伟,30\n李娜,25\n"), 0600)
	os.WriteFile(invalidTypePath, []byte("name,age\nAlice,\"30\nBob,25\n"), 0600)
	os.WriteFile(invalidExtPath, []byte("dummy"), 0600)
	os.Mkdir(dirPath, 0755)
	os.WriteFile(emptyPath, []byte{}, 0600)
	// Create a large file (1MB of CSV data)
	largeData := strings.Repeat("a,b\n", 1024*1024/4) // ~1MB
	os.WriteFile(largePath, []byte(largeData), 0600)

	tests := []struct {
		name    string
		path    string
		dest    any
		want    any
		wantErr string
	}{
		{
			name:    "Empty path",
			path:    "",
			dest:    &[][]string{},
			wantErr: "path cannot be empty or root",
		},
		{
			name:    "Root path",
			path:    ".",
			dest:    &[][]string{},
			wantErr: "path cannot be empty or root",
		},
		{
			name:    "Path too long",
			path:    longPath,
			dest:    &[][]string{},
			wantErr: "path too long",
		},
		{
			name:    "File not exist",
			path:    nonexistentPath,
			dest:    &[][]string{},
			wantErr: "file does not exist",
		},
		{
			name:    "Path is directory",
			path:    dirPath,
			dest:    &[][]string{},
			wantErr: "path is a directory, not a file",
		},
		{
			name:    "Invalid extension",
			path:    invalidExtPath,
			dest:    &[][]string{},
			wantErr: "file must have .csv extension",
		},
		{
			name:    "Empty file",
			path:    emptyPath,
			dest:    &[][]string{},
			wantErr: "file is empty",
		},
		{
			name: "Valid file",
			path: validPath,
			dest: &[][]string{},
			want: &[][]string{{"name", "age"}, {"Alice", "30"}, {"Bob", "25"}},
		},
		{
			name: "Unicode file",
			path: unicodePath,
			dest: &[][]string{},
			want: &[][]string{{"名字", "年龄"}, {"张伟", "30"}, {"李娜", "25"}},
		},
		{
			name:    "Invalid CSV",
			path:    invalidTypePath,
			dest:    &[][]string{},
			wantErr: "extraneous or missing \" in quoted-field",
		},
		{
			name:    "Invalid dest type",
			path:    validPath,
			dest:    &struct{}{},
			wantErr: "destination must be *[][]string",
		},
		{
			name: "Large file",
			path: largePath,
			dest: &[][]string{},
			want: func() *[][]string {
				records := make([][]string, 1024*1024/4)
				for i := range records {
					records[i] = []string{"a", "b"}
				}
				return &records
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := csv.ReadFile(tt.path, tt.dest)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("ReadFile() error = %v, wantErr containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("ReadFile() unexpected error = %v", err)
			}
			if tt.want != nil && !reflect.DeepEqual(tt.dest, tt.want) {
				t.Errorf("ReadFile() dest = %v, want %v", tt.dest, tt.want)
			}
		})
	}
}

func TestWriteFile(t *testing.T) {
	tempDir := t.TempDir()
	validPath := filepath.Join(tempDir, "test.csv")
	unicodePath := filepath.Join(tempDir, "unicode.csv")
	invalidExtPath := filepath.Join(tempDir, "test.txt")
	longPath := filepath.Join(tempDir, string(make([]rune, 4097)))
	subDirPath := filepath.Join(tempDir, "subdir/test.csv")

	tests := []struct {
		name    string
		data    any
		path    string
		perm    os.FileMode
		want    [][]string
		wantErr string
	}{
		{
			name:    "Empty path",
			path:    "",
			data:    [][]string{{"name", "age"}, {"Alice", "30"}},
			wantErr: "path cannot be empty or root",
		},
		{
			name:    "Root path",
			path:    ".",
			data:    [][]string{{"name", "age"}, {"Alice", "30"}},
			wantErr: "path cannot be empty or root",
		},
		{
			name:    "Path too long",
			path:    longPath,
			data:    [][]string{{"name", "age"}, {"Alice", "30"}},
			wantErr: "path too long",
		},
		{
			name:    "Invalid extension",
			path:    invalidExtPath,
			data:    [][]string{{"name", "age"}, {"Alice", "30"}},
			wantErr: "file must have .csv extension",
		},
		{
			name:    "Empty records",
			data:    [][]string{},
			path:    validPath,
			wantErr: "records cannot be empty",
		},
		{
			name:    "Invalid data type",
			data:    "not a [][]string",
			path:    validPath,
			wantErr: "data must be [][]string",
		},
		{
			name: "Valid write",
			data: [][]string{{"name", "age"}, {"Alice", "30"}},
			path: validPath,
			want: [][]string{{"name", "age"}, {"Alice", "30"}},
		},
		{
			name: "Unicode write",
			data: [][]string{{"名字", "年龄"}, {"张伟", "30"}},
			path: unicodePath,
			want: [][]string{{"名字", "年龄"}, {"张伟", "30"}},
		},
		{
			name: "Write with custom perm",
			data: [][]string{{"name", "age"}, {"Bob", "25"}},
			path: filepath.Join(tempDir, "custom_perm.csv"),
			perm: 0644,
			want: [][]string{{"name", "age"}, {"Bob", "25"}},
		},
		{
			name: "Write to subdir",
			data: [][]string{{"name", "age"}, {"Charlie", "40"}},
			path: subDirPath,
			want: [][]string{{"name", "age"}, {"Charlie", "40"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var perm []os.FileMode
			if tt.perm != 0 {
				perm = []os.FileMode{tt.perm}
			}
			err := csv.WriteFile(tt.data, tt.path, perm...)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("WriteFile() error = %v, wantErr containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("WriteFile() unexpected error = %v", err)
			}
			var got [][]string
			err = csv.ReadFile(tt.path, &got)
			if err != nil {
				t.Errorf("Failed to read written file: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WriteFile() content = %v, want %v", got, tt.want)
			}
		})
	}
}

// Existing tests for Marshal and Unmarshal remain unchanged
func TestMarshal(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected []byte
		err      string
	}{
		{
			name:     "valid records",
			input:    [][]string{{"a", "b"}, {"1", "2"}},
			expected: []byte("a,b\n1,2\n"),
			err:      "",
		},
		{
			name:     "empty records",
			input:    [][]string{},
			expected: nil,
			err:      "records cannot be empty",
		},
		{
			name:     "invalid type",
			input:    "not a [][]string",
			expected: nil,
			err:      "data must be [][]string",
		},
		{
			name:     "unicode records",
			input:    [][]string{{"名字", "年龄"}, {"张伟", "30"}},
			expected: []byte("名字,年龄\n张伟,30\n"),
			err:      "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := csv.Marshal(tt.input)
			if tt.err != "" {
				if err == nil || !strings.Contains(err.Error(), tt.err) {
					t.Errorf("Marshal() error = %v, wantErr containing %q", err, tt.err)
				}
				return
			}
			if err != nil {
				t.Errorf("Marshal() unexpected error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("Marshal() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestUnmarshal(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		dest any
		want any
		err  string
	}{
		{
			name: "valid CSV",
			data: []byte("a,b\n1,2\n"),
			dest: &[][]string{},
			want: &[][]string{{"a", "b"}, {"1", "2"}},
			err:  "",
		},
		{
			name: "unicode CSV",
			data: []byte("名字,年龄\n张伟,30\n"),
			dest: &[][]string{},
			want: &[][]string{{"名字", "年龄"}, {"张伟", "30"}},
			err:  "",
		},
		{
			name: "empty data",
			data: []byte{},
			dest: &[][]string{},
			err:  "CSV data cannot be empty",
		},
		{
			name: "invalid dest type",
			data: []byte("a,b\n1,2\n"),
			dest: &struct{}{},
			err:  "destination must be *[][]string",
		},
		{
			name: "malformed CSV",
			data: []byte("a,b\n1,\"2\n"),
			dest: &[][]string{},
			err:  "extraneous or missing \" in quoted-field",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := csv.Unmarshal(tt.data, tt.dest)
			if tt.err != "" {
				if err == nil || !strings.Contains(err.Error(), tt.err) {
					t.Errorf("Unmarshal() error = %v, wantErr containing %q", err, tt.err)
				}
				return
			}
			if err != nil {
				t.Errorf("Unmarshal() unexpected error = %v", err)
			}
			if tt.want != nil && !reflect.DeepEqual(tt.dest, tt.want) {
				t.Errorf("Unmarshal() dest = %v, want %v", tt.dest, tt.want)
			}
		})
	}
}
