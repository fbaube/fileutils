package fileutils

import (
	"bufio"
	"bytes"
	"os"

	"github.com/pkg/errors"
)

func (in AbsFilePath) Enslice() []AbsFilePath {
	out := make([]AbsFilePath, 0)
	out = append(out, in)
	return out
}

// MustOpenRW opens (and returns) the filepath as a writable file.
func MustOpenRW(path AbsFilePath) (*os.File, error) {
	f, e := os.OpenFile(string(path), os.O_RDWR, 0666)
	if e != nil {
		return nil, errors.Wrapf(e, "fu.MustOpenRW.OpenFile<%s>", path)
	}
	fi, e := os.Lstat(string(path))
	if e != nil || !fi.Mode().IsRegular() {
		return nil, errors.Wrapf(e, "fu.MustOpenRW.notaFile<%s>", path)
	}
	return f, nil
}

// MustOpenRO opens (ansd returns) the filepath as a readable file.
func TryOpenRO(path AbsFilePath) (*os.File, error) {
	f, e := os.Open(string(path))
	if e != nil {
		return nil, errors.Wrapf(e, "fu.TryOpenRO.OpenFile<%s>", path)
	}
	fi, e := os.Lstat(string(path))
	if e != nil || fi.IsDir() {
		return nil, errors.Wrapf(e, "fu.TryOpenRO.notaFile<%s>", path)
	}
	return f, nil
}

// MustCreateEmpty opens the filepath as a writable empty file.
func MustCreateEmpty(path AbsFilePath) (*os.File, error) {
	// Create creates the named file with mode 0666 (before umask),
	// truncating it if it already exists. If successful, methods
	// on the returned File can be used for I/O; the associated
	// file descriptor has mode O_RDWR. If there is an error,
	// it will be of type *PathError.
	f, e := os.Create(string(path))
	if e != nil {
		return nil, errors.Wrapf(e, "fu.MustCreateEmpty.Create<%s>", path)
	}
	fi, e := os.Stat(string(path))
	if e != nil || !fi.Mode().IsRegular() {
		return nil, errors.Wrapf(e, "fu.MustCreateEmpty.notaFile<%s>", path)
	}
	return f, nil
}

// SameContents returns: Are the two files' contents identical ?
func SameContents(f1, f2 *os.File) bool {
	s1 := bufio.NewScanner(f1)
	s2 := bufio.NewScanner(f2)
	for s1.Scan() {
		s2.Scan()
		if !bytes.Equal(s1.Bytes(), s2.Bytes()) {
			return false
		}
	}
	return true
}
