package fileutils

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
)

// DirExists returns true *iff* the directory
// exists and is in fact a directory.
func DirExists(path AbsFilePath) bool {
	fi, err := os.Lstat(string(path))
	return (err == nil && fi.IsDir())
}

// FileSize returns the size *iff* the
// filepath exists and is in fact a file.
func FileSize(path AbsFilePath) int {
	fi, err := os.Lstat(string(path))
	if err == nil && !fi.IsDir() {
		return int(fi.Size())
	}
	return 0
}

// OpenExistingDir returns the directory
// *iff* it exists and can be opened.
func OpenExistingDir(path AbsFilePath) (*os.File, error) {
	f, e := os.Open(string(path))
	if e != nil {
		return nil, errors.Wrapf(e, "fu.OpenExistingDir.Open<%s>", path)
	}
	fi, e := os.Lstat(string(path))
	if e != nil || !fi.IsDir() {
		return nil, errors.New(fmt.Sprintf("fu.MustOpenExistingDir.notaDir<%s>", path))
	}
	return f, nil
}

// OpenOrCreateDir returns true if (a) the directory exists and can be
// opened, or (b) it does not exist, and/but it can be created anew.
func OpenOrCreateDir(path AbsFilePath) (f *os.File, e error) {

	// Does it already exist ?
	f = Must(OpenExistingDir(path)) /*
		if e == nil {
			return f, nil
		} */
	// Try to create it
	e = os.Mkdir(string(path), 0777)
	if e == nil {
		// Now we *have* to open it
		f = Must(OpenExistingDir(path)) /*
			if e != nil {
				return nil, errors.New(fmt.Sprintf("fu.MustOpenDir<%s>", path))
			} */
		// Succeeded
		return f, nil
	}
	// err != nil
	return nil, e
}

// DirectoryContents returns the results of `(*os.File)Readdir(..)`.
// `File.Name()` might be a relative filepath but if it was opened
// okay then it at least functions as an absolute filepath.
// `Readdir(..)` reads the contents of the directory associated
// with the `File` argument and returns a slice of `FileInfo`
// values, as would be returned by `Lstat(..)`, in directory order.
func DirectoryContents(f *os.File) ([]os.FileInfo, error) {
	f = Must(OpenExistingDir(AbsFilePath(f.Name()))) /*
		if e != nil {
			return nil, errors.Wrapf(e,
				"fu.DirectoryContents.MustOpenExistingDir<%s>", f.Name())
		} */
	defer f.Close()
	// 0 means No limit, read'em all
	fis, e := f.Readdir(0)
	if e != nil {
		return nil, errors.Wrapf(e, "fu.DirectoryContents.Readdir<%s>", f.Name())
	}
	return fis, nil
}

// DirectoryFiles is like `DirectoryContents(..)` except that results
// that are directories (not files) are nil'ed out. If there were
// entries but none were files, it return (`0,nil`).
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
