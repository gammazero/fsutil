package fsutil_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/gammazero/fsutil"
)

func TestDirEmpty(t *testing.T) {
	_, err := fsutil.DirEmpty("")
	requireError(t, err)

	tmpDir := t.TempDir()
	_, err = fsutil.DirEmpty(filepath.Join(tmpDir, "nosuchdir"))
	requireError(t, err)

	empty, err := fsutil.DirEmpty(tmpDir)
	requireNoError(t, err)
	if !empty {
		t.Fatal("expected empty directory")
	}

	file, err := os.CreateTemp(tmpDir, "")
	requireNoError(t, err)
	requireNoError(t, file.Close())

	empty, err = fsutil.DirEmpty(tmpDir)
	requireNoError(t, err)
	if empty {
		t.Fatal("expected non-empty directory")
	}
}

func TestDirExists(t *testing.T) {
	_, err := fsutil.DirExists("")
	requireError(t, err)

	tmpDir := t.TempDir()
	exists, err := fsutil.DirExists(tmpDir)
	requireNoError(t, err)
	if !exists {
		t.Fatal("expected directory to exist")
	}

	notDir := filepath.Join(tmpDir, "nosuchdir")
	exists, err = fsutil.DirExists(notDir)
	requireNoError(t, err)
	if exists {
		t.Fatal("expected directory to not exist")
	}

	file, err := os.CreateTemp(t.TempDir(), "")
	requireNoError(t, err)
	requireNoError(t, file.Close())

	_, err = fsutil.DirExists(file.Name())
	requireErrorContains(t, err, "not a directory")

	// If running on Windows, skip write-only directory tests.
	if runtime.GOOS == "windows" {
		t.SkipNow()
	}

	requireNoError(t, os.Chmod(tmpDir, 0222))
	_, err = fsutil.DirExists(notDir)
	requireError(t, err)

	requireNoError(t, os.Chmod(tmpDir, 0777))
}

func TestDirWritable(t *testing.T) {
	err := fsutil.DirWritable("")
	requireError(t, err)

	tmpDir := t.TempDir()

	wrDir := filepath.Join(tmpDir, "readwrite")
	err = fsutil.DirWritable(wrDir)
	requireNoError(t, err)

	// Check that DirWritable created directory.
	fi, err := os.Stat(wrDir)
	requireNoError(t, err)
	if !fi.IsDir() {
		t.Fatal("expected IsDir to return true")
	}

	err = fsutil.DirWritable(wrDir)
	requireNoError(t, err)

	// Check that DirWritable returns error for non-directory.
	file, err := os.CreateTemp(tmpDir, "")
	requireNoError(t, err)
	requireNoError(t, file.Close())
	err = fsutil.DirWritable(file.Name())
	requireErrorContains(t, err, "not a directory")

	// If running on Windows, skip read-only directory tests.
	if runtime.GOOS == "windows" {
		t.SkipNow()
	}

	roDir := filepath.Join(tmpDir, "readonly")
	if err = os.Mkdir(roDir, 0500); err != nil {
		panic(err)
	}

	err = fsutil.DirWritable(roDir)
	if !errors.Is(err, fs.ErrPermission) {
		t.Fatal("wrong error")
	}

	roChild := filepath.Join(roDir, "child")
	err = fsutil.DirWritable(roChild)
	if !errors.Is(err, fs.ErrPermission) {
		t.Fatal("expected permission error, got", err)
	}
}

func TestFileChanged(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "")
	requireNoError(t, err)
	requireNoError(t, file.Close())

	var modTime time.Time
	var changed bool
	modTime, changed, err = fsutil.FileChanged(file.Name(), modTime)
	requireNoError(t, err)
	if !changed {
		t.Fatal("expected changed to be true")
	}
	if modTime.IsZero() {
		t.Fatal("expected modification time to be non-zero")
	}

	var newModTime time.Time
	newModTime, changed, err = fsutil.FileChanged(file.Name(), modTime)
	requireNoError(t, err)
	if changed {
		t.Fatal("expected changed to be false")
	}
	if newModTime != modTime {
		t.Fatal("expected newModTime to be same as modTime")
	}

	_, _, err = fsutil.FileChanged(filepath.Join(t.TempDir(), "nosuchfile"), modTime)
	requireError(t, err)
}

func TestFileExists(t *testing.T) {
	fileName := filepath.Join(t.TempDir(), "somefile")
	if fsutil.FileExists(fileName) {
		t.Fatal("expected file to not exist")
	}

	file, err := os.Create(fileName)
	requireNoError(t, err)
	requireNoError(t, file.Close())

	if !fsutil.FileExists(fileName) {
		t.Fatal("expected file to exist")
	}
}

func TestExpandHome(t *testing.T) {
	dir, err := fsutil.ExpandHome("")
	requireNoError(t, err)
	if dir != "" {
		t.Fatal("expected dir to be empty string")
	}

	origDir := filepath.Join("somedir", "somesub")
	dir, err = fsutil.ExpandHome(origDir)
	requireNoError(t, err)
	if dir != origDir {
		t.Fatal("expected dir to be same as origDir")
	}

	_, err = fsutil.ExpandHome(filepath.FromSlash("~nosuchuser/somedir"))
	requireError(t, err)

	homeEnv := "HOME"
	if runtime.GOOS == "windows" {
		homeEnv = "USERPROFILE"
	}
	homeDir := filepath.Join(t.TempDir(), "testhome")
	t.Setenv(homeEnv, homeDir)

	const subDir = "mytmp"
	origDir = filepath.Join("~", subDir)
	dir, err = fsutil.ExpandHome(origDir)
	requireNoError(t, err)
	expect := filepath.Join(homeDir, subDir)
	if dir != expect {
		t.Fatalf("expected dir to be %q, got %q", expect, dir)
	}

	requireNoError(t, os.Unsetenv(homeEnv))
	_, err = fsutil.ExpandHome(origDir)
	requireError(t, err)
}

func requireError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error")
	}
}

func requireNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal("expected no error")
	}
}

func requireErrorContains(t *testing.T, err error, s string) {
	t.Helper()
	requireError(t, err)
	if !strings.Contains(err.Error(), s) {
		t.Fatalf("error does not contain %q, error is %q", s, err)
	}
}
