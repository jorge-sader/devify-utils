package sanitize_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/devify-me/devify-utils/sanitize"
)

func TestString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"happy: basic string", "hello world", "hello world", false},
		{"happy: with unsafe chars", "hello<world>", "hello world", false},
		{"happy: control chars", "hello\x00world", "helloworld", false},
		{"happy: multiple spaces", "hello   world ", "hello world", false},
		{"happy: unicode", "héllo wörld", "héllo wörld", false},
		{"edge: empty", "", "", true},
		{"edge: only unsafe", "<>{}|\\^~", "", true},
		{"edge: only control", "\x00\x01", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sanitize.String(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("String() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHostname(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"happy: domain", "example.com", "example.com", false},
		{"happy: IP", "192.168.0.1", "192.168.0.1", false},
		{"happy: with hyphen", "sub-domain.example.com", "sub-domain.example.com", false},
		{"edge: invalid chars", "example!.com", "", true},
		{"edge: spaces", "example com", "", true},
		{"edge: empty", "", "", true},
		{"edge: unicode", "héllo.com", "", true}, // invalid per regex
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sanitize.Hostname(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Hostname() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Hostname() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtension(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"happy: basic", ".txt", ".txt", false},
		{"happy: no dot", "jpg", ".jpg", false},
		{"happy: unicode", ".文档", ".文档", false},
		{"happy: unsafe chars", ".txt!", ".txt", false},
		{"edge: empty", "", "", true},
		{"edge: only dot", ".", "", true},
		{"edge: multiple dots", "..txt", ".txt", false},
		{"edge: invalid chars", ".txt<>", ".txt", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sanitize.Extension(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Extension() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Extension() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"happy: basic", "file.txt", "file.txt", false},
		{"happy: unicode", "文件.文档", "文件.文档", false},
		{"happy: unsafe chars", "file<>.txt", "file.txt", false},
		{"happy: control chars", "file\x00.txt", "file.txt", false},
		{"happy: multiple underscores", "file__name.txt", "file_name.txt", false},
		{"edge: no ext", "file", "file", false},
		{"edge: empty", "", "", true},
		{"edge: only dot", ".", "", true},
		{"edge: reserved like CON", "CON.txt", "", true},
		{"edge: max length", strings.Repeat("a", 300) + ".txt", strings.Repeat("a", 255-len(".txt")) + ".txt", false},
		{"edge: invalid base", "<>", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sanitize.FileName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("FileName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FileName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDirName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"happy: basic", "dir", "dir", false},
		{"happy: unicode", "目录", "目录", false},
		{"happy: with hyphen", "sub-dir", "sub-dir", false},
		{"happy: hidden prefix", ".dir", "dir_dir", false},
		{"happy: unsafe chars", "dir<>", "dir", false},
		{"edge: empty", "", "", true},
		{"edge: slashes", "/dir/", "dir", false},
		{"edge: multiple underscores", "dir__name", "dir_name", false},
		{"edge: control chars", "dir\x00", "dir", false},
		{"edge: max length", strings.Repeat("a", 300), strings.Repeat("a", 255), false},
		{"edge: only unsafe", "<>", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sanitize.DirName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("DirName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DirName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		allowNav bool
		want     string
		wantErr  bool
	}{
		{"happy: basic dir", "path/to/dir", false, "path/to/dir/", false},
		{"happy: basic file", "path/to/file.txt", false, "path/to/file.txt", false},
		{"happy: unicode", "路径/到/文件.文档", false, "路径/到/文件.文档", false},
		{"happy: with nav allow", "./path/../to/file.txt", true, "./to/file.txt", false},
		{"happy: no nav", "./path/../to/file.txt", false, "to/file.txt", false},
		{"happy: absolute dir", "/absolute/path/to/dir", false, "/absolute/path/to/dir/", false},
		{"happy: absolute file", "/absolute/path/to/file.txt", false, "/absolute/path/to/file.txt", false},
		{"happy: absolute with nav", "/absolute/./path/../to/dir", true, "/absolute/to/dir/", false},
		{"edge: empty", "", false, "", true},
		{"edge: only slashes", "//", false, "", true},
		{"edge: invalid comp", "path/<>/file.txt", false, "path/file.txt", false},
		{"edge: leading dot slash", "./dir", true, "./dir/", false},
		{"edge: leading parent", "../dir", true, "../dir/", false},
		{"edge: max length", strings.Repeat("a/", 2050) + "file.txt", false, "", true},
		{"edge: dir trailing", "dir", false, "dir/", false},
		{"edge: file no trailing", "file.txt", false, "file.txt", false},
		{"edge: absolute invalid comp", "/path/<>/dir", false, "/path/dir/", false},
		{"edge: absolute empty comp", "/path//dir", false, "/path/dir/", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sanitize.Path(tt.input, tt.allowNav)
			if (err != nil) != tt.wantErr {
				t.Errorf("Path() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// Normalize for cross-platform: convert to / for comparison
			got = filepath.ToSlash(got)
			tt.want = filepath.ToSlash(tt.want)
			if got != tt.want {
				t.Errorf("Path() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUrl(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		requireProtocol bool
		want            string
		wantErr         bool
	}{
		{"happy: with protocol", "https://github.com/user/app-repo", true, "https://github.com/user/app-repo", false},
		{"happy: with path", "https://github.com/user/app-repo/path", true, "https://github.com/user/app-repo/path", false},
		{"happy: unicode in path", "https://github.com/user/文件", true, "https://github.com/user/文件", false},
		{"happy: without protocol allowed", "github.com/user/app-repo", false, "github.com/user/app-repo", false},
		{"happy: http protocol", "http://example.com", true, "http://example.com", false},
		{"edge: empty", "", true, "", true},
		{"edge: invalid chars in host", "https://example!.com", true, "", true},
		{"edge: spaces", "https://example com", true, "", true},
		{"edge: no protocol when required", "github.com/user/app-repo", true, "", true},
		{"edge: unicode in host", "https://héllo.com", true, "", true},
		{"edge: only protocol", "https://", true, "", true},
		{"edge: invalid protocol", "ftp://example.com", true, "", true},
		{"edge: control chars", "https://example.com\x00", true, "https://example.com", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sanitize.Url(tt.input, tt.requireProtocol)
			if (err != nil) != tt.wantErr {
				t.Errorf("Url() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Url() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasFileExtension(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"happy: with ext", "file.txt", true},
		{"happy: unicode ext", "文件.文档", true},
		{"edge: no ext", "file", false},
		{"edge: only dot", ".", false},
		{"edge: hidden", ".hidden", false}, // .hidden is base, no ext
		{"edge: multiple dots", "file.tar.gz", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitize.HasFileExtension(tt.input); got != tt.want {
				t.Errorf("HasFileExtension() = %v, want %v", got, tt.want)
			}
		})
	}
}
