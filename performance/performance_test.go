package performance_test

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/devify-me/devify-utils/csv"
	"github.com/devify-me/devify-utils/filesystem"
	"github.com/devify-me/devify-utils/performance"
	"github.com/devify-me/devify-utils/sanitize"
)

func TestMeasureExecutionTime(t *testing.T) {
	duration := performance.MeasureExecutionTime(func() {
		time.Sleep(10 * time.Millisecond)
	})
	if duration < 10*time.Millisecond || duration > 20*time.Millisecond {
		t.Errorf("MeasureExecutionTime() = %v, want ~10ms", duration)
	}
}

func TestBenchmarkCSVMarshalLogic(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    string
		wantErr bool
	}{
		{
			name:    "Valid records",
			input:   [][]string{{"name", "age"}, {"Alice", "30"}},
			want:    "name,age\nAlice,30\n",
			wantErr: false,
		},
		{
			name:    "Empty records",
			input:   [][]string{},
			want:    "",
			wantErr: true,
		},
		{
			name:    "Invalid type",
			input:   "not a [][]string",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := csv.Marshal(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("CSVMarshal error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !strings.Contains(string(data), tt.want) {
				t.Errorf("CSVMarshal produced incorrect output: %s, want %s", string(data), tt.want)
			}
			// Measure to simulate benchmark
			duration := performance.MeasureExecutionTime(func() {
				csv.Marshal(tt.input)
			})
			if duration > 5*time.Millisecond {
				t.Errorf("CSVMarshal too slow: %v", duration)
			}
		})
	}
}

func TestBenchmarkSanitizeFilenameLogic(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "Basic filename",
			input:   "test<file>.txt",
			want:    "test_file_.txt",
			wantErr: false,
		},
		{
			name:    "Unicode filename",
			input:   "文件<>.文档",
			want:    "文件__.文档",
			wantErr: false,
		},
		{
			name:    "Empty filename",
			input:   "",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Reserved filename",
			input:   "CON.txt",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := filesystem.SanitizeFilename(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("SanitizeFilename() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SanitizeFilename() = %q, want %q", got, tt.want)
			}
			// Measure to simulate benchmark
			duration := performance.MeasureExecutionTime(func() {
				filesystem.SanitizeFilename(tt.input)
			})
			if duration > 5*time.Millisecond {
				t.Errorf("SanitizeFilename too slow: %v", duration)
			}
		})
	}
}

func TestBenchmarkSanitizePathLogic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		allowNav bool
		want     string
		wantErr  bool
	}{
		{
			name:     "Basic path",
			input:    "path/to/<file>.txt",
			allowNav: false,
			want:     "path/to/file.txt",
			wantErr:  false,
		},
		{
			name:     "Unicode path",
			input:    "路径/到/文件.文档",
			allowNav: false,
			want:     "路径/到/文件.文档",
			wantErr:  false,
		},
		{
			name:     "Empty path",
			input:    "",
			allowNav: false,
			want:     "",
			wantErr:  true,
		},
		{
			name:     "Directory path",
			input:    "path/to/dir",
			allowNav: false,
			want:     "path/to/dir/",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sanitize.Path(tt.input, tt.allowNav)
			if (err != nil) != tt.wantErr {
				t.Errorf("SanitizePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got = filepath.ToSlash(got) // Normalize for cross-platform
			if got != tt.want {
				t.Errorf("SanitizePath() = %q, want %q", got, tt.want)
			}
			// Measure to simulate benchmark
			duration := performance.MeasureExecutionTime(func() {
				sanitize.Path(tt.input, tt.allowNav)
			})
			if duration > 5*time.Millisecond {
				t.Errorf("SanitizePath too slow: %v", duration)
			}
		})
	}
}
