// Package performance provides utilities for measuring performance of functions in devify-utils.
// It includes functions to measure execution time and is designed to support benchmarking and profiling.
package performance

import (
	"testing"
	"time"

	"github.com/devify-me/devify-utils/csv"
	"github.com/devify-me/devify-utils/filesystem"
	"github.com/devify-me/devify-utils/sanitize"
)

// Performance is a placeholder struct for future performance-related utilities.
type Performance struct{}

// MeasureExecutionTime measures the time taken to execute a function.
// It returns the duration in nanoseconds.
func MeasureExecutionTime(fn func()) time.Duration {
	start := time.Now()
	fn()
	return time.Since(start)
}

// BenchmarkCSVMarshal benchmarks the csv.Marshal function.
func BenchmarkCSVMarshal(b *testing.B) {
	records := [][]string{{"name", "age"}, {"Alice", "30"}, {"Bob", "25"}}
	for i := 0; i < b.N; i++ {
		csv.Marshal(records)
	}
}

// BenchmarkSanitizeFilename benchmarks the filesystem.SanitizeFilename function.
func BenchmarkSanitizeFilename(b *testing.B) {
	for i := 0; i < b.N; i++ {
		filesystem.SanitizeFilename("test<file>.txt")
	}
}

// BenchmarkSanitizePath benchmarks the sanitize.Path function.
func BenchmarkSanitizePath(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sanitize.Path("path/to/<file>.txt", false)
	}
}
