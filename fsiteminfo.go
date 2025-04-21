package fileutils

// FSItemInfo is an interface that supplements the three
// other common interfaces implemented by [FSItem], namely
// [fs.FileInfo] and [fs.DirEntry] and [Stringser], and
// can be implemented by using only information provided
// by those three interfaces.
//
// FSItemInfo is not really used anywhere.
//
// FIXME: Index into associated list of CLI directory 
// specs that were expanded (or their ContentityFS's), 
// to preserver info about provenance. 
// .
type FSItemInfo interface {

     // BASICS

     // IsExist is a convenience function. 
     IsExist() bool
     // CreationPath is the path (abs or rel) used to create it.
     // It is tipicly implemented by an embedded [Filepaths]. 
     CreationPath() string
     // Permissions returns the standard Unix bits.
     Permissions() int

     // IsDirty has semantics TBD. 
     // IsDirty() bool
     // 
     // Refresh was sposta be a simple-to-call func to update 
     // the embedded [fs.FileInfo] and check four things:
     // existence, item type, file size, modification time.
     // However it was buggy, exhibiting unwanted recursion. 
     // Also, the func is not really necessary, cos we do not 
     // envision such a dynamic system. If this func is srsly 
     // needed, try instead to make a new instance and then 
     // compare it field-by-field to the old one.
     // Refresh() error

     // TYPE INFO

     // IsFile says whether it is a regular file,
     // i.e.. Mode().IsRegular().
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

     // NoContents means EITHER (a) it cannot have content 
     // OR (b) it can but the length of the content is zero.
     NoContents() bool
     // HasContents means BOTH (a) it can have content 
     // AND (b) the length of that content is non-zero.
     HasContents() bool
}

