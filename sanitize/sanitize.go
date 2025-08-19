// Package sanitize provides utilities for sanitizing strings, hostnames, file extensions, filenames, directory names, paths, and URLs.
//
// This package offers functions to clean and validate various types of input, ensuring they are safe for use in file systems, network operations, or other contexts.
// It removes unsafe characters, normalizes spaces, and enforces platform-specific constraints (e.g., reserved filenames, path length limits).
// All functions return errors for invalid or empty inputs, making them suitable for robust input validation in the devify-utils library.
package sanitize

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"unicode"
)

// String sanitizes a string by removing control characters, replacing unsafe characters with spaces, and normalizing whitespace.
//
// The function removes all control characters, replaces characters like <, >, {, }, |, \, ^, and ~ with spaces,
// trims leading/trailing spaces, and collapses multiple spaces into a single space. If the resulting string is empty,
// an error is returned.
//
// Example:
//
//	s, err := String("Hello\t<World>  !")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(s) // Prints "Hello World !"
//
// Parameters:
//   - input: The string to sanitize.
//
// Returns:
//   - string: The sanitized string with control characters removed and spaces normalized.
//   - error: An error if the sanitized string is empty.
func String(input string) (string, error) {
	// Remove control characters
	var builder strings.Builder
	for _, r := range input {
		if !unicode.IsControl(r) {
			builder.WriteRune(r)
		}
	}
	result := builder.String()
	// Replace unsafe characters with spaces
	unsafe := regexp.MustCompile(`[<>{}|\\^~]`)
	result = unsafe.ReplaceAllString(result, " ")
	// Normalize spaces
	result = strings.TrimSpace(result)
	result = regexp.MustCompile(`\s+`).ReplaceAllString(result, " ")
	if result == "" {
		return "", errors.New("sanitized string is empty")
	}
	return result, nil
}

// Hostname sanitizes a hostname or IP address to ensure it contains only valid characters.
//
// The function first applies String sanitization to remove control characters and normalize spaces,
// then validates that the result contains only alphanumeric characters, dots, and hyphens, as per
// hostname and IP address conventions. An error is returned if the input is empty or contains invalid characters.
//
// Example:
//
//	h, err := Hostname("example.com")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(h) // Prints "example.com"
//
// Parameters:
//   - input: The hostname or IP address to sanitize.
//
// Returns:
//   - string: The sanitized hostname or IP address.
//   - error: An error if the input is empty or contains invalid characters.
func Hostname(input string) (string, error) {
	result, err := String(input)
	if err != nil {
		return "", err
	}
	// Additional hostname-specific sanitization (e.g., ensure valid chars)
	hostnameRegex := regexp.MustCompile(`^[a-zA-Z0-9.-]+$`)
	if !hostnameRegex.MatchString(result) {
		return "", errors.New("invalid hostname or IP format")
	}
	return result, nil
}

// Extension sanitizes a file extension to ensure it is safe and valid (e.g., ".txt", ".文档").
//
// The function converts the extension to lowercase, removes unsafe characters (keeping Unicode letters, numbers, and dots),
// ensures it starts with a single dot, and removes multiple dots. An error is returned if the sanitized extension is empty or invalid.
//
// Example:
//
//	ext, err := Extension("TXT")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(ext) // Prints ".txt"
//
// Parameters:
//   - ext: The file extension to sanitize (e.g., "txt" or ".txt").
//
// Returns:
//   - string: The sanitized file extension with a leading dot (e.g., ".txt").
//   - error: An error if the sanitized extension is empty or invalid.
func Extension(ext string) (string, error) {
	// Convert to lowercase and trim whitespace
	ext = strings.ToLower(strings.TrimSpace(ext))
	// Remove unsafe characters, allow Unicode letters, numbers, and dot
	safeExt := regexp.MustCompile(`[^\p{L}\p{N}.]`)
	ext = safeExt.ReplaceAllString(ext, "")
	// Ensure it starts with a dot
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	// Remove multiple dots, keep only the last one
	parts := strings.Split(ext, ".")
	if len(parts) > 2 {
		ext = "." + parts[len(parts)-1]
	}
	// Return error if the extension is empty or just a dot
	if ext == "." || ext == "" {
		return "", errors.New("sanitized extension is empty or invalid")
	}
	return ext, nil
}

