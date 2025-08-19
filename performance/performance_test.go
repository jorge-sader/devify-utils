package performance_test

import (
	"errors"
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

func TestBenchmarkWrapper(t *testing.T) {
	tests := []struct {
		name        string
		fn          func() error
		iterations  int
		wantErr     bool
		wantMaxTime float64 // Nanoseconds
	}{
		{
			name: "Valid function",
			fn: func() error {
				time.Sleep(1 * time.Millisecond)
				return nil
			},
			iterations:  10,
			wantMaxTime: float64(2 * time.Millisecond.Nanoseconds()),
			wantErr:     false,
		},
		{
			name: "CSV Marshal",
			fn: func() error {
				_, err := csv.Marshal([][]string{{"name", "age"}, {"Alice", "30"}})
				return err
			},
			iterations:  10,
			wantMaxTime: float64(5 * time.Millisecond.Nanoseconds()),
			wantErr:     false,
		},
		{
			name:        "Zero iterations",
			fn:          func() error { return nil },
			iterations:  0,
			wantErr:     true,
			wantMaxTime: 0,
		},
		{
			name:        "Negative iterations",
			fn:          func() error { return nil },
			iterations:  -1,
			wantErr:     true,
			wantMaxTime: 0,
		},
		{
			name: "Single iteration",
			fn: func() error {
				time.Sleep(1 * time.Millisecond)
				return nil
			},
			iterations:  1,
			wantMaxTime: float64(2 * time.Millisecond.Nanoseconds()),
			wantErr:     false,
		},
		{
			name: "Function with error",
			fn: func() error {
				return errors.New("test error")
			},
			iterations:  10,
			wantErr:     true,
			wantMaxTime: 0,
		},
		{
			name: "Multiple iterations",
			fn: func() error {
				time.Sleep(500 * time.Microsecond)
				return nil
			},
			iterations:  100,
			wantMaxTime: float64(2 * time.Millisecond.Nanoseconds()),
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			avgTime, err := performance.BenchmarkWrapper(tt.fn, tt.iterations)
			if (err != nil) != tt.wantErr {
				t.Errorf("BenchmarkWrapper() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && avgTime > tt.wantMaxTime {
				t.Errorf("BenchmarkWrapper() average time = %v ns, want <= %v ns", avgTime, tt.wantMaxTime)
			}
			if tt.wantErr && avgTime != 0 {
				t.Errorf("BenchmarkWrapper() average time = %v, want 0 on error", avgTime)
			}
		})
	}
}

func TestBenchmarkCSVMarshalLogic(t *testing.T) {
	tests := []struct {
		name        string
		input       any
		want        string
		wantErr     bool
		wantMaxTime float64 // Nanoseconds
	}{
		{
			name:        "Valid records",
			input:       [][]string{{"name", "age"}, {"Alice", "30"}},
			want:        "name,age\nAlice,30\n",
			wantErr:     false,
			wantMaxTime: float64(5 * time.Millisecond.Nanoseconds()),
		},
		{
			name:        "Empty records",
			input:       [][]string{},
			want:        "",
			wantErr:     true,
			wantMaxTime: float64(1 * time.Millisecond.Nanoseconds()),
		},
		{
			name:        "Invalid type",
			input:       "not a [][]string",
			want:        "",
			wantErr:     true,
			wantMaxTime: float64(1 * time.Millisecond.Nanoseconds()),
		},
		{
			name:        "Unicode records",
			input:       [][]string{{"名字", "年龄"}, {"张伟", "30"}},
			want:        "名字,年龄\n张伟,30\n",
			wantErr:     false,
			wantMaxTime: float64(5 * time.Millisecond.Nanoseconds()),
		},
		{
			name: "Large records",
			input: func() [][]string {
				records := make([][]string, 1000)
				for i := range records {
					records[i] = []string{"a", "b"}
				}
				return records
			}(),
			want:        strings.Repeat("a,b\n", 1000),
			wantErr:     false,
			wantMaxTime: float64(50 * time.Millisecond.Nanoseconds()),
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
			// Simulate benchmark loop
			for i := 0; i < 10; i++ {
				_, err := csv.Marshal(tt.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("CSVMarshal loop error = %v, wantErr %v", err, tt.wantErr)
				}
			}
			// Measure to simulate benchmark
			duration := performance.MeasureExecutionTime(func() {
				csv.Marshal(tt.input)
			})
			if float64(duration.Nanoseconds()) > tt.wantMaxTime {
				t.Errorf("CSVMarshal too slow: %v", duration)
			}
			// Simulate BenchmarkWrapper
			avgTime, err := performance.BenchmarkWrapper(func() error {
				_, err := csv.Marshal(tt.input)
				return err
			}, 10)
			if (err != nil) != tt.wantErr {
				t.Errorf("BenchmarkWrapper with CSVMarshal error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && avgTime > tt.wantMaxTime {
				t.Errorf("BenchmarkWrapper with CSVMarshal too slow: %v ns", avgTime)
			}
		})
	}
}

func TestBenchmarkSanitizeFilenameLogic(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		want        string
		wantErr     bool
		wantMaxTime float64 // Nanoseconds
	}{
		{
			name:        "Basic filename",
			input:       "test<file>.txt",
			want:        "test_file_.txt",
			wantErr:     false,
			wantMaxTime: float64(5 * time.Millisecond.Nanoseconds()),
		},
		{
			name:        "Unicode filename",
			input:       "文件<>.文档",
			want:        "文件__.文档",
			wantErr:     false,
			wantMaxTime: float64(5 * time.Millisecond.Nanoseconds()),
		},
		{
			name:        "Empty filename",
			input:       "",
			want:        "",
			wantErr:     true,
			wantMaxTime: float64(1 * time.Millisecond.Nanoseconds()),
		},
		{
			name:        "Reserved filename",
			input:       "CON.txt",
			want:        "",
			wantErr:     true,
			wantMaxTime: float64(1 * time.Millisecond.Nanoseconds()),
		},
		{
			name:        "Long filename",
			input:       strings.Repeat("a", 300) + ".txt",
			want:        strings.Repeat("a", 255-len(".txt")) + ".txt",
			wantErr:     false,
			wantMaxTime: float64(5 * time.Millisecond.Nanoseconds()),
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
			// Simulate benchmark loop
			for i := 0; i < 10; i++ {
				_, err := filesystem.SanitizeFilename(tt.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("SanitizeFilename loop error = %v, wantErr %v", err, tt.wantErr)
				}
			}
			// Measure to simulate benchmark
			duration := performance.MeasureExecutionTime(func() {
				filesystem.SanitizeFilename(tt.input)
			})
			if float64(duration.Nanoseconds()) > tt.wantMaxTime {
				t.Errorf("SanitizeFilename too slow: %v", duration)
			}
			// Simulate BenchmarkWrapper
			avgTime, err := performance.BenchmarkWrapper(func() error {
				_, err := filesystem.SanitizeFilename(tt.input)
				return err
			}, 10)
			if (err != nil) != tt.wantErr {
				t.Errorf("BenchmarkWrapper with SanitizeFilename error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && avgTime > tt.wantMaxTime {
				t.Errorf("BenchmarkWrapper with SanitizeFilename too slow: %v ns", avgTime)
			}
		})
	}
}

