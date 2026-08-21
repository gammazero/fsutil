# fsutil

[![GoDoc](https://pkg.go.dev/badge/github.com/gammazero/fsutil)](https://pkg.go.dev/github.com/gammazero/fsutil)
[![Build Status](https://github.com/gammazero/fsutil/actions/workflows/go.yml/badge.svg)](https://github.com/gammazero/fsutil/actions/workflows/go.yml)
[![codecov](https://codecov.io/gh/gammazero/fsutil/graph/badge.svg?token=U2Y5KBC0H3)](https://codecov.io/gh/gammazero/fsutil)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Filesystem utility for common file and directory checks.

## Packages

### [fsutil](https://pkg.go.dev/github.com/gammazero/fsutil)

Common file and directory checks: `DirEmpty`, `DirExists`, `DirWritable`, `ExpandHome`, `FileChanged`, `FileExists`, `IsSubpath`.

Notes:
- `DirWritable` creates the directory if it does not exist.
- `IsSubpath` is purely lexical; it does not follow symlinks or consult the filesystem.
- `ExpandHome` expands only a bare `~` prefix; `~user` forms return an error.
- `FileExists` returns true for any stat error other than "not exist" (e.g., permission denied), so true means "not known to be absent".

### [fsutil/atomicfile](https://pkg.go.dev/github.com/gammazero/fsutil/atomicfile)

Creates a temporary file that is renamed to the specified path when `Close` is called. This prevents a partially written file from being visible when writes are in progress or when a failure occurs during writing. The temporary file is created in the same directory as the target path so that the rename stays on one filesystem. `Discard` removes the temporary file without renaming it. `Close` does not call `Sync`; callers wanting durability use `CloseSync`, which syncs the file, renames it, and syncs the containing directory so that the rename itself survives a crash. `WriteFile` atomically and durably writes a byte slice to a file in a single call. `CopyFile` copies a regular file, preserving its permission mode; the destination is never observed partially written.

### [fsutil/disk](https://pkg.go.dev/github.com/gammazero/fsutil/disk)

Reports disk usage on multiple platforms: aix, darwin (macOS), freebsd, linux, openbsd, windows. In the returned `UsageStats`, `Total` and `Used` describe the whole disk, while `Free` and `Percent` reflect the user's view (root typically has ~5% reserved space that users cannot use).