// FileName sanitizes a filename to ensure it is safe for file systems across Linux, macOS, and Windows.
//
// The function separates the base name and extension, sanitizes the base by removing unsafe characters and control characters,
// checks for reserved filenames (e.g., "CON", "NUL"), and sanitizes the extension using Extension.
// The sanitized filename is limited to 255 characters to comply with common filesystem limits.
// An error is returned if the filename is empty, reserved, or invalid after sanitization.
//
// Example:
//
//	f, err := FileName("my<file>.txt")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(f) // Prints "myfile.txt"
//
// Parameters:
//   - filename: The filename to sanitize.
//
// Returns:
//   - string: The sanitized filename, including the extension if present.
//   - error: An error if the filename is empty, reserved, or invalid after sanitization.
func FileName(filename string) (string, error) {
	// Handle special case for "."
	if filename == "." {
		return "", errors.New("sanitized filename is empty or invalid")
	}
	// Extract extension and base name
	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)
	// Sanitize base name
	base = strings.TrimSpace(base)
	if base == "" {
		return "", errors.New("filename base is empty")
	}
	// Check for reserved filenames
	reservedNames := []string{
		"CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
	}
	if slices.ContainsFunc(reservedNames, func(s string) bool { return strings.EqualFold(base, s) }) {
		return "", errors.New("filename is a reserved name: " + base)
	}
	// Remove unsafe characters from base name, allow Unicode letters, numbers, underscores, and hyphens
	unsafe := regexp.MustCompile(`[^\p{L}\p{N}_-]`)
	base = unsafe.ReplaceAllString(base, "")
	// Remove control characters
	base = strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1
		}
		return r
	}, base)
	// Collapse multiple underscores and trim
	base = regexp.MustCompile(`_+`).ReplaceAllString(base, "_")
	base = strings.Trim(base, "_")
	// Return error if base is empty after sanitization
	if base == "" {
		return "", errors.New("sanitized filename base is empty")
	}
	// Sanitize extension
	var sanitizedExt string
	var err error
	if ext != "" {
		sanitizedExt, err = Extension(ext)
		if err != nil {
			return "", err
		}
	} else {
		sanitizedExt = "" // No extension provided
	}
	// Combine base and extension
	filename = base + sanitizedExt
	// Ensure filename isn't too long (limit to 255 characters, common filesystem limit)
	if len(filename) > 255 {
		filename = filename[:255-len(sanitizedExt)] + sanitizedExt
	}
	// Return error if the result is empty or invalid
	if filename == "" || filename == "." {
		return "", errors.New("sanitized filename is empty or invalid")
	}
	return filename, nil
}

// DirName sanitizes a directory name to ensure it is safe for file systems across Linux, macOS, and Windows.
//
// The function removes unsafe characters, control characters, and leading/trailing slashes, ensures the name
// does not start with a dot (to avoid hidden directories), and collapses multiple underscores.
// The sanitized directory name is limited to 255 characters to comply with common filesystem limits.
// An error is returned if the directory name is empty or invalid after sanitization.
//
// Example:
//
//	d, err := DirName("my/dir<test>")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(d) // Prints "mydirtest"
//
// Parameters:
//   - dirname: The directory name to sanitize.
//
// Returns:
//   - string: The sanitized directory name.
//   - error: An error if the directory name is empty or invalid after sanitization.
func DirName(dirname string) (string, error) {
	// Trim whitespace and remove leading/trailing slashes
	dirname = strings.TrimSpace(dirname)
	dirname = strings.Trim(dirname, "/\\")
	// Return error if empty after trimming
	if dirname == "" {
		return "", errors.New("directory name is empty")
	}
	// Ensure the name doesn't start with a dot (hidden directory)
	if strings.HasPrefix(dirname, ".") {
		dirname = "dir_" + strings.TrimLeft(dirname, ".")
	}
	// Remove unsafe characters, allow Unicode letters, numbers, underscores, and hyphens
	unsafe := regexp.MustCompile(`[^\p{L}\p{N}_-]`)
	dirname = unsafe.ReplaceAllString(dirname, "")
	// Remove control characters
	dirname = strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1
		}
		return r
	}, dirname)
	// Collapse multiple underscores and trim
	dirname = regexp.MustCompile(`_+`).ReplaceAllString(dirname, "_")
	dirname = strings.Trim(dirname, "_")
	// Return error if the result is empty
	if dirname == "" {
		return "", errors.New("sanitized directory name is empty")
	}
	// Ensure directory name isn't too long (limit to 255 characters)
	if len(dirname) > 255 {
		dirname = dirname[:255]
	}
	return dirname, nil
}