func TestBenchmarkSanitizePathLogic(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		allowNav    bool
		want        string
		wantErr     bool
		wantMaxTime float64 // Nanoseconds
	}{
		{
			name:        "Basic path",
			input:       "path/to/<file>.txt",
			allowNav:    false,
			want:        "path/to/file.txt",
			wantErr:     false,
			wantMaxTime: float64(5 * time.Millisecond.Nanoseconds()),
		},
		{
			name:        "Unicode path",
			input:       "路径/到/文件.文档",
			allowNav:    false,
			want:        "路径/到/文件.文档",
			wantErr:     false,
			wantMaxTime: float64(5 * time.Millisecond.Nanoseconds()),
		},
		{
			name:        "Empty path",
			input:       "",
			allowNav:    false,
			want:        "",
			wantErr:     true,
			wantMaxTime: float64(1 * time.Millisecond.Nanoseconds()),
		},
		{
			name:        "Directory path",
			input:       "path/to/dir",
			allowNav:    false,
			want:        "path/to/dir/",
			wantErr:     false,
			wantMaxTime: float64(5 * time.Millisecond.Nanoseconds()),
		},
		{
			name:        "Long path",
			input:       strings.Repeat("a/", 1000) + "file.txt",
			allowNav:    false,
			want:        strings.Repeat("a/", 1000) + "file.txt",
			wantErr:     false,
			wantMaxTime: float64(100 * time.Millisecond.Nanoseconds()),
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
			// Simulate benchmark loop
			for i := 0; i < 10; i++ {
				_, err := sanitize.Path(tt.input, tt.allowNav)
				if (err != nil) != tt.wantErr {
					t.Errorf("SanitizePath loop error = %v, wantErr %v", err, tt.wantErr)
				}
			}
			// Measure to simulate benchmark
			duration := performance.MeasureExecutionTime(func() {
				sanitize.Path(tt.input, tt.allowNav)
			})
			if float64(duration.Nanoseconds()) > tt.wantMaxTime {
				t.Errorf("SanitizePath too slow: %v", duration)
			}
			// Simulate BenchmarkWrapper
			avgTime, err := performance.BenchmarkWrapper(func() error {
				_, err := sanitize.Path(tt.input, tt.allowNav)
				return err
			}, 10)
			if (err != nil) != tt.wantErr {
				t.Errorf("BenchmarkWrapper with SanitizePath error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && avgTime > tt.wantMaxTime {
				t.Errorf("BenchmarkWrapper with SanitizePath too slow: %v ns", avgTime)
			}
		})
	}
}
