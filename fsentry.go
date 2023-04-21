package fileutils

import (
	"io/fs"
)

// FSEntry handles read-only access
// and attributes for files, dirs,
// and symilnks; it extends [fs.File]
// (also R/O), which comprises
//   - Stat() (FileInfo, error)
//   - Read([]byte) (int, error)
//   - Close() error
//
// Related there is also interface
// [io/fs.ReadFileFS):
//   - FS
//   - For ReadFile(s),
//   - Success returns a nil error, not io.EOF.
//   - The caller may modify the returned byte slice.
//   - This method should return a copy of the underlying data.
//     ReadFile(name string) ([]byte, error)
//
// Related there is also interface
// [io/fs.ReadDirFS]:
//   - FS
//   - ReadDir(name string) ([]DirEntry, error)
//
// Related there is also interface
// [io/fs.ReadDirFile]:
//   - File
//   - ReadDir reads the contents of the directory and returns
//     a slice of up to n DirEntry values in directory order.
//     Subsequent calls on the same file will yield further
//     DirEntry values.
//     If n > 0, ReadDir returns at most n DirEntry structures.
//     In this case, if ReadDir returns an empty slice, it will
//     return a non-nil error explaining why.
//     At the end of a directory, the error is io.EOF.
//     (ReadDir must return io.EOF itself, not an error wrapping io.EOF.)
//     If n <= 0, ReadDir returns all the DirEntry values from
//     the directory in a single slice. In this case, if ReadDir
//     succeeds (reads all the way to the end of the directory),
//     it returns the slice and a nil error. If it encounters an
//     error before the end of the directory, ReadDir returns the
//     DirEntry list read until that point and a non-nil error.
//   - ReadDir(n int) ([]DirEntry, error)
//
// FSEntry adds convenience funcs for
// easy access to file attributes.
//
// It can co-exist with interface
// [orderednodes.Nord].
//
// It could also be rewritten to extend
// [os.file], which provides R/W access, but
// that is not the model of package [io/fs].
// .
type FSEntry interface {
	// ON.Nord // this should remain decoupled !
	fs.File
	IsFile() bool
	IsDir() bool
	IsSymlink() bool
	// File: nr bytes; Dir: nr files.
	Size() int
	// These methods use [Lstat] rather than [Stat],
	// so symlinks must be resolved "manually".
	ResolveSymlink() (string, error)
	// Refresh reloads the [FileInfo]
	// and returns: Did it change ?
	Refresh() bool
}

/* LSTAT v STAT

https://stackoverflow.com/questions/52654988/using-os-lstat-return-value-in-go

Stat  returns info about the target file.
Lstat returns info about the symlink itself.
It is for when you don't want to automatically follow.
os.Stat  gives you no way to tell the path is a symlink.

https://golangprojectstructure.com/check-if-file-exists-in-go-code/

If you are looking up a symbolic link,
- os.Stat follows the link and provides info
  about the file at its target location.
- os.Lstat providess info about the symbolic
  link itself, without actually following it.

*/

/* TERMINOLOGY

https://superuser.com/questions/1467102/is-there-a-general-term-for-the-items-in-a-directory

The POSIX readdir documentation uses the word "entry".

Strictly, a file (in the POSIX sense) may have multiple
directory entries (hard linked files, or any directory),
or none at all (unlinked but still open files). The file
and the entry are separate objects.

Remember that in POSIX, "file" includes all types of inode
(including directory), not just regular files). Directory
entries (filenames) are references to files / inodes.
*/
