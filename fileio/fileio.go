package fileio

import (
	"errors"
	"os"
	"path/filepath"
)

type Serializer interface {
	Marshal(data any) ([]byte, error)
	Unmarshal(data []byte, dest any) error
	ReadFile(path string, dest any) error
	WriteFile(data any, path string, perm ...os.FileMode) error
}

var (
	ErrEmptyPath    = errors.New("path cannot be empty or root")
	ErrPathTooLong  = errors.New("path too long")
	ErrFileNotExist = errors.New("file does not exist")
	ErrIsDir        = errors.New("path is a directory, not a file")
)

func ValidatePath(path string, ext string) error {
	if path == "" || path == "." {
		return ErrEmptyPath
	}
	if len(path) > 4096 {
		return ErrPathTooLong
	}
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrFileNotExist
		}
		return err
	}
	if info.IsDir() {
		return ErrIsDir
	}
	if filepath.Ext(path) != ext {
		return errors.New("file must have " + ext + " extension")
	}
	return nil
}

func EnsureDir(path string, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if dir != "." {
		return os.MkdirAll(dir, perm)
	}
	return nil
}
