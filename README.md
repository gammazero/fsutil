# fsutil

[![GoDoc](https://pkg.go.dev/badge/github.com/gammazero/fsutil)](https://pkg.go.dev/github.com/gammazero/fsutil)
[![Build Status](https://github.com/gammazero/fsutil/actions/workflows/go.yml/badge.svg)](https://github.com/gammazero/fsutil/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/gammazero/fsutil)](https://goreportcard.com/report/github.com/gammazero/fsutil)
[![codecov](https://codecov.io/gh/gammazero/fsutil/graph/badge.svg?token=U2Y5KBC0H3)](https://codecov.io/gh/gammazero/fsutil)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Filesystem utility for common file and directory checks.

## Packages

### [fsutil](https://pkg.go.dev/github.com/gammazero/fsutil)

Common file and directory checks: `DirEmpty`, `DirExists`, `DirWritable`, `ExpandHome`, `FileChanged`, `FileExists`.

### [fsutil/atomicfile](https://pkg.go.dev/github.com/gammazero/fsutil/atomicfile)

Creates a temporary file that is renamed to the specified path when `Close` is called. This prevents a partially written file from being visible when writes are in progress or when a failure occurs during writing.

### [fsutil/disk](https://pkg.go.dev/github.com/gammazero/fsutil/disk)

Reports disk usage on multiple platforms: aix, darwin (macOS), freebsd, linux, openbsd, windows.
