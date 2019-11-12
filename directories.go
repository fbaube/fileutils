package fileutils

import (
	"fmt"
	"os"
)

// DirExists returns true *iff* the directory
// exists and is in fact a directory.
func DirExists(path AbsFilePath) bool {
	fi, err := os.Lstat(path.S())
	return (err == nil && fi.IsDir())
}

// FileSize returns the size *iff* the
// filepath exists and is in fact a file.
func FileSize(path AbsFilePath) int {
	fi, err := os.Lstat(path.S())
	if err == nil && !fi.IsDir() {
		return int(fi.Size())
	}
	return 0
}

// OpenExistingDir returns the directory *iff* it exists and can be opened
// for reading. Note that the `os.File` can be nil without error. Thus we
// cannot (or: *do not*) distinguish btwn non-existence and an actual error.
// OTOH if it exists but is not a directory, return an error.
func OpenExistingDir(path AbsFilePath) (f *os.File, e error) {
	// "Open(s) opens the file for reading. If successful, methods on the returned
	// file can be used for reading; the associated FD has mode O_RDONLY."
	f, e = os.Open(path.S())
	if e != nil {
		return nil, nil // fmt.Errorf("fu.OpenExistingDir.Open<%s>: %w", path, e)
	}
	if f == nil {
		panic("fu.OpenExistingDir.Open: " + path + ": no error but nil file ?!")
	}
	fi, e := os.Lstat(path.S())
	if e != nil || !fi.IsDir() {
		return nil, fmt.Errorf("fu.mustOpenExistingDir.notaDir<%s>: %w", path, e)
	}
	return f, nil
}

// OpenOrCreateDir returns true if (a) the directory exists and can be
// opened, or (b) it does not exist, and/but it can be created anew.
func OpenOrCreateDir(path AbsFilePath) (f *os.File, e error) {
	// Does it already exist ?
	f, e = OpenExistingDir(path)
	if e == nil {
		return f, nil
	}
	// If error, maybe it just dusnt exist, so try to create it
	e = os.Mkdir(path.S(), 0777)
	// If error, give up.
	if e != nil {
		return nil, fmt.Errorf("fu.OpenOrCreateDir<%s>: can't do either: %w", path, e)
	}
	// Now we *have* to open it
	return Must(OpenExistingDir(path)), nil
}

// DirectoryContents returns the results of `(*os.File)Readdir(..)`.
// `File.Name()` might be a relative filepath but if it was opened
// okay then it at least functions as an absolute filepath.
// If the path is not a directory then it panics. <br/>
// `Readdir(..)` reads the contents of the directory associated
// with the `File` argument and returns a slice of `FileInfo`
// values, as would be returned by `Lstat(..)`, in directory order.
func DirectoryContents(f *os.File) ([]os.FileInfo, error) {
	f = Must(OpenExistingDir(AbsFilePath(f.Name())))
	defer f.Close()
	// 0 means No limit, read'em all
	fis, e := f.Readdir(0)
	if e != nil {
		return nil, fmt.Errorf("fu.DirectoryContents.Readdir<%s>: %w", f.Name(), e)
	}
	return fis, nil
}

// DirectoryFiles is like `DirectoryContents(..)` except that
// results that are directories (not files) are nil'ed out. If
// there were entries but none were files, it return (`0,nil,nil`).
func DirectoryFiles(f *os.File) (int, []os.FileInfo, error) {
	fis, e := DirectoryContents(f)
	if e != nil {
		return 0, nil, fmt.Errorf("fu.DirectoryFiles<%s>: %w", f.Name(), e)
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
