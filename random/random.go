// Package random provides utilities for generating random strings, integers, floats, booleans, hex, base64, UUIDs, and choices.
// It uses crypto/rand for secure randomness and handles edge cases with errors where appropriate.
package random

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"

	"github.com/google/uuid"
)

// String generates a random string of n characters using characters from validCharacters.
// If no validCharacters is provided, uses default alphanumeric characters plus _ and +.
// Returns an empty string for invalid inputs (negative n or empty validCharacters).
func String(n int, validCharacters ...string) string {
	if n < 0 {
		return ""
	}

	// Set character set: use provided characters or default
	chars := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890_+")
	if len(validCharacters) > 0 {
		chars = []rune(validCharacters[0])
	}
	if len(chars) == 0 {
		return ""
	}

	s := make([]rune, n)
	for i := range s {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			// In case of error, use a simple modulo-based fallback
			s[i] = chars[i%len(chars)]
			continue
		}
		s[i] = chars[idx.Int64()]
	}
	return string(s)
}

// Int generates a random integer in the range [min, max] (inclusive).
// Returns an error if min > max or if randomness generation fails.
func Int(min, max int) (int, error) {
	if min > max {
		return 0, fmt.Errorf("min (%d) must be less than or equal to max (%d)", min, max)
	}

	// Calculate the range (max - min + 1) to include both min and max
	rangeBig := big.NewInt(int64(max - min + 1))
	// Generate random number in [0, rangeBig)
	n, err := rand.Int(rand.Reader, rangeBig)
	if err != nil {
		return 0, fmt.Errorf("failed to generate random number: %w", err)
	}
	// Shift the result to [min, max]
	return int(n.Int64()) + min, nil
}

// Hex generates a random hexadecimal string of n characters (0-9, a-f).
// Returns an error if n is negative or if randomness generation fails.
func Hex(n int) (string, error) {
	if n < 0 {
		return "", fmt.Errorf("length must be non-negative, got %d", n)
	}
	// Each hex char is 4 bits, so we need ceil(n/2) bytes
	bytesNeeded := (n + 1) / 2
	b := make([]byte, bytesNeeded)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	// Encode to hex and trim to exact length
	hexStr := hex.EncodeToString(b)
	if len(hexStr) > n {
		hexStr = hexStr[:n]
	}
	return hexStr, nil
}

// Base64 generates a random base64-encoded string of n characters (A-Z, a-z, 0-9, +, /).
// Returns an error if n is negative or if randomness generation fails.
func Base64(n int) (string, error) {
	if n < 0 {
		return "", fmt.Errorf("length must be non-negative, got %d", n)
	}
	// Each base64 char encodes 6 bits, so 4 chars encode 3 bytes (24 bits)
	// We need ceil(n*3/4) bytes to get at least n base64 chars
	bytesNeeded := (n*3 + 3) / 4
	b := make([]byte, bytesNeeded)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	// Encode to base64 and trim to exact length
	b64Str := base64.StdEncoding.EncodeToString(b)
	if len(b64Str) > n {
		b64Str = b64Str[:n]
	}
	return b64Str, nil
}

// UUID generates a random UUID (version 4) in the format 8-4-4-4-12.
// Returns an error if UUID generation fails.
func UUID() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("failed to generate UUID: %w", err)
	}
	return id.String(), nil
}

// Float64 generates a random float64 in the range [min, max].
// Returns an error if min > max, values are NaN/inf, or randomness generation fails.
func Float64(min, max float64) (float64, error) {
	if min > max {
		return 0, fmt.Errorf("min (%f) must be less than or equal to max (%f)", min, max)
	}
	if math.IsNaN(min) || math.IsNaN(max) {
		return 0, fmt.Errorf("min and max must not be NaN")
	}
	if math.IsInf(min, 0) || math.IsInf(max, 0) {
		return 0, fmt.Errorf("min and max must be finite")
	}

	if min == max {
		return min, nil
	}

	// Generate a random number in [0, 1]
	maxInt := big.NewInt(1<<53 + 1)
	n, err := rand.Int(rand.Reader, maxInt)
	if err != nil {
		return 0, fmt.Errorf("failed to generate random number: %w", err)
	}
	fraction := float64(n.Int64()) / float64(1<<53)

	// Scale to [min, max]
	return min + fraction*(max-min), nil
}

// Alphanumeric generates a random alphanumeric string of n characters (A-Z, a-z, 0-9).
// Returns an empty string if n is negative.
func Alphanumeric(n int) string {
	if n < 0 {
		return ""
	}
	return String(n, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
}

// Boolean generates a random boolean (true or false).
// Returns an error if randomness generation fails.
func Boolean() (bool, error) {
	b := make([]byte, 1)
	_, err := rand.Read(b)
	if err != nil {
		return false, fmt.Errorf("failed to generate random byte: %w", err)
	}
	return b[0]&1 == 1, nil
}

// Choice selects a random element from a slice of strings.
// Returns an error if the slice is empty or if randomness generation fails.
func Choice(items []string) (string, error) {
	if len(items) == 0 {
		return "", fmt.Errorf("items slice is empty")
	}
	idx, err := Int(0, len(items)-1)
	if err != nil {
		return "", fmt.Errorf("failed to select random index: %w", err)
	}
	return items[idx], nil
}
