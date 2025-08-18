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

// String sanitizes a string by removing control characters and normalizing spaces.
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

// Hostname sanitizes a hostname or IP address.
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

// Extension ensures a file extension is valid and safe (e.g., ".txt", ".文档").
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

// FileName sanitizes a filename to be safe for file systems.
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

// DirName sanitizes a directory name to be safe for file systems.
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

// Path sanitizes a file path to be safe for file systems.
// It removes unsafe characters, normalizes path separators, resolves relative components,
// limits length, allows Unicode letters and numbers in components, and returns an error if the result is empty or invalid.
// If allowNav is true, relative components ('.' and '..') are preserved in the input for resolution and leading ./ or ../ are reattached if needed; otherwise, they are resolved without preserving leading relatives.
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

// Url sanitizes a URL string by removing control characters, trimming whitespace, and validating a basic format.
// It optionally requires a protocol (http:// or https://), defaulting to true.
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
	urlRegex := regexp.MustCompile(`^(https?://)?[a-zA-Z0-9.-]+(\.[a-zA-Z0-9.-]+)*(/[\p{L}\p{N}_./-]*)?$`)
	if !urlRegex.MatchString(result) {
		return "", errors.New("invalid url format")
	}

	// Enforce protocol if required
	if reqProto && !strings.HasPrefix(strings.ToLower(result), "http://") && !strings.HasPrefix(strings.ToLower(result), "https://") {
		return "", errors.New("url must have protocol")
	}

	return result, nil
}

// HasFileExtension checks if the component has a valid file extension.
func HasFileExtension(comp string) bool {
	ext := filepath.Ext(comp)
	return ext != "" && ext != comp
}
