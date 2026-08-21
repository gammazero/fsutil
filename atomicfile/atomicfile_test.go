package atomicfile_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gammazero/fsutil"
	"github.com/gammazero/fsutil/atomicfile"
)

func TestCreate(t *testing.T) {
	mode := os.FileMode(0666)
	path := filepath.Join(t.TempDir(), "testfile")
	f, err := atomicfile.Create(path, mode)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := f.Discard(); err != nil && !errors.Is(err, os.ErrClosed) {
			t.Fatal(err)
		}
	}()

	if f.Name() != path {
		t.Fatal("expected final name, got:", f.Name())
	}
	if f.TempName() == path {
		t.Fatal("temp name should not equal final name")
	}

	_, err = f.Write([]byte("hello"))
	if err != nil {
		t.Fatal(err)
	}

	if !fsutil.FileExists(f.TempName()) {
		t.Fatalf("temp file should exist: %s", err)
	}
	if fsutil.FileExists(f.Name()) {
		t.Fatal("file should not exist")
	}
	fi, err := os.Stat(f.TempName())
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode() != mode {
		t.Fatal("expected temp file to have mode", mode, "has", fi.Mode())
	}

	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := os.Remove(f.Name())
		if err != nil {
			t.Fatal(err)
		}
	}()

	if fsutil.FileExists(f.TempName()) {
		t.Fatal("temp file should not exist")
	}
	if !fsutil.FileExists(f.Name()) {
		t.Fatalf("file should exist")
	}

	fi, err = os.Stat(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode() != mode {
		t.Fatal("expected final file to have mode", mode, "has", fi.Mode())
	}

	err = f.Close()
	if err == nil || !errors.Is(err, os.ErrClosed) {
		t.Fatal("expected os.ErrClosed on Close after Close")
	}
}

func TestDiscard(t *testing.T) {
	path := filepath.Join(t.TempDir(), "testfile")
	f, err := atomicfile.Create(path, 0666)
	if err != nil {
		t.Fatal(err)
	}

	_, err = f.Write([]byte("hello"))
	if err != nil {
		t.Fatal(err)
	}

	err = f.Discard()
	if err != nil {
		t.Fatal(err)
	}

	if fsutil.FileExists(f.TempName()) {
		t.Fatal("temp file should not exist")
	}
	if fsutil.FileExists(f.Name()) {
		t.Fatal("file should not exist")
	}

	err = f.Discard()
	if err == nil || !errors.Is(err, os.ErrClosed) {
		t.Fatal("expected os.ErrClosed on Discard after Discard")
	}

	err = f.Close()
	if err == nil || !errors.Is(err, os.ErrClosed) {
		t.Fatal("expected os.ErrClosed on Close after Discard")
	}
}

func TestFileExists(t *testing.T) {
	dir := t.TempDir()
	file, err := os.CreateTemp(dir, "somefile")
	if err != nil {
		panic("cannot create temp file")
	}
	if err = file.Close(); err != nil {
		panic(err)
	}
	t.Cleanup(func() {
		_ = os.Remove(file.Name())
	})

	mode := os.FileMode(0644)
	f, err := atomicfile.Create(file.Name(), mode)
	if err != nil {
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Remove(f.Name())
	}()

	fi, err := os.Stat(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode() != mode {
		t.Fatal("expected final file to have mode", mode, "has", fi.Mode())
	}
}

func TestBadDirectory(t *testing.T) {
	dir := t.TempDir()
	name := filepath.Join(dir, "no-such-dir", "my-file")
	f, err := atomicfile.Create(name, 0600)
	if err == nil {
		t.Fatal("expected error creating file in non-existent directory")
	}
	if f != nil {
		t.Fatal("Create should return nil on error")
	}
}

func TestDirExistsAtFilename(t *testing.T) {
	dir := t.TempDir()
	name := filepath.Join(dir, "testblocked")
	err := os.Mkdir(name, 0700)
	if err != nil {
		panic(err)
	}
	t.Cleanup(func() {
		_ = os.Remove(name)
	})

	f, err := atomicfile.Create(name, 0600)
	if err != nil {
		t.Fatal(err)
	}
	err = f.Close()
	if err == nil || !errors.Is(err, os.ErrExist) {
		t.Fatalf("expected error %q, got: %s", os.ErrExist, err)
	}
	if fsutil.FileExists(f.TempName()) {
		t.Fatal("expected temp file to be removed, but it still exists")
	}
	exists, err := fsutil.DirExists(name)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("directory should still exis")
	}
}

func TestWriteFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "testfile")
	data := []byte("hello atomic world")
	mode := os.FileMode(0640)

	if err := atomicfile.WriteFile(path, data, mode); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(data) {
		t.Fatalf("expected content %q, got %q", data, got)
	}

	fi, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode() != mode {
		t.Fatal("expected file to have mode", mode, "has", fi.Mode())
	}

	// No temp files left behind.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 file in directory, found %d", len(entries))
	}

	// Overwrites an existing file.
	data2 := []byte("replaced")
	if err = atomicfile.WriteFile(path, data2, mode); err != nil {
		t.Fatal(err)
	}
	got, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(data2) {
		t.Fatalf("expected content %q, got %q", data2, got)
	}

	// Error when directory does not exist.
	badPath := filepath.Join(dir, "no-such-dir", "testfile")
	if err = atomicfile.WriteFile(badPath, data, mode); err == nil {
		t.Fatal("expected error writing file in non-existent directory")
	}
}

func TestCopyFile(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()
	src := filepath.Join(srcDir, "src")
	dst := filepath.Join(dstDir, "dst")
	data := []byte("copy me")
	mode := os.FileMode(0640)

	if err := os.WriteFile(src, data, mode); err != nil {
		t.Fatal(err)
	}

	if err := atomicfile.CopyFile(src, dst); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(data) {
		t.Fatalf("expected content %q, got %q", data, got)
	}

	// Mode is preserved from the source file.
	fi, err := os.Stat(dst)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode() != mode {
		t.Fatal("expected copied file to have mode", mode, "has", fi.Mode())
	}

	// No temp files left behind.
	entries, err := os.ReadDir(dstDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 file in directory, found %d", len(entries))
	}

	// Overwrites an existing destination.
	data2 := []byte("new content")
	if err = os.WriteFile(src, data2, mode); err != nil {
		t.Fatal(err)
	}
	if err = atomicfile.CopyFile(src, dst); err != nil {
		t.Fatal(err)
	}
	got, err = os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(data2) {
		t.Fatalf("expected content %q, got %q", data2, got)
	}

	// Error when source does not exist.
	err = atomicfile.CopyFile(filepath.Join(srcDir, "nosuchfile"), dst)
	if err == nil || !errors.Is(err, fs.ErrNotExist) {
		t.Fatal("expected fs.ErrNotExist, got:", err)
	}

	// Error when source is not a regular file.
	if err = atomicfile.CopyFile(srcDir, dst); err == nil {
		t.Fatal("expected error copying non-regular file")
	}

	// Failed copy does not disturb an existing destination.
	got, err = os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(data2) {
		t.Fatalf("expected content %q, got %q", data2, got)
	}
}

func TestCloseSync(t *testing.T) {
	path := filepath.Join(t.TempDir(), "testfile")
	f, err := atomicfile.Create(path, 0600)
	if err != nil {
		t.Fatal(err)
	}

	if _, err = f.Write([]byte("hello")); err != nil {
		t.Fatal(err)
	}

	if err = f.CloseSync(); err != nil {
		t.Fatal(err)
	}

	if fsutil.FileExists(f.TempName()) {
		t.Fatal("temp file should not exist")
	}
	if !fsutil.FileExists(f.Name()) {
		t.Fatal("file should exist")
	}
	got, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "hello" {
		t.Fatalf("expected content %q, got %q", "hello", got)
	}

	err = f.CloseSync()
	if err == nil || !errors.Is(err, os.ErrClosed) {
		t.Fatal("expected os.ErrClosed on CloseSync after CloseSync")
	}
	err = f.Close()
	if err == nil || !errors.Is(err, os.ErrClosed) {
		t.Fatal("expected os.ErrClosed on Close after CloseSync")
	}
}

func TestRemoveTempFail(t *testing.T) {
	dir := t.TempDir()
	name := filepath.Join(dir, "testblocked")
	err := os.Mkdir(name, 0700)
	if err != nil {
		panic(err)
	}
	t.Cleanup(func() {
		_ = os.Remove(name)
	})

	f, err := atomicfile.Create(name, 0600)
	if err != nil {
		t.Fatal(err)
	}

	// Cause removal of the temp file to fail.
	err = os.Remove(f.TempName())
	if err != nil {
		t.Fatal(err)
	}
	err = os.Mkdir(f.TempName(), 0700)
	if err != nil {
		panic(err)
	}
	t.Cleanup(func() {
		_ = os.Remove(name)
	})
	// Repeat the same test using a non-empty directory.
	file, err := os.CreateTemp(f.TempName(), "somefile")
	if err != nil {
		panic("cannot create temp file")
	}
	if err = file.Close(); err != nil {
		panic(err)
	}
	t.Cleanup(func() {
		_ = os.Remove(file.Name())
	})

	err = f.Close()
	if err == nil || !errors.Is(err, os.ErrExist) {
		t.Fatalf("expected error %q, got: %s", os.ErrExist, err)
	}

	// Should be multiple errors: failure to rename, failure to remove temp.
	errStr := err.Error()
	if !strings.Contains(errStr, "rename") {
		t.Fatal("missing expected rename error")
	}
	if !strings.Contains(errStr, "remove") {
		t.Fatal("missing expected remove error")
	}

	exists, err := fsutil.DirExists(f.TempName())
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("directory should still exis")
	}
}
