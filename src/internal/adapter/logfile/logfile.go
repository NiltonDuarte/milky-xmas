// Package logfile sends the standard logger's output to a file as well as to standard error.
package logfile

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

// EnvVar names the environment variable that points at the log file.
const EnvVar = "MILKY_LOG_FILE"

// maxSize is the size at which the previous log is kept as path+".1" and a new file starts.
const maxSize = 10 << 20

// Setup appends the standard logger's output to the file at path, creating it and its folder with
// owner-only permissions. A file above maxSize is moved to path+".1" first, replacing the older
// copy. The returned function closes the file. An empty path leaves the logger unchanged.
func Setup(path string) (func() error, error) {
	if path == "" {
		return func() error { return nil }, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("create log folder: %w", err)
	}
	if info, err := os.Stat(path); err == nil && info.Size() > maxSize {
		if err := os.Rename(path, path+".1"); err != nil {
			return nil, fmt.Errorf("rotate log file: %w", err)
		}
	}
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}
	log.SetOutput(io.MultiWriter(os.Stderr, file))
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	return file.Close, nil
}
