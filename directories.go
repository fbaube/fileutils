package fileutils

import (
	"fmt"
	"os"
	"io/ioutil"
	"path"
	FP "path/filepath"
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

func MakeDirectoryExist(path string) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			if err = os.Mkdir(path, os.ModePerm); err != nil {
				return fmt.Errorf("Can't create directory <%s>: %w", path, err)
			}
		} else {
			return fmt.Errorf("Can't access directory <%s>: %w", path, err)
		}
	}
	return nil
}

func ClearOutDirectory(path string) error {
  dir, err := os.Open(path)
  if err != nil {
    return fmt.Errorf("Can't access directory <%s>: %w", path, err)
	}
  defer dir.Close()
	names, err := dir.Readdirnames(-1)
  if err != nil {
    return fmt.Errorf("Can't read directory <%s>: %w", path, err)
	}
	for _, name := range names {
		if err = os.RemoveAll(FP.Join(path, name)); err != nil {
			return fmt.Errorf("error clearing file %s: %v", name, err)
		}
	}
	return nil
}

// CopyDirRecursivelyFromTo copies a whole directory recursively.
// Both argument should be directories !!
func CopyDirRecursivelyFromTo(src string, dst string) error {
	var err error
	var fds []os.FileInfo
	var srcinfo os.FileInfo

	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}

	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}

	if fds, err = ioutil.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = CopyDirRecursivelyFromTo(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		} else {
			if err = CopyFileFromTo(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}