// Path sanitizes a file path to ensure it is safe for file systems across Linux, macOS, and Windows.
//
// The function normalizes path separators, sanitizes each component (directories and the optional file),
// and resolves relative components ('.' and '..') using filepath.Clean. If allowNav is true, leading ./ or ../
// is preserved in the output. The function ensures the path is not empty, not root, and does not exceed 4096 characters.
// A trailing separator is added for directory paths. An error is returned if the path is invalid or empty after sanitization.
//
// Example:
//
//	p, err := Path("dir/../file.txt", true)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(p) // Prints "./file.txt" (preserves leading ./ if allowNav is true)
//
// Parameters:
//   - path: The file path to sanitize.
//   - allowNav: If true, preserves leading ./ or ../ in the output; otherwise, resolves them fully.
//
// Returns:
//   - string: The sanitized file path with normalized separators and a trailing separator for directories.
//   - error: An error if the path is empty, invalid, or exceeds the maximum length.
func Path(path string, allowNav bool) (string, error) {
	// Preserve leading ./ or ../ for relative paths
	hasLeadingDotSlash := strings.HasPrefix(path, "./")
	hasLeadingParentSlash := strings.HasPrefix(path, "../")
	// Trim whitespace and normalize separators
	path = strings.TrimSpace(path)
	if path == "" {
		return "", errors.New("path is empty")
	}
	slashedPath := filepath.ToSlash(path)
	isAbs := strings.HasPrefix(slashedPath, "/")
	// Split into components and sanitize each
	components := strings.Split(slashedPath, "/")
	var cleanComponents []string
	for i, comp := range components {
		if comp == "" {
			continue // Skip empties
		}
		var sanitizedComp string
		var err error
		if comp == "." || comp == ".." {
			sanitizedComp = comp // Preserve nav
		} else if i == len(components)-1 && HasFileExtension(comp) {
			sanitizedComp, err = FileName(comp)
		} else {
			sanitizedComp, err = DirName(comp)
		}
		if err != nil {
			continue
		}
		cleanComponents = append(cleanComponents, sanitizedComp)
	}
	// Build and clean relative path
	relativePath := filepath.Join(cleanComponents...)
	relativePath = filepath.Clean(relativePath)
	if relativePath == "." {
		relativePath = ""
	}
	// Build full path
	var finalPath string
	if isAbs {
		finalPath = string(os.PathSeparator) + relativePath
	} else {
		finalPath = relativePath
	}
	// Error if empty or root
	if finalPath == "" || finalPath == "." || finalPath == string(os.PathSeparator) {
		return "", errors.New("sanitized path is empty")
	}
	// Reattach leading ./ or ../ if applicable
	if allowNav {
		slashedFinal := filepath.ToSlash(finalPath)
		if hasLeadingDotSlash && !strings.HasPrefix(slashedFinal, "./") {
			finalPath = "." + string(os.PathSeparator) + finalPath
		} else if hasLeadingParentSlash && !strings.HasPrefix(slashedFinal, "../") {
			finalPath = ".." + string(os.PathSeparator) + finalPath
		}
	}
	// Ensure path isn't too long
	if len(finalPath) > 4096 {
		return "", errors.New("sanitized path exceeds maximum length")
	}
	// Add trailing separator for directories
	if !HasFileExtension(filepath.Base(finalPath)) {
		finalPath += string(os.PathSeparator)
	}
	return finalPath, nil
}

// Url sanitizes a URL string by removing control characters, trimming whitespace, and validating its format.
//
// The function ensures the URL contains valid characters and optionally requires a protocol (http:// or https://).
// By default, a protocol is required unless requireProtocol is set to false. An error is returned if the URL is empty
// or does not match the expected format (optional protocol, alphanumeric hostname, and optional path with safe characters).
//
// Example:
//
//	u, err := Url("https://example.com/path", true)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(u) // Prints "https://example.com/path"
//
// Parameters:
//   - input: The URL string to sanitize.
//   - requireProtocol: Optional boolean indicating if a protocol (http:// or https://) is required (defaults to true).
//
// Returns:
//   - string: The sanitized URL string.
//   - error: An error if the URL is empty, has an invalid format, or lacks a required protocol.
func Url(input string, requireProtocol ...bool) (string, error) {
	reqProto := true
	if len(requireProtocol) > 0 {
		reqProto = requireProtocol[0]
	}
	// Remove control characters
	var builder strings.Builder
	for _, r := range input {
		if !unicode.IsControl(r) {
			builder.WriteRune(r)
		}
	}
	result := strings.TrimSpace(builder.String())
	if result == "" {
		return "", errors.New("sanitized url is empty")
	}
	// Basic URL validation regex: optional protocol, host (alphanum.-), optional path (Unicode letters/numbers/_-./)
	urlRegex := regexp.MustCompile(`^(https?://)?[a-zA-Z0-9.-]+(\.[a-zA-Z0-9.-]+)*(/[\p{L}\p{N}_./-]*(\?[\p{L}\p{N}_./&=?-]*)?)?$`)
	if !urlRegex.MatchString(result) {
		return "", errors.New("invalid url format")
	}
	// Enforce protocol if required
	if reqProto && !strings.HasPrefix(strings.ToLower(result), "http://") && !strings.HasPrefix(strings.ToLower(result), "https://") {
		return "", errors.New("url must have protocol")
	}
	return result, nil
}

// HasFileExtension checks if the provided string has a valid file extension.
//
// A valid extension is a non-empty suffix starting with a dot (e.g., ".txt").
// The function returns true if the string has an extension and false otherwise.
//
// Example:
//
//	if HasFileExtension("document.txt") {
//	    fmt.Println("Has valid extension")
//	} else {
//	    fmt.Println("No valid extension")
//	}
//
// Parameters:
//   - comp: The string to check for a file extension.
//
// Returns:
//   - bool: True if the string has a valid file extension, false otherwise.
func HasFileExtension(comp string) bool {
	ext := filepath.Ext(comp)
	return ext != "" && ext != comp
}
