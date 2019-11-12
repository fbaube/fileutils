// Package fileutils does handy stuff, like for example:
// - read in a file and figure out its content type (struct `CheckedPath`)
// - resolve a relative path into an absolute path
// - search a directory tree for certain file types
//
// Files in this directory use Markdown, so use `godoc2md` on'em.
//
// Many functions in this package are agnostic about getting an absolute
// or a relative path argument, so these functions take string arguments.
// However other functions are opinionated, so for convenience & correctness
// we define the types `AbsFilePath`, which is a new type based on `string`.
// This can be very handy in data structures, where a field `RelFilePath`
// can store a path as it was supplied by the user (or used in a file
// cross-reference), while `AbsFilePath` can represent the path as fully
// resolved. Note that the runtime resolves relative paths relative to the
// current working directory, but at least one function here takes a `WRT`
// argument that can define a different reference point.
//
// Because this package supports the processing of mixed content in
// the three markup formats supported by LwDITA (Lightweight DITA),
// it introduces the idea of an `MType`, analogous to a MIME type,
// stored as a `[3]string` slice; see file `mtype.go`
//
// Note that for simplicity and correctness, this package should
// depend as much as possible on these stdlib libraries:
// - (`path`)[https://golang.org/pkg/path/] <br/>
// functions `Base(s) Clean(s) Dir(s) Ext(s) IsAbs(s) Join(s..) Split(s) Match(..)` <br/>
// Package `path` has utility routines for manipulating slash-separated paths.
// Use only for paths separated by forward slashes, such as URL paths. This
// package does not deal with Windows paths with drive letters or backslashes;
// to do O/S paths, use `package path/filepath`
// - (`filepath`)[https://golang.org/pkg/path/filepath/] <br/>
// Functions as for `path` above plus `Abs(s) EvalSymlinks(s) FromSlash(s) 
// Glob(s) Rel(base,target) SplitList(s) ToSlash(s) VolumeName(s)
// Walk(root string, walkFn WalkFunc) type_WalkFunc`
//
package fileutils
