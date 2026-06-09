package atomicfile

import (
	"errors"
	"os"
	"path/filepath"
)

// File is an os.File that does an atomic rename when [Close] is called.
type File struct {
	*os.File
	path string
}

// Create creates a new temporary file at the given path, opens the file for
// reading and writing, and returns the resulting file. The temporary file is
// renamed to the given path when [Close] is called.
func Create(path string, mode os.FileMode) (*File, error) {
	f, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+"-")
	if err != nil {
		return nil, err
	}
	if err = os.Chmod(f.Name(), mode); err != nil {
		_ = f.Close()
		_ = os.Remove(f.Name())
		return nil, err
	}
	return &File{
		File: f,
		path: path,
	}, nil
}

// Close closes the file and renames it to the final name.
func (f *File) Close() error {
	err := f.File.Close()
	if err != nil {
		if errors.Is(err, os.ErrClosed) {
			_ = os.Remove(f.TempName())
		}
		return err
	}

	return os.Rename(f.TempName(), f.Name())
}

// Discard closes the temproary file and removes it without renaming it.
func (f *File) Discard() error {
	if err := f.File.Close(); err != nil {
		_ = os.Remove(f.TempName())
		return err
	}
	return os.Remove(f.TempName())
}

// Name returns the final name of the file.
func (f *File) Name() string {
	return f.path
}

// TempName returns the temporary name of the file.
func (f *File) TempName() string {
	return f.File.Name()
}
