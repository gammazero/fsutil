package atomicfile_test

import (
	"errors"
	"os"
	"path/filepath"
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
	fi, err := os.Stat(name)
	if err != nil {
		t.Fatal(err)
	}
	if !fi.IsDir() {
		t.Fatal("file should not have replaced directory")
	}
}
