package fileutils

/*
// Flags to OpenFile wrapping those of the underlying system.
// Not all flags may be implemented on a given system.
const (
    // Specify exactly one of O_RDONLY, O_WRONLY, or O_RDWR
    O_RDONLY int = syscall.O_RDONLY // open the file R/O
    O_WRONLY int = syscall.O_WRONLY // open the file write-only
    O_RDWR   int = syscall.O_RDWR   // open the file R/W
    // The remaining values may be or'ed in to control behavior.
    O_APPEND int = syscall.O_APPEND // append data to the file when writing.
    O_CREATE int = syscall.O_CREAT  // create a new file if none exists.
    O_EXCL   int = syscall.O_EXCL   // used with O_CREATE, file must not exist.
    O_SYNC   int = syscall.O_SYNC   // open for synchronous I/O.
    O_TRUNC  int = syscall.O_TRUNC  // truncate regular writable file after open.
)
*/

/*
when file already exists, either truncate it or fail:

openOpts := os.O_RDWR|os.O_CREATE
if OKtoTruncateWhenExists {
    openOpts |= os.O_TRUNC // file will be truncated
} else {
    openOpts |= os.O_EXCL  // file must not exist
}
f, err := os.OpenFile(filePath, openOpts, 0644)
// ... do stuff
*/

/*
https://pkg.go.dev/os
var (
	// ErrInvalid indicates an invalid argument.
	// Methods on File will return this error when the receiver is nil.
	ErrInvalid = fs.ErrInvalid // "invalid argument"

	ErrPermission = fs.ErrPermission // "permission denied"
	ErrExist      = fs.ErrExist      // "file already exists"
	ErrNotExist   = fs.ErrNotExist   // "file does not exist"
	ErrClosed     = fs.ErrClosed     // "file already closed"

	ErrNoDeadline = errNoDeadline()  // "file type does not support deadline"

	ErrDeadlineExceeded = errDeadlineExceeded() // "i/o timeout"
)

https://stackoverflow.com/questions/12518876/how-to-check-if-a-file-exists-in-go
Instead of using os.Create, you should use
os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666) .
That way you'll get an error if the file already exists.
Also this doesn't have a race condition with something
else making the file, unlike a version that checks for
existence beforehand.
https://groups.google.com/g/golang-nuts/c/Ayx-BMNdMFo/m/4rL8FFHr8v4J
Often an os.Exists function is not really needed.
For instance: if you are going to open the file, why check
whether it exists first? Simply call os.IsNotExist(err) after
you try to open the file, and deal with non-existence there.

os.IsExist(err) is good for cases when you expect the
file to not exist yet, but the file actually exists:
os.Symlink, os.Mkdir,
os.OpenFile(target, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
(os.IsExist(err) will trigger when target exists because
O_EXCL means that file should not exist yet)

os.IsNotExist(err) is good for more common cases where you
expect the file to exists, but it actually doesn't exist:
os.Chdir, os.Stat, os.Open, os.OpenFile(without os.O_EXCL),
os.Chmod, os.Chown, os.Close, os.Read, os.ReadAt, os.ReadDir,
os.Readdirnames, os.Seek, os.Truncate
*/

import (
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3" // to get init()
	"io/fs"
	"os"
	FP "path/filepath"
	S "strings"
)

// ResolvePath is needed because functions in package
// path/filepath do not handle "~" (home directory)
// well. If an error occurs (for whatever reason),
// we punt: simply return the original input argument.
//
func ResolvePath(s string) string {
	// ABSOLUTE FILEPATH
	if FP.IsAbs(s) {
		return s
	}
	// RELATIVE FILEPATH (except ~)
	if !S.HasPrefix(s, "~") {
		s2, e := FP.Abs(s)
		if e == nil {
			return s2
		}
		// NOTE: Should this be a panic ?
		fmt.Fprintf(os.Stderr, "fp.Abs(%s) failed: %s \n", s, e)
		return s
	}
	// HOME DIR FILEPATH
	homedir, e := os.UserHomeDir()
	if e != nil {
		return s
	}
	if s == "~" {
		return homedir
	}
	if !S.HasPrefix(s, "~/") {
		fmt.Fprintf(os.Stderr,
			"not allowed to access other user's homedir:", s)
		return s
	}
	return FP.Join(homedir, s[2:])
}

// FileAtPath checks that the file exists AND that it is "regular"
// (not dir, symlnk, pipe), and also returns permissions.
//
// Return values:
//  - (true, permissions:0nnn, nil) if a regular file exists
//  - (false, fs.FileMode, nil) if something else exists (incl. dir)
//  - (false, 0, nil) if nothing at all exists
//  - (false, 0, anError) if some unusual error was returned (failing disk?)
// Notes & caveats:
//  - File emptiness (i.e. length 0) is not checked
//  - "~" for user home dir is not expanded and will fail
//
func FileAtPath(aPath string) (bool, fs.FileMode, error) {
	aPath = ResolvePath(aPath)
	// If nothing exists at the filepath,
	// Stat returns os.ErrNotExist
	fi, e := os.Stat(aPath)
	// fi.IsRegular() excludes directories, pipes, symlinks,
	// append-only, exclusive-access, and other detritus.
	if e == nil {
		// Something exists !
		// (fi should never be nil)
		if fi.Mode().IsRegular() {
			return true, fi.Mode().Perm(), nil
		}
		// Non-regular, incl. dirs, symlinks, pipes.
		// The exact type is indicated in the FileMode.
		return false, fi.Mode(), nil
	}
	if errors.Is(e, os.ErrNotExist) {
		// Nothing exists !
		return false, 0, nil
	}
	// Something WEIRD happened.
	// Maybe this should panic instead.
	return false, 0, e
}
