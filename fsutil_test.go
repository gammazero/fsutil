package fsutil_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/gammazero/fsutil"
	"github.com/stretchr/testify/require"
)

func TestDirEmpty(t *testing.T) {
	_, err := fsutil.DirEmpty("")
	require.Error(t, err)

	tmpDir := t.TempDir()
	_, err = fsutil.DirEmpty(filepath.Join(tmpDir, "nosuchdir"))
	require.Error(t, err)

	empty, err := fsutil.DirEmpty(tmpDir)
	require.NoError(t, err)
	require.True(t, empty)

	file, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)
	require.NoError(t, file.Close())

	empty, err = fsutil.DirEmpty(tmpDir)
	require.NoError(t, err)
	require.False(t, empty)
}

func TestDirExists(t *testing.T) {
	_, err := fsutil.DirExists("")
	require.Error(t, err)

	tmpDir := t.TempDir()
	exists, err := fsutil.DirExists(tmpDir)
	require.NoError(t, err)
	require.True(t, exists)

	notDir := filepath.Join(tmpDir, "nosuchdir")
	exists, err = fsutil.DirExists(notDir)
	require.NoError(t, err)
	require.False(t, exists)

	file, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)
	require.NoError(t, file.Close())

	_, err = fsutil.DirExists(file.Name())
	require.ErrorContains(t, err, "not a directory")

	// If running on Windows, skip write-only directory tests.
	if runtime.GOOS == "windows" {
		t.SkipNow()
	}

	require.NoError(t, os.Chmod(tmpDir, 0222))
	_, err = fsutil.DirExists(notDir)
	require.Error(t, err)

	require.NoError(t, os.Chmod(tmpDir, 0777))
}

func TestDirWritable(t *testing.T) {
	err := fsutil.DirWritable("")
	require.Error(t, err)

	err = fsutil.DirWritable("~nosuchuser/tmp")
	require.Error(t, err)

	tmpDir := t.TempDir()

	wrDir := filepath.Join(tmpDir, "readwrite")
	err = fsutil.DirWritable(wrDir)
	require.NoError(t, err)

	// Check that DirWritable created directory.
	fi, err := os.Stat(wrDir)
	require.NoError(t, err)
	require.True(t, fi.IsDir())

	err = fsutil.DirWritable(wrDir)
	require.NoError(t, err)

	// Check that DirWritable returns error for non-directory.
	file, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)
	require.NoError(t, file.Close())
	err = fsutil.DirWritable(file.Name())
	require.ErrorContains(t, err, "not a directory")

	// If running on Windows, skip read-only directory tests.
	if runtime.GOOS == "windows" {
		t.SkipNow()
	}

	roDir := filepath.Join(tmpDir, "readonly")
	if err = os.Mkdir(roDir, 0500); err != nil {
		panic(err)
	}

	err = fsutil.DirWritable(roDir)
	require.ErrorIs(t, err, fs.ErrPermission)

	roChild := filepath.Join(roDir, "child")
	err = fsutil.DirWritable(roChild)
	require.ErrorIs(t, err, fs.ErrPermission)
}

func TestFileChanged(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)
	require.NoError(t, file.Close())

	var modTime time.Time
	var changed bool
	modTime, changed, err = fsutil.FileChanged(file.Name(), modTime)
	require.NoError(t, err)
	require.True(t, changed)
	require.False(t, modTime.IsZero())

	var newModTime time.Time
	newModTime, changed, err = fsutil.FileChanged(file.Name(), modTime)
	require.NoError(t, err)
	require.False(t, changed)
	require.Equal(t, modTime, newModTime)

	_, _, err = fsutil.FileChanged(filepath.Join(t.TempDir(), "nosuchfile"), modTime)
	require.Error(t, err)
}

func TestFileExists(t *testing.T) {
	fileName := filepath.Join(t.TempDir(), "somefile")
	require.False(t, fsutil.FileExists(fileName))

	file, err := os.Create(fileName)
	require.NoError(t, err)
	file.Close()

	require.True(t, fsutil.FileExists(fileName))
}

func TestExpand(t *testing.T) {
	dir, err := fsutil.Expand("")
	require.NoError(t, err)
	require.Equal(t, "", dir)

	origDir := filepath.Join("somedir", "somesub")
	dir, err = fsutil.Expand(origDir)
	require.NoError(t, err)
	require.Equal(t, origDir, dir)

	_, err = fsutil.Expand(filepath.FromSlash("~nosuchuser/somedir"))
	require.Error(t, err)

	const homeEnv = "HOME"
	origHome := os.Getenv(homeEnv)
	defer func() {
		os.Setenv(homeEnv, origHome)
	}()
	homeDir := filepath.FromSlash("/tmp/testhome")
	os.Setenv(homeEnv, homeDir)

	const subDir = "mytmp"
	origDir = filepath.Join("~", subDir)
	dir, err = fsutil.Expand(origDir)
	require.NoError(t, err)
	require.Equal(t, filepath.Join(homeDir, subDir), dir)

	os.Unsetenv(homeEnv)
	_, err = fsutil.Expand(origDir)
	require.Error(t, err)
}
