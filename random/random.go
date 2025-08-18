// Package random provides utilities for generating random strings, integers, floats, booleans, hex, base64, UUIDs, and choices.
//
// This package uses crypto/rand for cryptographically secure randomness, making it suitable for security-sensitive applications.
// It handles edge cases with appropriate error returns for invalid inputs or randomness generation failures.
// All functions are designed to be easy to use and integrate with other devify-utils packages.
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

// String generates a random string of n characters using the provided character set or a default alphanumeric set.
//
// If no validCharacters are provided, the default set includes lowercase and uppercase letters, digits, and the characters
// '_' and '+'. The function returns an empty string if n is negative or if the provided character set is empty.
// If randomness generation fails, it falls back to a deterministic selection to avoid panics.
//
// Example:
//
//	s := String(10) // Uses default alphanumeric set
//	fmt.Println(s)  // Prints a random 10-character string, e.g., "aB7xY9_pQ2"
//	s = String(5, "abc")
//	fmt.Println(s)  // Prints a random 5-character string using 'a', 'b', 'c', e.g., "abcca"
//
// Parameters:
//   - n: The length of the random string to generate.
//   - validCharacters: Optional string of valid characters to use. If empty, defaults to alphanumeric plus '_' and '+'.
//
// Returns:
//   - string: A random string of length n, or an empty string if n is negative or validCharacters is empty.
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

// Int generates a random integer in the range [min, max] (inclusive) using crypto/rand.
//
// The function ensures that min is less than or equal to max, returning an error if this condition is not met.
// It uses cryptographically secure randomness for generating the integer.
//
// Example:
//
//	n, err := Int(1, 10)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(n) // Prints a random integer between 1 and 10, e.g., 7
//
// Parameters:
//   - min: The minimum value of the range (inclusive).
//   - max: The maximum value of the range (inclusive).
//
// Returns:
//   - int: A random integer in the range [min, max].
//   - error: An error if min > max or if randomness generation fails.
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

// Hex generates a random hexadecimal string of n characters (0-9, a-f) using crypto/rand.
//
// The function ensures that n is non-negative and generates the required number of random bytes,
// encoding them as a hexadecimal string. If n is odd, it generates enough bytes to cover the requested length
// and trims the result. An error is returned if randomness generation fails.
//
// Example:
//
//	hexStr, err := Hex(8)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(hexStr) // Prints a random 8-character hex string, e.g., "1a2b3c4d"
//
// Parameters:
//   - n: The length of the hexadecimal string to generate.
//
// Returns:
//   - string: A random hexadecimal string of length n.
//   - error: An error if n is negative or if randomness generation fails.
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

// Base64 generates a random base64-encoded string of n characters (A-Z, a-z, 0-9, +, /) using crypto/rand.
//
// The function ensures that n is non-negative and generates the required number of random bytes,
// encoding them as a base64 string. The result is trimmed to the exact length requested.
// An error is returned if randomness generation fails.
//
// Example:
//
//	b64Str, err := Base64(8)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(b64Str) // Prints a random 8-character base64 string, e.g., "ABcd1234"
//
// Parameters:
//   - n: The length of the base64 string to generate.
//
// Returns:
//   - string: A random base64-encoded string of length n.
//   - error: An error if n is negative or if randomness generation fails.
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

// UUID generates a random UUID (version 4) in the format 8-4-4-4-12 (e.g., "123e4567-e89b-12d3-a456-426614174000").
//
// The function uses the github.com/google/uuid package to generate a cryptographically secure UUID.
// An error is returned if UUID generation fails.
//
// Example:
//
//	id, err := UUID()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(id) // Prints a random UUID, e.g., "123e4567-e89b-12]_d3-a456-426614174000"
//
// Returns:
//   - string: A random UUID string.
//   - error: An error if UUID generation fails.
func UUID() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("failed to generate UUID: %w", err)
	}
	return id.String(), nil
}

// Float64 generates a random float64 in the range [min, max] using crypto/rand.
//
// The function ensures that min is less than or equal to max and that both values are finite and not NaN.
// If min equals max, the function returns min without generating randomness.
// An error is returned if the input constraints are violated or if randomness generation fails.
//
// Example:
//
//	f, err := Float64(0.0, 10.0)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(f) // Prints a random float64 between 0.0 and 10.0, e.g., 7.234
//
// Parameters:
//   - min: The minimum value of the range (inclusive).
//   - max: The maximum value of the range (inclusive).
//
// Returns:
//   - float64: A random float64 in the range [min, max].
//   - error: An error if min > max, values are NaN or infinite, or randomness generation fails.
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

// Alphanumeric generates a random alphanumeric string of n characters (A-Z, a-z, 0-9) using crypto/rand.
//
// The function is a convenience wrapper around String, using a predefined alphanumeric character set.
// It returns an empty string if n is negative.
//
// Example:
//
//	s := Alphanumeric(6)
//	fmt.Println(s) // Prints a random 6-character alphanumeric string, e.g., "Xy7pQ2"
//
// Parameters:
//   - n: The length of the random string to generate.
//
// Returns:
//   - string: A random alphanumeric string of length n, or an empty string if n is negative.
func Alphanumeric(n int) string {
	if n < 0 {
		return ""
	}
	return String(n, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
}

// Boolean generates a random boolean (true or false) using crypto/rand.
//
// The function generates a single random byte and uses its least significant bit to determine the boolean value.
// An error is returned if randomness generation fails.
//
// Example:
//
//	b, err := Boolean()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(b) // Prints true or false randomly
//
// Returns:
//   - bool: A random boolean value (true or false).
//   - error: An error if randomness generation fails.
func Boolean() (bool, error) {
	b := make([]byte, 1)
	_, err := rand.Read(b)
	if err != nil {
		return false, fmt.Errorf("failed to generate random byte: %w", err)
	}
	return b[0]&1 == 1, nil
}

// Choice selects a random element from a slice of strings using crypto/rand.
//
// The function ensures the input slice is not empty and uses the Int function to select a random index.
// An error is returned if the slice is empty or if randomness generation fails.
//
// Example:
//
//	items := []string{"apple", "banana", "cherry"}
//	choice, err := Choice(items)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(choice) // Prints a random item, e.g., "banana"
//
// Parameters:
//   - items: A slice of strings to choose from.
//
// Returns:
//   - string: A randomly selected string from the input slice.
//   - error: An error if the slice is empty or if randomness generation fails.
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
