package fileutils

import(
	"os"
	"io/fs"
)

// GatherDirTreeList walks a directory tree (fetched via [os.DirFS]) to 
// gather a list of all item names (file, directories, symlinks, more).
// 
// NOTE that for the argument "inpath", it makes no difference whether:
//  - inpath is relative or absolute
//  - inpath ends with a trailing slash or not
//  - inpath is a directory or a symlink to a directory
// 
// In the walk process, the item names (as returned by [fs.Walkdir]) are
// relative to `inpath` and do not include any information about `inpath`
// itself, i.e. the portions of the item paths "above" the directory node
// `inpath`, so the caller has to sort that out.
//
// Regarding values of the argument `inpath`: 
//  - A valid directory returns at least one item: 
//    ".", representing `inpath` itself.
//  - A valid file or non-existent path item returns
//    only one item, the `inpath` itself.
//  - A symlink to a directory it is followed; behavior 
//    for a symlink to a file is not easily summarised. 
// 
// The docu for [os.Dirfs] states:
// The result implements io/fs.StatFS, io/fs.ReadFileFS and io/fs.ReadDirFS.
//
// Therefore API calls that work are:
//  - Stat errors should be of type *PathError.
//  - Stat(name string) (fs.FileInfo, error)
//  - Readfile on success returns a nil error, not [io.EOF].
//    The caller is permitted to modify the returned byte slice.
//    This method should return a copy of the underlying data.
//  - ReadFile(name string) ([]byte, error)
//  - ReadDir reads the named directory and returns a list of
//    directory entries sorted by filename.
//    ReadDir(name string) ([]fs.DirEntry, error)
// . 
func GatherDirTreeList(path string) (paths []string) {
     	fsys := os.DirFS(path)
	fs.WalkDir(fsys, ".",
		func(pathbase string, de fs.DirEntry, e error) error {
			paths = append(paths, pathbase)
		return nil
	})
	return paths
}

