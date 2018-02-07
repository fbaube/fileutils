package fileutils

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
)

// IsDirectory returns true IFF the directory exists and is in fact a directory.
func IsDirectory(path string) bool {
	fi, err := os.Lstat(path)
	return (err == nil && fi.IsDir())
}

// MustOpenExistingDir returns the directory IFF it exists and can be opened.
func MustOpenExistingDir(path string) (*os.File, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, errors.Wrap(e,
			fmt.Sprintf("fileutils.MustOpenExistingDir.Open<%s>", path))
	}
	fi, e := os.Lstat(path)
	if e != nil || !fi.IsDir() {
		return nil, fmt.Errorf("%s: not a directory", path)
	}
	return f, nil
}

// MustOpenOrCreateDir returns true if (a) the directory exists and can
// be opened, or (b) it does not exist, and/but it can be created anew.
func MustOpenOrCreateDir(path string) (*os.File, error) {
	// Does it already exist ?
	f, e := MustOpenExistingDir(path)
	if e == nil {
		return f, nil
	}
	// Try to create it
	e = os.Mkdir(path, 0777)
	if e == nil {
		// Now we *have* to open it
		f, e = MustOpenExistingDir(path)
		if e != nil {
			return nil, fmt.Errorf("fileutils.MustOpenDir<%s>", path)
		}
		// Succeeded
		return f, nil
	}
	// err != nil
	return nil, e
}

// DirectoryContents returns the results of (os.*File)Readdir(..):
//
// Readdir reads the contents of the directory associated with file
// and returns a slice of FileInfo values, as would be returned by
// Lstat(..), in directory order.
func DirectoryContents(f *os.File) ([]os.FileInfo, error) {
	f, e := MustOpenExistingDir(f.Name())
	if e != nil {
		return nil, errors.Wrap(e,
			fmt.Sprintf("fileutils.DirectoryContents.MustOpenExistingDir<%s>", f.Name()))
	}
	defer f.Close()
	// 0 means No limit, read'em all
	fis, e := f.Readdir(0)
	if e != nil {
		return nil, errors.Wrap(e,
			fmt.Sprintf("fileutils.DirectoryContents.Readdir<%s>", f.Name()))
	}
	return fis, nil
}

// DirectoryFiles is like DirectoryContents except that results that are
// directories (not files) are nil'ed out. if there were entries but none
// were files, the first return value is set to zero and the slice to nil.
func DirectoryFiles(f *os.File) (int, []os.FileInfo, error) {
	fis, e := DirectoryContents(f)
	if e != nil {
		return 0, nil, errors.Wrap(e, "futil.GetDirFiles")
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
