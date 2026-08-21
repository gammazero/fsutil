package atomicfile

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

// File is an os.File that does an atomic rename when [Close] is called.
type File struct {
	*os.File
	path string
}

// Create creates a new temporary file in the same directory as path, with a
// generated name, opens the file for reading and writing with the given mode,
// and returns the resulting file. The temporary file is renamed to the given
// path when [Close] is called.
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

// CloseSync syncs the file contents to storage, closes the file and renames
// it to the final name, and then syncs the directory containing the file.
// This makes both the file contents and the rename durable across a crash,
// which [Close] alone does not guarantee.
func (f *File) CloseSync() error {
	err := f.Sync()
	if err != nil {
		if !errors.Is(err, os.ErrClosed) {
			_ = f.File.Close()
			_ = os.Remove(f.TempName())
		}
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return syncDir(filepath.Dir(f.path))
}

// WriteFile atomically writes data to the named file. The data is written to
// a temporary file that is synced to storage and then renamed to path, and
// the containing directory is synced as well. The file is never observed
// partially written: either it contains all of data, durably, or any previous
// file at path is left in place.
func WriteFile(path string, data []byte, mode os.FileMode) error {
	f, err := Create(path, mode)
	if err != nil {
		return err
	}
	if _, err = f.Write(data); err != nil {
		_ = f.Discard()
		return err
	}
	return f.CloseSync()
}

// CopyFile copies the regular file src to dst, preserving the source file's
// permission mode. The copy is written atomically: dst is never observed
// partially written, and any existing file at dst is replaced only after the
// copy is complete and synced to storage.
func CopyFile(src, dst string) error {
	fi, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !fi.Mode().IsRegular() {
		return fmt.Errorf("cannot copy non-regular file: %s", src)
	}

	in, err := os.Open(src) //nolint:gosec
	if err != nil {
		return err
	}
	defer in.Close() //nolint:errcheck

	out, err := Create(dst, fi.Mode().Perm())
	if err != nil {
		return err
	}
	if _, err = io.Copy(out, in); err != nil {
		_ = out.Discard()
		return err
	}
	return out.CloseSync()
}

// syncDir flushes a directory to storage, making a completed rename within
// that directory durable. Directories cannot be synced on Windows, so this is
// a no-op there.
func syncDir(dir string) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	err = d.Sync()
	if cerr := d.Close(); err == nil {
		err = cerr
	}
	return err
}

// Discard closes the temporary file and removes it without renaming it.
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

// TempName returns the temporary name of the file. Calling [Close] renames
// this file to its final name, and [Discard] removes it; either way, no file
// exists at this name afterward.
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
