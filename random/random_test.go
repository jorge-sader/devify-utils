package random_test

import (
	"math"
	"regexp"
	"testing"

	"github.com/devify-me/devify-utils/random"
)

func TestString(t *testing.T) {
	tests := []struct {
		name            string
		n               int
		validCharacters []string
		wantLen         int
		wantRegex       *regexp.Regexp
		wantEmpty       bool
	}{
		{"happy: default chars", 10, nil, 10, regexp.MustCompile(`^[a-zA-Z0-9_+]{10}$`), false},
		{"happy: custom chars", 5, []string{"abc"}, 5, regexp.MustCompile(`^[abc]{5}$`), false},
		{"edge: n=0", 0, nil, 0, nil, true},
		{"edge: n<0", -1, nil, 0, nil, true},
		{"edge: empty chars", 10, []string{""}, 0, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := random.String(tt.n, tt.validCharacters...)
			if len(got) != tt.wantLen {
				t.Errorf("String() len = %d, want %d", len(got), tt.wantLen)
			}
			if tt.wantEmpty && got != "" {
				t.Errorf("String() = %q, want empty", got)
			}
			if tt.wantRegex != nil && !tt.wantRegex.MatchString(got) {
				t.Errorf("String() = %q, does not match regex %s", got, tt.wantRegex)
			}
		})
	}
	// Variance check: run multiple times for uniqueness
	set := make(map[string]bool)
	for i := 0; i < 100; i++ {
		got := random.String(10)
		if set[got] {
			t.Errorf("String() duplicate in 100 runs: %q", got)
		}
		set[got] = true
	}
}

func TestInt(t *testing.T) {
	tests := []struct {
		name     string
		min      int
		max      int
		wantErr  bool
		checkRun int // times to run for range check
	}{
		{"happy: min=max", 5, 5, false, 1},
		{"happy: range", 1, 10, false, 100},
		{"edge: min>max", 10, 1, true, 0},
		{"edge: large range", -100, 100, false, 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < tt.checkRun; i++ {
				got, err := random.Int(tt.min, tt.max)
				if (err != nil) != tt.wantErr {
					t.Errorf("Int() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !tt.wantErr && (got < tt.min || got > tt.max) {
					t.Errorf("Int() = %d, out of range [%d, %d]", got, tt.min, tt.max)
				}
			}
		})
	}
}

func TestHex(t *testing.T) {
	tests := []struct {
		name      string
		n         int
		wantLen   int
		wantRegex *regexp.Regexp
		wantErr   bool
	}{
		{"happy: even", 4, 4, regexp.MustCompile(`^[0-9a-f]{4}$`), false},
		{"happy: odd", 5, 5, regexp.MustCompile(`^[0-9a-f]{5}$`), false},
		{"edge: n=0", 0, 0, nil, false},
		{"edge: n<0", -1, 0, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := random.Hex(tt.n)
			if (err != nil) != tt.wantErr {
				t.Errorf("Hex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantLen {
				t.Errorf("Hex() len = %d, want %d", len(got), tt.wantLen)
			}
			if tt.wantRegex != nil && !tt.wantRegex.MatchString(got) {
				t.Errorf("Hex() = %q, does not match regex %s", got, tt.wantRegex)
			}
		})
	}
	// Variance check
	set := make(map[string]bool)
	for i := 0; i < 100; i++ {
		got, _ := random.Hex(10)
		if set[got] {
			t.Errorf("Hex() duplicate in 100 runs: %q", got)
		}
		set[got] = true
	}
}

func TestBase64(t *testing.T) {
	tests := []struct {
		name      string
		n         int
		wantLen   int
		wantRegex *regexp.Regexp
		wantErr   bool
	}{
		{"happy: multiple of 4", 4, 4, regexp.MustCompile(`^[A-Za-z0-9+/]{4}$`), false},
		{"happy: not multiple", 5, 5, regexp.MustCompile(`^[A-Za-z0-9+/]{5}$`), false},
		{"edge: n=0", 0, 0, nil, false},
		{"edge: n<0", -1, 0, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := random.Base64(tt.n)
			if (err != nil) != tt.wantErr {
				t.Errorf("Base64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantLen {
				t.Errorf("Base64() len = %d, want %d", len(got), tt.wantLen)
			}
			if tt.wantRegex != nil && !tt.wantRegex.MatchString(got) {
				t.Errorf("Base64() = %q, does not match regex %s", got, tt.wantRegex)
			}
		})
	}
	// Variance check
	set := make(map[string]bool)
	for i := 0; i < 100; i++ {
		got, _ := random.Base64(10)
		if set[got] {
			t.Errorf("Base64() duplicate in 100 runs: %q", got)
		}
		set[got] = true
	}
}

