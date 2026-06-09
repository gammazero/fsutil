// Package atomicfile creates a temporary file that is renamed to the specified
// path when Close is called. The prevents creating a partially written file
// when there are writes in progress or when there is a failure while writing
// to the file.
package atomicfile
