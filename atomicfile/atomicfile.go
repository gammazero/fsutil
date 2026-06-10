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
	if err = f.Chmod(mode); err != nil {
		_ = f.Close()
		_ = os.Remove(f.Name())
		return nil, err
	}
	return &File{
		File: f,
		path: path,
	}, nil
}

// Close closes the file and renames it to the final name. Close does not call
// file.Sync, and it is up to the user to do so before calling Close. If
// renaming the file fails, an attempt is made to remove the temporary file; if
// that also fails, both errors are returned.
func (f *File) Close() error {
	err := f.closeTemp()
	if err != nil {
		return err
	}

	if err = os.Rename(f.TempName(), f.Name()); err != nil {
		// Remove temp file on failed Rename.
		if rmErr := os.Remove(f.TempName()); rmErr != nil {
			return errors.Join(err, rmErr)
		}
		return err
	}
	return nil
}

// Discard closes the temproary file and removes it without renaming it.
func (f *File) Discard() error {
	if err := f.closeTemp(); err != nil {
		return err
	}
	return os.Remove(f.TempName())
}

// Name returns the final name of the file. This file will not exist until
// after a successful [Close]. Call [TempName] to get the name of the temporary
// version of this file.
func (f *File) Name() string {
	return f.path
}

// TempName returns the temporary name of the file. Calling [Close] or
// [Discard] removes this file.
func (f *File) TempName() string {
	return f.File.Name()
}

func (f *File) closeTemp() error {
	if err := f.File.Close(); err != nil {
		// Remove temp file on failed close, unless already closed.
		if !errors.Is(err, os.ErrClosed) {
			if rmErr := os.Remove(f.TempName()); rmErr != nil {
				return errors.Join(err, rmErr)
			}
		}
		return err
	}
	return nil
}
