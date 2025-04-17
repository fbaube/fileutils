package fileutils

import "io/fs"

// FSItemInfo is an interface that extends two common interfaces,
// [fs.FileInfo] and [fs.DirEntry], and can be implemented by
// using only information provided by those two interfaces.
//
// FIXME: Maybe these first three sub-interfaces should just
// be listed at struct [FSInfo] as being implemented by it;
// its embedded [fs.FileInfo] does a lot of the work, and an
// embedded [Errer] would also; also, don't forget interface
// [stringutils.Stringser] too, as defined for an FSItem,
// separate from defining Stringser for a Contentity. 
// FSItemInfo is not really used anywhere, so [FSItem] 
// could just have a func init that verifies that the 
// four(!) interfaces are implemented. 
//
// FIXME: Index into associated list of CLI directory specs that
// were expanded (or their ContentityFS's), to preserver info
// about provenance. 
// .
type FSItemInfo interface {
/*
        // type FileInfo interface {
	// Name() string   // base name of the file
	// Size() int64    // length in bytes for regular files; else system-dep
	// Mode() FileMode // file mode bits
	// ModTime() time.Time // mod time
	// IsDir() bool    // abbrev for Mode().IsDir()
	// Sys() any       // info from underlying data source (can be nil)
	fs.FileInfo

	// type DirEntry interface {
	// Name() string   // base name 
	// IsDir() bool    // 
	// Type() FileMode // a subset of the usual FileMode bits: FileMode.Type
	// Info() (FileInfo, error) // if symlink, is re link, not its target 
	fs.DirEntry

	// type Error struct w methods: 
	// func (p *Errer) Error() string {
	// func (p *Errer) SetError(e error) {
	// func (p *Errer) GetError() error {
	// func (p *Errer) HasError() bool {
	// func (p *Errer) HasPathError() bool {
	// func (p *Errer) GetPathError() *fs.PathError {
	// func (p *Errer) ClearError() {
	Errer 
*/
     // BASICS

     // IsExist is a convenience function. // ,updated by [Refresh].
     IsExist() bool
     // CreationPath is the path (abs or rel) used to create it.
     // It is tipicly implemented by an embedded [Filepaths]. 
     CreationPath() string
     // Permissions returns the standard Unix bits.
     Permissions() int 
     // IsDirty has semantics TBD. 
     IsDirty() bool
     // Refresh (buggy: TBS) updates the embedded [fs.FileInfo] and checks 
     // four things: existence, item type, file size, modification time.
     // NOTE: This is not really necessary, cos we do not envision such
     // a dynamic system. If this is srsly needed, try instead to make 
     // a new instance and then compare it field-by-field to the old one.
     // Refresh() error

     // TYPE INFO

     // FICode4L returns one of "FILE", "DIRR", "SYML", "OTHR".
     FICode4L() string 
     // IsFile says whether it is a regular file. 
     IsFile() bool 
     // IsDir says whether it is a directory. It is
     // pass-thru from the embedded [fs.FileInfo].
     IsDir() bool 
     // IsDirlike means (a) it can NOT contain own content and
     // (b) it is/has link(s) to other items that can be further
     // examined; this all means: it is a directory or a symlink.
     IsDirlike() bool 
     // IsSymlink is a convenience function.
     IsSymlink() bool
     // HasMultiHardlinks might not be portable.
     HasMultiHardlinks() bool 

     // CONTENT-RELATED

     // IsEmpty means either (a) it cannot have content OR 
     // (b) it can but the length of the content is zero.
     // FIXME: Rename to NoContents()
     IsEmpty() bool
     // HasContents means both (a) it can have content 
     // AND (b) the length of that content is non-zero.
     HasContents() bool

     // EMBEDDED INTERFACES

     // DirEntryInfo implements [fs.DirEntry] by returning interface
     // [fs.FileInfo]. This should be named Info as in [fs.DirEntry.Info]
     // but that would collide with interface [Stringser).
     DirEntryInfo() fs.FileInfo 
}

