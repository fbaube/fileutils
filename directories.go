package fileutils

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
)

// IsDirectory returns true IFF the directory exists and is in fact a directory.
func IsDirectory(path AbsFilePath) bool {
	fi, err := os.Lstat(string(path))
	return (err == nil && fi.IsDir())
}

// MustOpenExistingDir returns the directory IFF it exists and can be opened.
func MustOpenExistingDir(path AbsFilePath) (*os.File, error) {
	f, e := os.Open(string(path))
	if e != nil {
		return nil, errors.Wrapf(e, "fu.MustOpenExistingDir.Open<%s>", path)
	}
	fi, e := os.Lstat(string(path))
	if e != nil || !fi.IsDir() {
		return nil, errors.New(fmt.Sprintf("fu.MustOpenExistingDir.notaDir<%s>", path))
	}
	return f, nil
}

// MustOpenOrCreateDir returns true if (a) the directory exists and can
// be opened, or (b) it does not exist, and/but it can be created anew.
func MustOpenOrCreateDir(path AbsFilePath) (*os.File, error) {
	// Does it already exist ?
	f, e := MustOpenExistingDir(path)
	if e == nil {
		return f, nil
	}
	// Try to create it
	e = os.Mkdir(string(path), 0777)
	if e == nil {
		// Now we *have* to open it
		f, e = MustOpenExistingDir(path)
		if e != nil {
			return nil, errors.New(fmt.Sprintf("fu.MustOpenDir<%s>", path))
		}
		// Succeeded
		return f, nil
	}
	// err != nil
	return nil, e
}

// DirectoryContents returns the results of (os.*File)Readdir(..).
// File.Name() might be a relative filepath but if it was Open()ed
// okay then it at least functions as an absolute filepath.
// Readdir reads the contents of the directory associated with file
// and returns a slice of FileInfo values, as would be returned by
// Lstat(..), in directory order.
func DirectoryContents(f *os.File) ([]os.FileInfo, error) {
	f, e := MustOpenExistingDir(AbsFilePath(f.Name()))
	if e != nil {
		return nil, errors.Wrapf(e,
			"fu.DirectoryContents.MustOpenExistingDir<%s>", f.Name())
	}
	defer f.Close()
	// 0 means No limit, read'em all
	fis, e := f.Readdir(0)
	if e != nil {
		return nil, errors.Wrapf(e, "fu.DirectoryContents.Readdir<%s>", f.Name())
	}
	return fis, nil
}

// DirectoryFiles is like DirectoryContents except that results that are
// directories (not files) are nil'ed out. if there were entries but none
// were files, the first return value is set to zero and the slice to nil.
func DirectoryFiles(f *os.File) (int, []os.FileInfo, error) {
	fis, e := DirectoryContents(f)
	if e != nil {
		return 0, nil, errors.Wrapf(e, "fu.DirectoryFiles<%s>", f.Name())
	}
	nFiles := 0
	for i := range fis {
		if fis[i].IsDir() {
			fis[i] = nil
		} else {
			nFiles++
		}
	}
	if nFiles == 0 {
		fis = nil
	}
	return nFiles, fis, nil
}
