package fileutils

// FSItemer is implemented by *[FileMeta], which embeds [os.FileInfo].
type FSItemer interface {
     // Exists is a convenience function.
     Exists() bool 
     // Refresh does not check for changed type, it only checks
     // (a) existence, and (b) file size, writing to stdout if
     // either has changed. 
     Refresh() 
     // IsFile is a (somewhat foolproofed) convenience function.
     IsFile() bool 
     // IsDir is a (somewhat foolproofed) convenience function.
     IsDir() bool 
     // IsDirlike means (a) it can NOT contain own content 
     // and (b) it is/has link(s) to other items (meaning 
     // it is a directory or a symlink).
     IsDirlike() bool 
     // IsSymlink is a (somewhat foolproofed) convenience function.
     IsSymlink() bool
     // IsEmpty uses file length.
     IsEmpty() bool
     // HasContents uses file length.
     HasContents() bool
     // HasMultiHardlinks might not be portable.
     HasMultiHardlinks() bool 
}

