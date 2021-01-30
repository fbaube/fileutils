// Original code:
// https://github.com/nnev/frank/writeatomically.go
// Licensed under the ISC license.

package fileutils

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
)

// Tempdir checks and returns the value of the envar `TMPDIR`.
func TempDir(dest string) string {
	tempdir := os.Getenv("TMPDIR")
	if tempdir == "" {
		// Convenient for development: decreases the chance that we
		// cannot move files due to TMPDIR being on a different file
		// system than dest.
		tempdir = filepath.Dir(dest)
	}
	return tempdir
}

// WriteAtomic is TBS.
func WriteAtomic(dest string, write func(w io.Writer) error) (err error) {
	f, err := os.CreateTemp(TempDir(dest), "atomic-")
	if err != nil {
		return err
	}
	defer func() {
		// Clean up (best effort) in case we are returning with an error:
		if err != nil {
			// Prevent file descriptor leaks.
			f.Close()
			// Remove the tempfile to avoid filling up the file system.
			os.Remove(f.Name())
		}
	}()

	// Use a buffered writer to minimize write(2) syscalls.
	bufw := bufio.NewWriter(f)
	w := io.Writer(bufw)

	if err := write(w); err != nil {
		return err
	}
	if err := bufw.Flush(); err != nil {
		return err
	}
	// Chmod the file world-readable (TempFile
	// creates files with mode 0600) before renaming.
	if err := f.Chmod(0644); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), dest)
}

/*
func main() {
    if err := writeAtomically("demo.txt", func(w io.Writer) error {
        _, err := w.Write([]byte("demo"))
        return err
    }); err != nil {
        log.Fatal(err)
    }
}
*/
