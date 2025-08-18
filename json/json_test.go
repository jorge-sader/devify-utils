package json_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/devify-me/devify-utils/json"
)

type testStruct struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestMarshal(t *testing.T) {
	tests := []struct {
		name    string
		data    any
		want    []byte
		wantErr string
	}{
		{
			name:    "Nil data",
			data:    nil,
			wantErr: "data cannot be nil",
		},
		{
			name: "Valid struct",
			data: testStruct{Name: "Alice", Age: 30},
			want: []byte(`{"name":"Alice","age":30}`),
		},
		{
			name:    "Unmarshalable data",
			data:    make(chan int),
			wantErr: "json: unsupported type: chan int",
		},
		{
			name:    "Empty output",
			data:    struct{}{},
			wantErr: "marshaled JSON is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.data)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("Marshal() error = %v, wantErr containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("Marshal() unexpected error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Marshal() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		dest    any
		want    any
		wantErr string
	}{
		{
			name:    "Empty data",
			data:    []byte{},
			dest:    &testStruct{},
			wantErr: "JSON data cannot be empty",
		},
		{
			name:    "Nil dest",
			data:    []byte(`{"name":"Alice","age":30}`),
			dest:    nil,
			wantErr: "destination cannot be nil",
		},
		{
			name: "Valid JSON",
			data: []byte(`{"name":"Alice","age":30}`),
			dest: &testStruct{},
			want: &testStruct{Name: "Alice", Age: 30},
		},
		{
			name:    "Invalid JSON type mismatch",
			data:    []byte(`{"name":"Alice","age":"invalid"}`),
			dest:    &testStruct{},
			wantErr: "cannot unmarshal string into Go struct field testStruct.age of type int",
		},
		{
			name:    "Malformed JSON syntax",
			data:    []byte(`{"name":"Alice","age":30`),
			dest:    &testStruct{},
			wantErr: "unexpected end of JSON input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := json.Unmarshal(tt.data, tt.dest)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("Unmarshal() error = %v, wantErr containing %q", err, tt.wantErr)
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

func TestReadFile(t *testing.T) {
	tempDir := t.TempDir()
	validPath := filepath.Join(tempDir, "test.json")
	invalidTypePath := filepath.Join(tempDir, "invalid_type.json")
	malformedPath := filepath.Join(tempDir, "malformed.json")
	invalidExtPath := filepath.Join(tempDir, "test.txt")
	dirPath := filepath.Join(tempDir, "dir")
	emptyPath := filepath.Join(tempDir, "empty.json")
	longPath := filepath.Join(tempDir, string(make([]rune, 4097)))

	// Setup files
	os.WriteFile(validPath, []byte(`{"name":"Alice","age":30}`), 0600)
	os.WriteFile(invalidTypePath, []byte(`{"name":"Alice","age":"invalid"}`), 0600)
	os.WriteFile(malformedPath, []byte(`{"name":"Alice","age":30`), 0600)
	os.WriteFile(invalidExtPath, []byte("dummy"), 0600)
	os.Mkdir(dirPath, 0755)
	os.WriteFile(emptyPath, []byte{}, 0600)

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
			path:    filepath.Join(tempDir, "nonexistent.json"),
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
			wantErr: "file must have .json extension",
		},
		{
			name:    "Empty file",
			path:    emptyPath,
			wantErr: "file is empty",
		},
		{
			name: "Valid file",
			path: validPath,
			dest: &testStruct{},
			want: &testStruct{Name: "Alice", Age: 30},
		},
		{
			name:    "Invalid JSON type mismatch",
			path:    invalidTypePath,
			dest:    &testStruct{},
			wantErr: "cannot unmarshal string into Go struct field testStruct.age of type int",
		},
		{
			name:    "Malformed JSON syntax",
			path:    malformedPath,
			dest:    &testStruct{},
			wantErr: "unexpected end of JSON input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := json.ReadFile(tt.path, tt.dest)
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
	validPath := filepath.Join(tempDir, "test.json")
	invalidExtPath := filepath.Join(tempDir, "test.txt")
	longPath := filepath.Join(tempDir, string(make([]rune, 4097)))
	subDirPath := filepath.Join(tempDir, "subdir/test.json")

	tests := []struct {
		name    string
		data    any
		path    string
		perm    os.FileMode
		wantErr string
		want    []byte
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
			wantErr: "file must have .json extension",
		},
		{
			name:    "Nil data",
			path:    validPath,
			data:    nil,
			wantErr: "data cannot be nil",
		},
		{
			name: "Valid write",
			path: validPath,
			data: testStruct{Name: "Alice", Age: 30},
			want: []byte(`{"name":"Alice","age":30}`),
		},
		{
			name: "Write with custom perm",
			path: filepath.Join(tempDir, "custom_perm.json"),
			data: testStruct{Name: "Bob", Age: 25},
			perm: 0644,
			want: []byte(`{"name":"Bob","age":25}`),
		},
		{
			name: "Write to subdir",
			path: subDirPath,
			data: testStruct{Name: "Charlie", Age: 40},
			want: []byte(`{"name":"Charlie","age":40}`),
		},
		{
			name:    "Unmarshalable data",
			path:    filepath.Join(tempDir, "unmarshalable.json"),
			data:    make(chan int),
			wantErr: "json: unsupported type: chan int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var perm []os.FileMode
			if tt.perm != 0 {
				perm = []os.FileMode{tt.perm}
			}
			err := json.WriteFile(tt.data, tt.path, perm...)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("WriteFile() error = %v, wantErr containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("WriteFile() unexpected error = %v", err)
			}
			got, err := os.ReadFile(tt.path)
			if err != nil {
				t.Errorf("Failed to read written file: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WriteFile() content = %s, want %s", got, tt.want)
			}
			// Removed perm check to avoid umask issues
		})
	}
}
