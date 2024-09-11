package fileutils

import "io/fs"

// FSItemInfo is an interface that extends two common interfaces,
// [fs.FileInfo] and [fs.DirEntry], and can be implemented by
// using only information provided by those two interfaces. 
// .
type FSItemInfo interface {

     fs.FileInfo
     fs.DirEntry 

     // BASICS

     // IsExist is a convenience function, updated by [Refresh].
     IsExist() bool
     // CreationPath is the path (abs or rel) used to create it.
     // It is implemented by the embedded [Filepaths]. 
     CreationPath() string
     // IsDirty has semantics TBD. 
     IsDirty() bool
     // Refresh updates the embedded [fs.FileInfo] and checks four 
     // things: existence, item type, file size, modification time.     
     Refresh() error
     // Permissions returns the standard Unix bits.
     Permissions() int 

     // TYPE INFO

     // Code4L returns one of "FILE", "DIRR", "SYML", "OTHR".
     Code4L() string 
     // IsFile says whether it is a regular file. 
     IsFile() bool 
     // IsDir says whether it is a directory. It is
     // pass-thru from the embedded [fs.FileInfo].
     IsDir() bool 
     // IsDirlike means (a) it can NOT contain own content and
     // (b) it is/has link(s) to other items that can be further
     // examined (meaning: it is a directory or a symlink).
     IsDirlike() bool 
     // IsSymlink is a convenience function.
     IsSymlink() bool
     // HasMultiHardlinks might not be portable.
     HasMultiHardlinks() bool 

     // CONTENT-RELATED

     // IsEmpty means either (a) it cannot have content OR 
     // (b) it can but the length of the content is zero.
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

