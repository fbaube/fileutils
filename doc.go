// Package fileutils does handy stuff at the level of files and directories.
// This includes basic file content snalysis an categorisation.
//
// AbsoluteFilePath/AbsFP versus RelativeFilePath/RelFP
//
// Many functions in this package are agnostic about getting an absolute
// or a relative path argument, so these functions take string arguments.
// However other functions are opinionated about only accepting absolute
// filepaths, so for convenience & correctness we define the type
// [AbsFilePath], a new type based on string.
//
// This can be handy in data structures, where a string field RelFilePath
// can store a path as it was supplied by the user (or used in a file
// cross-reference), while a field AbsFilePath can represent the path as
// fully resolved. Note that the runtime resolves relative paths relative
// to the current working directory, but at least one function here takes
// a "WRT" argument that can define a different reference point.
//
// The type MType
//
// Because this package supports the processing of mixed content in
// the three markup formats supported by LwDITA (Lightweight DITA),
// it introduces the idea of an MType, analogous to a MIME type.
// See file mtype.go .
// 
// Dependencies
//
// This package imports github.com/fbaube/(stringutils,wasmutils)
//
// Dependencies on packages os, path, filepath, fs 
//
// Note that for simplicity and correctness, this package should
// depend as much as possible on these stdlib libraries:
//
// 1) "os" https://golang.org/pkg/os/
//
// A mix of pure functions and os.File methods.
// See comments below in source file.
//
// 2) "path" https://golang.org/pkg/path/
//
// Funcs: Base(s) Clean(s) Dir(s) Ext(s) IsAbs(s) Join(s..) Split(s) Match(..)
//
// Package path has utility routines for manipulating slash-separated paths.
// Use only for paths separated by forward slashes, such as URL paths. This
// package does not deal with Windows paths with drive letters or backslashes;
// to do O/S paths, use package path/filepath .
//
// 3) "filepath" https://golang.org/pkg/path/filepath/
//
// Funcs: as for package "path" above but optimised for file paths,
// plus: Abs(s) EvalSymlinks(s) FromSlash(s) Glob(s) Rel(base,target)
// SplitList(s) ToSlash(s) VolumeName(s)
// Walk(root string, walkFn WalkFunc) type_WalkFunc
//
// 4) fs https://golang.org/pkg/io/fs/
//
// Miscellaneous notes about file and directory functions in package [os]:
// 
// - func Getwd() (dir string, err error)
// - func Mkdir[All](name string, perm FileMode) error
// - func Readlink(name string) (string, error) // returns the dest of the symlink
// - func Remove(name string) error // "rm" file or empty dir
// - func RemoveAll(path string) error // "rm" path and any children it has.
//   It removes everything it can but returns the first error it encounters.
//   If the path does not exist, RemoveAll returns nil (no error).
// - func Rename(oldpath, newpath string) error // "mv" oldpath to newpath.
//   If newpath already exists and is not a directory, Rename replaces it.
//   Maybe per-OS restrictions if oldpath & newpath are in different dirs.
// - func SameFile(fi1, fi2 FileInfo) bool
// - func Symlink(oldname, newname string) error // newname as symlink to oldname.
// - func Truncate(name string, size int64) error // changes the file's size.
//   If the file is a symlink, it changes the size of the link's target.
//
// os.File c-tors:
// - func Create(name string) (*File, error) is TruncreateRW().
//   If exists, truncate. If not, create mode 0666. If OK, is RW (O_RDWR).
// - func Open(name string) (*File, error) is OpenExistingRO().
//   Opens named file RO (O_RDONLY).
// - func OpenFile(name string, flag int, perm FileMode) (*File, error)
//   is generalized open call (Open and Create are usual). If exist, named
//   file is opened with specified flag (O_RDONLY etc.). If not exist, and
//   O_CREATE flag is passed, file is created with mode perm (before umask).
//
// os.File methods:
// - func (f *File) Readdir(n int) ([]FileInfo, error) // reads the dir's
//   contents and returns up to n []FileInfo, as if returned by Lstat, in
//   directory order. More calls on same file yield further FileInfo's. <br/>
//   If n > 0, Readdir returns max n []FileInfo. In this case, if Readdir
//   returns an empty slice, it will return a non-nil error explaining why.
//   At the end of a directory, the error is io.EOF. <br/>
//   If n <= 0, Readdir returns all []FileInfo from directory in single slice.
//   In this case, if Readdir succeeds (reads all the way to the end of the
//   directory), it returns the slice and a nil error. If it encounters an
//   error before the end of the directory, Readdir returns the FileInfo
//   read until that point and a non-nil error.
// - func (*File) Readdirnames(n int) (names []string, err error) // reads the
//   contents of the directory associated with arg "file" and returns max n
//   []filenames in the directory, in directory order. Subsequent calls on
//   the same file will yield further names. <br/>
//   If n > 0, Readdirnames returns at most n names. If so, if Readdirnames
//   returns an empty slice, it will return a non-nil error explaining why.
//   At the end of a directory, the error is io.EOF. <br/>
//   If n <= 0, Readdirnames returns all the names from the directory in one
//   slice. In this case, if Readdirnames succeeds (reads all the way to the
//   end of the directory), it returns the slice and a nil error. If it hits
//   an error before the end of the directory, Readdirnames returns the names
//   read until that point and a non-nil error.
// - func (f *File) Truncate(size int64) error // changes the size of the file,
//   but not the I/O offset.
//
// `os.FileInfo` c-tors:
// - func Stat(name string) (FileInfo, error) // returns a FileInfo describing
//   the named file.
// - func Lstat(name string) (FileInfo, error) // returns a FileInfo describing
//   the named file. If the file is a symlink, the returned FileInfo describes
//   the symlink. Lstat makes no attempt to follow the link.

// https://github.com/golang/go/issues/32300 (Rejected) 
// os: let FileInfo type also contain the full path, not just basename #32300

// There are two important rules for representing & handling non-files:
//  - A directory name (or path to a directory) MUST end with a slash
//    (or OS sep)
//  - We always want the immediate target of the name, and not anything
//    "downstream", like following a symlink.
//    For this reason, a symlink MUST NOT end with a slash (or OS sep).
//    This is because of the behavior of os.Lstat and os.Readlink.
//    If the name passed in ends in a slash (or OS sep?), these funcs
//    skip over the link to return the symlink's target, but NOTE that
//    only ONE link is ever skipped over; if there are multiple links,
//    in a chain, only the first is skipped over; for this reason the
//    behavior seems quite inconcsistent, so therefore we make this
//    rule that we strip off the trailing slash (or OS sep).

package fileutils

