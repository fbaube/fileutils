// Package fileutils does handy stuff, like for example:
// - resolve a relative path into an absolute path
// - snarf up an entire file and figure out its content type
// - search a directory tree for certain file types
//
// Files in this directory use Markdown, so use `godoc2md` on 'em.
//
// Many functions are agnostic about getting an absolute or relative
// path argument, so these functions take string arguments. However
// other functions are opinionated, so for convenience and correctness
// we define two types, `AbsFilePath` and `RelFilePath`, both new types
// based on `string`. These can be very handy in data structures, where
// one of each can be used, so that `RelFilePath` can store a path as
// it was supplied by the user (or used in a file cross-reference),
// while `AbsFilePath` can represent a path as fully resolved. Note
// that the runtime resolves relative paths relative to the current
// working directory, but at least one function here takes a `WRT`
// argument that can define a different reference point.
//
// Because this package supports the processing of mixed content in
// the three markup formats supported by LwDITA (Lightweight DITA),
// it introduces the idea of an `MMCtype`, analogous to a MIME type,
// stored as a `[3]string` slice; see file `mmctype.go`
//
package fileutils