func TestUUID(t *testing.T) {
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	set := make(map[string]bool)
	for i := 0; i < 100; i++ {
		got, err := random.UUID()
		if err != nil {
			t.Errorf("UUID() error = %v", err)
		}
		if !uuidRegex.MatchString(got) {
			t.Errorf("UUID() = %q, invalid format", got)
		}
		if set[got] {
			t.Errorf("UUID() duplicate in 100 runs: %q", got)
		}
		set[got] = true
	}
}

func TestFloat64(t *testing.T) {
	tests := []struct {
		name     string
		min      float64
		max      float64
		wantErr  bool
		checkRun int
	}{
		{"happy: min=max", 5.0, 5.0, false, 1},
		{"happy: range", 1.0, 10.0, false, 100},
		{"edge: min>max", 10.0, 1.0, true, 0},
		{"edge: NaN min", math.NaN(), 10.0, true, 0},
		{"edge: NaN max", 1.0, math.NaN(), true, 0},
		{"edge: inf", -math.Inf(1), math.Inf(1), true, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < tt.checkRun; i++ {
				got, err := random.Float64(tt.min, tt.max)
				if (err != nil) != tt.wantErr {
					t.Errorf("Float64() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !tt.wantErr {
					if got < tt.min || got > tt.max {
						t.Errorf("Float64() = %f, out of range [%f, %f]", got, tt.min, tt.max)
					}
					if math.IsNaN(got) {
						t.Errorf("Float64() = %f, unexpected NaN", got)
					}
				}
			}
		})
	}
}

func TestAlphanumeric(t *testing.T) {
	tests := []struct {
		name      string
		n         int
		wantLen   int
		wantRegex *regexp.Regexp
		wantEmpty bool
	}{
		{"happy: positive", 10, 10, regexp.MustCompile(`^[a-zA-Z0-9]{10}$`), false},
		{"edge: n=0", 0, 0, nil, true},
		{"edge: n<0", -1, 0, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := random.Alphanumeric(tt.n)
			if len(got) != tt.wantLen {
				t.Errorf("Alphanumeric() len = %d, want %d", len(got), tt.wantLen)
			}
			if tt.wantEmpty && got != "" {
				t.Errorf("Alphanumeric() = %q, want empty", got)
			}
			if tt.wantRegex != nil && !tt.wantRegex.MatchString(got) {
				t.Errorf("Alphanumeric() = %q, does not match regex %s", got, tt.wantRegex)
			}
		})
	}
	// Variance check
	set := make(map[string]bool)
	for i := 0; i < 100; i++ {
		got := random.Alphanumeric(10)
		if set[got] {
			t.Errorf("Alphanumeric() duplicate in 100 runs: %q", got)
		}
		set[got] = true
	}
}

func TestBoolean(t *testing.T) {
	trueCount := 0
	for i := 0; i < 100; i++ {
		got, err := random.Boolean()
		if err != nil {
			t.Errorf("Boolean() error = %v", err)
		}
		if got {
			trueCount++
		}
	}
	// Rough distribution check (should be around 50)
	if trueCount < 30 || trueCount > 70 {
		t.Errorf("Boolean() distribution skewed: %d true in 100 runs", trueCount)
	}
}

func TestChoice(t *testing.T) {
	tests := []struct {
		name     string
		items    []string
		wantErr  bool
		checkRun int
	}{
		{"happy: single", []string{"a"}, false, 1},
		{"happy: multiple", []string{"a", "b", "c"}, false, 100},
		{"edge: empty", []string{}, true, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := make(map[string]bool)
			for _, item := range tt.items {
				set[item] = true
			}
			for i := 0; i < tt.checkRun; i++ {
				got, err := random.Choice(tt.items)
				if (err != nil) != tt.wantErr {
					t.Errorf("Choice() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !tt.wantErr && !set[got] {
					t.Errorf("Choice() = %q, not in items", got)
				}
			}
		})
	}
}
