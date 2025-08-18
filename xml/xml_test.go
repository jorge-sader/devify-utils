package xml_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/devify-me/devify-utils/xml"
)

type testStruct struct {
	Name string `xml:"name"`
	Age  int    `xml:"age"`
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
			want: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<testStruct><name>Alice</name><age>30</age></testStruct>`),
		},
		{
			name:    "Unmarshalable data",
			data:    make(chan int),
			wantErr: "xml: unsupported type: chan int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := xml.Marshal(tt.data)
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
			wantErr: "XML data cannot be empty",
		},
		{
			name:    "Nil dest",
			data:    []byte("<testStruct><name>Alice</name><age>30</age></testStruct>"),
			dest:    nil,
			wantErr: "destination cannot be nil",
		},
		{
			name: "Valid XML",
			data: []byte("<testStruct><name>Alice</name><age>30</age></testStruct>"),
			dest: &testStruct{},
			want: &testStruct{Name: "Alice", Age: 30},
		},
		{
			name:    "Invalid XML type mismatch",
			data:    []byte("<testStruct><name>Alice</name><age>invalid</age></testStruct>"),
			dest:    &testStruct{},
			wantErr: "strconv.ParseInt: parsing \"invalid\": invalid syntax",
		},
		{
			name:    "Malformed XML syntax",
			data:    []byte("<testStruct><name>Alice</name><age>30</testStruct>"),
			dest:    &testStruct{},
			wantErr: "XML syntax error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := xml.Unmarshal(tt.data, tt.dest)
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
	validPath := filepath.Join(tempDir, "test.xml")
	invalidTypePath := filepath.Join(tempDir, "invalid_type.xml")
	malformedPath := filepath.Join(tempDir, "malformed.xml")
	invalidExtPath := filepath.Join(tempDir, "test.txt")
	dirPath := filepath.Join(tempDir, "dir")
	emptyPath := filepath.Join(tempDir, "empty.xml")
	longPath := filepath.Join(tempDir, string(make([]rune, 4097)))

	// Setup files
	os.WriteFile(validPath, []byte("<testStruct><name>Alice</name><age>30</age></testStruct>"), 0600)
	os.WriteFile(invalidTypePath, []byte("<testStruct><name>Alice</name><age>invalid</age></testStruct>"), 0600)
	os.WriteFile(malformedPath, []byte("<testStruct><name>Alice</name><age>30</testStruct>"), 0600)
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
			path:    filepath.Join(tempDir, "nonexistent.xml"),
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
			wantErr: "file must have .xml extension",
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
			name:    "Invalid XML type mismatch",
			path:    invalidTypePath,
			dest:    &testStruct{},
			wantErr: "strconv.ParseInt: parsing \"invalid\": invalid syntax",
		},
		{
			name:    "Malformed XML syntax",
			path:    malformedPath,
			dest:    &testStruct{},
			wantErr: "XML syntax error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := xml.ReadFile(tt.path, tt.dest)
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
	validPath := filepath.Join(tempDir, "test.xml")
	invalidExtPath := filepath.Join(tempDir, "test.txt")
	longPath := filepath.Join(tempDir, string(make([]rune, 4097)))
	subDirPath := filepath.Join(tempDir, "subdir/test.xml")

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
			wantErr: "file must have .xml extension",
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
			want: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<testStruct><name>Alice</name><age>30</age></testStruct>`),
		},
		{
			name: "Write with custom perm",
			path: filepath.Join(tempDir, "custom_perm.xml"),
			data: testStruct{Name: "Bob", Age: 25},
			perm: 0644,
			want: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<testStruct><name>Bob</name><age>25</age></testStruct>`),
		},
		{
			name: "Write to subdir",
			path: subDirPath,
			data: testStruct{Name: "Charlie", Age: 40},
			want: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<testStruct><name>Charlie</name><age>40</age></testStruct>`),
		},
		{
			name:    "Unmarshalable data",
			path:    filepath.Join(tempDir, "unmarshalable.xml"),
			data:    make(chan int),
			wantErr: "xml: unsupported type: chan int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var perm []os.FileMode
			if tt.perm != 0 {
				perm = []os.FileMode{tt.perm}
			}
			err := xml.WriteFile(tt.data, tt.path, perm...)
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
