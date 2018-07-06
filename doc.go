// Package fileutils does handy stuff, like resolve a relative path into
// a absolute path, and snarf up an entire file and figure out its content
// type, and search a directory tree for certain file types.
//
// Because this package supports the processing of mixed content in
// the three markup formats supported by LwDITA (Lightweight DITA),
// it introduces the idea of an `MMCtype`, analogous to a MIME type,
// stored as a `[3]string` slice; see file `mmctype.go`
//
// To avoid usage errors, it defines `AbsFilePath` and `RelFilePath`
// as new types based on `string`.
//
package fileutils
