package yaml_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/devify-me/devify-utils/yaml"
)

type testStruct struct {
	Name string `yaml:"name"`
	Age  int    `yaml:"age"`
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
			want:    nil,
			wantErr: "data cannot be nil",
		},
		{
			name: "Valid struct",
			data: testStruct{Name: "Alice", Age: 30},
			want: []byte("name: Alice\nage: 30\n"),
		},
		{
			name:    "Unmarshalable data",
			data:    make(chan int),
			want:    nil,
			wantErr: "cannot marshal type: chan int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := yaml.Marshal(tt.data)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("Marshal() error = %v, wantErr containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("Marshal() unexpected error = %v", err)
			}
			if !strings.EqualFold(string(got), string(tt.want)) {
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
			wantErr: "YAML data cannot be empty",
		},
		{
			name:    "Nil dest",
			data:    []byte("name: Alice\nage: 30\n"),
			dest:    nil,
			wantErr: "destination cannot be nil",
		},
		{
			name: "Valid YAML",
			data: []byte("name: Alice\nage: 30\n"),
			dest: &testStruct{},
			want: &testStruct{Name: "Alice", Age: 30},
		},
		{
			name:    "Invalid YAML type mismatch",
			data:    []byte("name: Alice\nage: invalid\n"),
			dest:    &testStruct{},
			wantErr: "cannot unmarshal !!str `invalid` into int",
		},
		{
			name:    "Malformed YAML syntax",
			data:    []byte(": invalid\n"),
			dest:    &testStruct{},
			wantErr: "did not find expected key",
		},
		{
			name:    "Duplicate key in map",
			data:    []byte("name: Alice\nname: Bob\n"),
			dest:    new(map[string]interface{}),
			wantErr: "mapping key \"name\" already defined",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := yaml.Unmarshal(tt.data, tt.dest)
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
	validYamlPath := filepath.Join(tempDir, "test.yaml")
	validYmlPath := filepath.Join(tempDir, "test.yml")
	invalidYamlPath := filepath.Join(tempDir, "invalid.yaml")
	malformedPath := filepath.Join(tempDir, "malformed.yaml")
	invalidExtPath := filepath.Join(tempDir, "test.txt")
	dirPath := filepath.Join(tempDir, "dir")
	emptyPath := filepath.Join(tempDir, "empty.yaml")
	nonexistentPath := filepath.Join(tempDir, "nonexistent.yaml")
	longPath := filepath.Join(tempDir, string(make([]rune, 4097)))

	// Setup files
	os.WriteFile(validYamlPath, []byte("name: Alice\nage: 30\n"), 0600)
	os.WriteFile(validYmlPath, []byte("name: Alice\nage: 30\n"), 0600)
	os.WriteFile(invalidYamlPath, []byte("name: Alice\nage: invalid\n"), 0600)
	os.WriteFile(malformedPath, []byte(": invalid\n"), 0600)
	os.WriteFile(invalidExtPath, []byte("dummy content\n"), 0600)
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
			dest:    &testStruct{},
			wantErr: "path cannot be empty or root",
		},
		{
			name:    "Root path",
			path:    ".",
			dest:    &testStruct{},
			wantErr: "path cannot be empty or root",
		},
		{
			name:    "Path too long",
			path:    longPath,
			dest:    &testStruct{},
			wantErr: "path too long",
		},
		{
			name:    "File not exist",
			path:    nonexistentPath,
			dest:    &testStruct{},
			wantErr: "file does not exist",
		},
		{
			name:    "Path is directory",
			path:    dirPath,
			dest:    &testStruct{},
			wantErr: "path is a directory, not a file",
		},
		{
			name:    "Invalid extension",
			path:    invalidExtPath,
			dest:    &testStruct{},
			wantErr: "file must have .yaml or .yml extension",
		},
		{
			name:    "Empty file",
			path:    emptyPath,
			dest:    &testStruct{},
			wantErr: "file is empty",
		},
		{
			name: "Valid yaml file",
			path: validYamlPath,
			dest: &testStruct{},
			want: &testStruct{Name: "Alice", Age: 30},
		},
		{
			name: "Valid yml file",
			path: validYmlPath,
			dest: &testStruct{},
			want: &testStruct{Name: "Alice", Age: 30},
		},
		{
			name:    "Invalid YAML type mismatch",
			path:    invalidYamlPath,
			dest:    &testStruct{},
			wantErr: "cannot unmarshal !!str `invalid` into int",
		},
		{
			name:    "Malformed YAML syntax",
			path:    malformedPath,
			dest:    &testStruct{},
			wantErr: "did not find expected key",
		},
		{
			name:    "Nil destination",
			path:    validYamlPath,
			dest:    nil,
			wantErr: "destination cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := yaml.ReadFile(tt.path, tt.dest)
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
	validYamlPath := filepath.Join(tempDir, "test.yaml")
	validYmlPath := filepath.Join(tempDir, "test.yml")
	invalidExtPath := filepath.Join(tempDir, "test.txt")
	longPath := filepath.Join(tempDir, string(make([]rune, 4097)))
	subDirPath := filepath.Join(tempDir, "subdir/test.yaml")

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
			wantErr: "file must have .yaml or .yml extension",
		},
		{
			name:    "Nil data",
			path:    validYamlPath,
			data:    nil,
			wantErr: "data cannot be nil",
		},
		{
			name: "Valid yaml write",
			path: validYamlPath,
			data: testStruct{Name: "Alice", Age: 30},
			want: []byte("name: Alice\nage: 30\n"),
		},
		{
			name: "Valid yml write",
			path: validYmlPath,
			data: testStruct{Name: "Bob", Age: 25},
			want: []byte("name: Bob\nage: 25\n"),
		},
		{
			name: "Write with custom perm",
			path: filepath.Join(tempDir, "custom_perm.yaml"),
			data: testStruct{Name: "Bob", Age: 25},
			perm: 0644,
			want: []byte("name: Bob\nage: 25\n"),
		},
		{
			name: "Write to subdir",
			path: subDirPath,
			data: testStruct{Name: "Charlie", Age: 40},
			want: []byte("name: Charlie\nage: 40\n"),
		},
		{
			name:    "Unmarshalable data",
			path:    filepath.Join(tempDir, "unmarshalable.yaml"),
			data:    make(chan int),
			wantErr: "cannot marshal type: chan int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var perm []os.FileMode
			if tt.perm != 0 {
				perm = []os.FileMode{tt.perm}
			}
			err := yaml.WriteFile(tt.data, tt.path, perm...)
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
			if !strings.EqualFold(string(got), string(tt.want)) {
				t.Errorf("WriteFile() content = %s, want %s", got, tt.want)
			}
			if tt.perm != 0 {
				info, _ := os.Stat(tt.path)
				if info.Mode().Perm() != tt.perm {
					t.Errorf("File perm = %v, want %v", info.Mode().Perm(), tt.perm)
				}
			}
		})
	}
}
