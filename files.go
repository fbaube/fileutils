package fileutils

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	FP "path/filepath"
	S "strings"

	"github.com/mgutz/str"
)

// AbsWRT is like "filepath.Abs(..)"": it can convert a possibly-relative
// filepath to an absolute filepath. The difference is that a relative
// filepath argument is not resolved w.r.t. the current working directory;
// it is instead done w.r.t. the supplied directory argument.
func AbsWRT(problyRelFP string, wrtDir string) string {
	if FP.IsAbs(problyRelFP) {
		return problyRelFP
	}
	return FP.Join(wrtDir, problyRelFP)
}

// OpenRW opens (and returns) the filepath as a writable file.
// An existing file is not truncated, merely opened. 
func OpenRW(path string) (f *os.File, e error) {
	f, e = os.OpenFile(path, os.O_RDWR, 0666)
	if e != nil {
		return nil, fmt.Errorf("fu.OpenRW.OpenFile<%s>: %w", path, e)
	}
	fi, e := os.Lstat(path)
	if e != nil || !fi.Mode().IsRegular() {
		return nil, fmt.Errorf("fu.OpenRW.notaFile<%s>: %w", path, e)
	}
	return f, nil
}

// OpenRO opens (and returns) the filepath as a readable file.
func OpenRO(path string) (f *os.File, e error) {
	f, e = os.Open(path)
	if e != nil {
		return nil, fmt.Errorf("fu.OpenRO.os.Open<%s>: %w", path, e)
	}
	fi, e := os.Lstat(path)
	if e != nil || fi.IsDir() {
		return nil, fmt.Errorf("fu.OpenRO.notaFile<%s>: %w", path, e)
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
	spath := path.S()
	f, e := os.Create(spath)
	if e != nil {
		return nil, fmt.Errorf("fu.CreateEmpty.Create<%s>: %w", spath, e)
	}
	fi, e := os.Stat(spath) // Lstat is not needed 
	if e != nil || !fi.Mode().IsRegular() {
		return nil, fmt.Errorf("fu.CreateEmpty.notaFile<%s>: %w", spath, e)
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
	s := S.TrimSpace(string(bb))
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

// CopyFileGreedily reads the entire file into memory,
// and is therefore memory-constrained !
func CopyFileGreedily(src string, dst string) error {
	var e error
	var data []byte
	// Read all content of src to data
	if data, e = os.ReadFile(src); e != nil {
		return e
	}
	// Write data to dst
	if e = os.WriteFile(dst, data, 0644); e != nil {
		return e
	}
	return nil
}

// CopyFileFromTo copies a single file from src to dst.
func CopyFileFromTo(src, dst string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if srcfd, err = os.Open(src); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.Create(dst); err != nil {
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	// Lstat might helpful here ?? 
	if srcinfo, err = os.Stat(src); err != nil { 
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}

func AppendToFileBaseName(name, toAppend string) string {
	if name == "" || toAppend == "" {
		return ""
	}
	ext := FP.Ext(name)
	return S.TrimSuffix(name, ext) + toAppend + ext
}
