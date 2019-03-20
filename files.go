package fileutils

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	fp "path/filepath"
	S "strings"

	"github.com/mgutz/str"
	"github.com/pkg/errors"
)

// AbsWRT is like filepath.Abs(..): it can convert a possibly-relative
// filepath to an absolute filepath. The difference is that a relative
// filepath argument is not resolved w.r.t. the current working directory;
// it is instead done w.r.t. the supplied directory argument.
func AbsWRT(maybeRelFP string, wrtDir string) string {
	if fp.IsAbs(maybeRelFP) {
		return maybeRelFP
	}
	return fp.Join(wrtDir, maybeRelFP)
}

/*
// MustOpenRW opens (and returns) the filepath as a writable file.
func MustOpenRW(path string) (*os.File, error) {
	f, e := os.OpenFile(path, os.O_RDWR, 0666)
	if e != nil {
		return nil, errors.Wrapf(e, "fu.MustOpenRW.OpenFile<%s>", path)
	}
	fi, e := os.Lstat(path)
	if e != nil || !fi.Mode().IsRegular() {
		return nil, errors.Wrapf(e, "fu.MustOpenRW.notaFile<%s>", path)
	}
	return f, nil
}
*/

// OpenRO opens (and returns) the filepath as a readable file.
func OpenRO(path AbsFilePath) (*os.File, error) {
	f, e := os.Open(string(path))
	if e != nil {
		return nil, errors.Wrapf(e, "fu.TryOpenRO.os.Open<%s>", path)
	}
	fi, e := os.Lstat(string(path))
	if e != nil || fi.IsDir() {
		return nil, errors.Wrapf(e, "fu.TryOpenRO.notaFile<%s>", path)
	}
	return f, nil
}

// CreateEmpty opens the filepath as a writable empty file,
// truncating it if it exists and is non-empty.
func CreateEmpty(path AbsFilePath) (*os.File, error) {
	// Create creates the named file with mode 0666 (before umask),
	// truncating it if it already exists. If successful, methods
	// on the returned File can be used for I/O; the associated
	// file descriptor has mode O_RDWR. If there is an error,
	// it will be of type *PathError.
	f, e := os.Create(string(path))
	if e != nil {
		return nil, errors.Wrapf(e, "fu.CreateEmpty.Create<%s>", path)
	}
	fi, e := os.Stat(string(path))
	if e != nil || !fi.Mode().IsRegular() {
		return nil, errors.Wrapf(e, "fu.CreateEmpty.notaFile<%s>", path)
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

// Exists returns true *iff* the file
// exists and is in fact a file.
func Exists(path string) bool {
	fi, err := os.Lstat(path)
	return (err == nil && fi.Mode().IsRegular())
}

// Exists returns true *iff* the file
// exists and is in fact a file.
func (afp AbsFilePath) Exists() bool {
	fi, err := os.Lstat(afp.S())
	return (err == nil && fi.Mode().IsRegular())
}

// IsNonEmpty returns true *iff* the file exists
// *and* contains at least one byte of data.
func IsNonEmpty(path string) bool {
	fi, err := os.Lstat(path)
	return (err == nil && fi.Mode().IsRegular() && fi.Size() > 0)
}

// IsXML returns true *iff* the file exists *and*
// appears to be XML. The check is simple though.
func IsXML(path string) bool {
	if !IsNonEmpty(path) {
		return false
	}
	file, e := os.Open(path)
	if e != nil {
		panic("fu.IsXML.os.Open<" + path + ">")
	}
	var bb []byte
	bb = make([]byte, 256)
	nRedd, e := file.Read(bb)
	if e != nil {
		panic("fu.IsXML")
	}
	// the minimum valid XML file is " <x/> "
	if nRedd < 4 {
		return false
	}
	var s string
	s = S.TrimSpace(string(bb))
	if !S.HasPrefix(s, "<") {
		return false
	}
	OKprefixes := []string{"<?", "<!", "<--"} // and tags!
	for _, ss := range OKprefixes {
		if S.HasPrefix(s, ss) {
			return true
		}
	}
	// We require valid XML tags to begin with A-Za-z
	return str.IsAlpha(str.CharAt(s, 1))
}

func CopyFileFromTo(src, dst string) error {
	if dst == "" {
		return nil
	}
	if src == "" {
		return fmt.Errorf("No source for copy to: " + src)
	}
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("Not a regular file: " + src)
	}
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destin, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destin.Close()
	nBytes, err := io.Copy(destin, source)
	if err != nil {
		return err
	}
	if nBytes != sourceFileStat.Size() {
		return fmt.Errorf("File copy error: had %d bytes, copied %d bytes",
			sourceFileStat.Size(), nBytes)
	}
	return nil
}
