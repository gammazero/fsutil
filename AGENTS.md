# AGENTS.md

This file provides guidance to coding agents when working with code in this repository.

## Commands

```bash
# Build
go build ./...

# Vet
go vet ./...

# Test all packages
go test ./...

# Test with coverage (matches CI)
go test -v ./... -coverprofile=coverage.txt -covermode=atomic

# Test a single package
go test ./atomicfile/...
go test ./disk/...

# Test a single test function
go test -run TestCreate ./atomicfile/...
```

## Architecture

This is a small Go utility library (`github.com/gammazero/fsutil`) with three independent scopes:

**Root package (`fsutil`)** - common filesystem checks: `DirEmpty`, `DirExists`, `DirWritable`, `ExpandHome`, `FileChanged`, `FileExists`, `IsSubpath`. No sub-package dependencies; only stdlib. Behavioral notes that matter to callers:
- `DirWritable` has a side effect: if the directory does not exist it is created (mode 0775). If it exists, writability is verified by creating and removing a temp file.
- `FileExists` returns true for any stat error other than "not exist" (e.g., permission denied), so true means "not known to be absent".
- `IsSubpath` is purely lexical: it resolves relative paths to absolute but does not follow symlinks or consult the filesystem.
- `ExpandHome` only expands a bare `~` prefix; `~user` forms return an error.

**`atomicfile`** - wraps `os.File` to perform an atomic rename-on-close. `Create` makes a temp file in the same directory as the target path (so the rename stays on one filesystem); `Close` renames it to the final name; `Discard` removes the temp file without renaming. `Name()` returns the final path (which does not exist until Close succeeds); `TempName()` returns the temp path. `Close` does NOT call Sync - callers wanting durability use `CloseSync` instead, which syncs the file, closes and renames it, and then syncs the parent directory so the rename itself survives a crash (the directory sync is a no-op on Windows, where directories cannot be synced). `WriteFile(path, data, mode)` is the one-shot helper: create, write, and `CloseSync` in a single call. `CopyFile(src, dst)` atomically and durably copies a regular file (rejecting directories and other non-regular files), preserving the source's permission mode; a failed copy leaves an existing destination intact. On failed close or rename, the temp file is removed and errors are joined. The `atomicfile` tests import the root package (`fsutil.FileExists`).

**`disk`** - cross-platform disk usage via `Usage(path)`. The public API lives in `usage.go`; platform-specific implementations are in `usage_unix.go`, `usage_windows.go`, and `usage_openbsd.go`. Uses `golang.org/x/sys` for syscalls. In `UsageStats`, `Total` and `Used` describe the whole disk, while `Free` and `Percent` reflect the user's view (root typically has ~5% reserved space that users cannot use).

The only external dependency is `golang.org/x/sys`, used exclusively by the `disk` package. All tests are in external test packages (`package foo_test`).
