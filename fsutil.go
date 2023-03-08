package fsutil

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"time"
)

// DirEmpty check if a directory is empty.
func DirEmpty(dir string) (bool, error) {
	f, err := os.Open(dir)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	// Not empty or error.
	return false, err
}

// DirExists checks if a directory exists.
func DirExists(dir string) (bool, error) {
	if dir == "" {
		return false, errors.New("directory not specified")
	}

	fi, err := os.Stat(dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	if !fi.IsDir() {
		return false, fmt.Errorf("not a directory: %s", dir)
	}
	return true, nil
}

// DirWritable checks if a directory is writable. If the directory does
// not exist it is created with writable permission.
func DirWritable(dir string) error {
	if dir == "" {
		return errors.New("directory not specified")
	}

	if _, err := os.Stat(dir); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// Directory does not exist, so create it.
			err = os.Mkdir(dir, 0775)
			if err == nil {
				return nil
			}
		}
		if errors.Is(err, fs.ErrPermission) {
			err = fs.ErrPermission
		}
		return fmt.Errorf("directory not writable: %s: %w", dir, err)
	}

	// Directory exists, check that a file can be written.
	file, err := os.CreateTemp(dir, "writetest")
	if err != nil {
		if errors.Is(err, fs.ErrPermission) {
			err = fs.ErrPermission
		}
		return fmt.Errorf("directory not writable: %s: %w", dir, err)
	}
	file.Close()
	return os.Remove(file.Name())
}

// FileChanged returns the modification time of a file and true if different
// from the given time.
func FileChanged(filePath string, modTime time.Time) (time.Time, bool, error) {
	fi, err := os.Stat(filePath)
	if err != nil {
		return modTime, false, fmt.Errorf("cannot stat file %s: %w", filePath, err)
	}
	if fi.ModTime() != modTime {
		return fi.ModTime(), true, nil
	}
	return modTime, false, nil
}

// FileExists returns true if the file exists.
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !errors.Is(err, fs.ErrNotExist)
}
