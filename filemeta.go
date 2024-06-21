package fileutils

// TODO maybe generalize
// it to have the funcs
//  - MustNoContent
//  - MustBeLeaf 

import (
	// SU "github.com/fbaube/stringutils"
	MU "github.com/fbaube/miscutils"
	"os"
	"fmt"
	// S "strings"
)

/* REF: os.FileInfo
Name() string       // base name of the file
Size() int64        // length in bytes for regular files; else TBS
Mode() FileMode     // file mode bits
ModTime() time.Time // modification time
IsDir() bool        // abbreviation for Mode().IsDir()
*/

/* REF: os.FileMode:
ModeDir        // d: is a directory
ModeAppend     // a: append-only
ModeExclusive  // l: exclusive use
ModeSymlink    // L: symbolic link
ModeDevice     // D: device file
ModeNamedPipe  // p: named pipe (FIFO)
ModeSocket     // S: Unix domain socket
ModeSetuid     // u: setuid
ModeSetgid     // g: setgid
ModeCharDevice // c: Unix char device, if also ModeDevice is set
ModeSticky     // t: sticky
ModeIrregular  // ?: non-regular file; nothing else is known
*/

// FileMeta (ptr to it) implements [FSItemer] (and also DirEntry?)
// and is the most basic level of file system metadata: the results
// of a call to [FP.LStat]  (or the contents of a record in sqlar
// orzip), lightly parsed.
// 
// NOTE that it is also used for directories and symlinks,
// so a more precise name would be FSItemMeta.
//
// This struct is "mostly" applicable to non-file FS nodes
// and other hierarchical structures (like XML). For example:
//  - for directories, Size() can be the number 
//    of files in it, and permissions can apply
//  - for XML elements, Size() can apply
//
// IsDir() is pass-thru. If the item is a directory, its name
// will end in a path separator (tipicly "/"). 
// 
// TODO: Size() is now pass-thru, but it could be overridden 
// for directories (to return child item count), and might be
// overridden for a file that is modifiable/dynamic in memory. 
// .
type FileMeta struct {
	os.FileInfo
	// path is used only for [Refresh] and is not exported. 
	path string
	exists bool
	// error
	MU.Errer
}

// compile-time interface check 
func init() {
     var pfm *FileMeta
     var fsir FSItemer 
     // _, ok := fm.(FSItemer)
     fsir = pfm
     // if !ok { panic("fileutils: FileMeta not implem FSItemer") }
     _ = fmt.Sprintf("pfm %T fsir %T \n", pfm, fsir)
}

// NewFileMeta replaces a call to [os.LStat]. This is necessary because
// a call of the form NewFileMeta(FileInfo) won't work because an error
// return from [os.LStat] indicates whether the file or dir (or symlink)
// exists. However no further analysis of the path is performed in this
// func, because that is more properly done by the caller.
//
// NOTE that if the file/dir does not exist, [exists] is false 
// and/but no error is indicated (i.e. [error] is nil).
//
// NOTE that by convention, directories should have a trailing 
// path separator, and it is enforced here. 
//
// NOTE not 100% sure how it behaves with relative filepaths. 
// .
func NewFileMeta(inpath string) *FileMeta {
     	if inpath == "" {
	   println("NewFileMeta GOT EMPTY PATH")
	   return nil
	   } 
	var p *FileMeta
	var e error
	p = new(FileMeta)
	p.FileInfo, e = os.Lstat(inpath)
	// There's a potential problem here, that FileInfo.Name()
	// might not be returning a trailing path sep, and it
	// might not be legal there either. So we want to rely
	// instead on the FileMeta.path
	if p.IsDir() { inpath = EnsureTrailingPathSep(inpath) }
	p.path = inpath
	p.SetError(e)
	if e == nil || !os.IsNotExist(e) {
		p.exists = true
		if p.FileInfo.Name() != inpath {
		   // NOTE false warning if they differ on trailing slash 
		   fmt.Printf("WEIRDNESS in filemeta: inpath<%s> " +
		   	filemetapath<%s>", inpath, p.FileInfo.Name())
			panic("FileMeta Problem")
			}
		// Is this necessary ? accurate ? 
		// p.exists = p.IsDir() || p.isFile() || p.isSymlink()
		// Make sure
		// p.ClearError()
	}
	return p
}

// Refresh does not check for changed type, it only checks (a) existence,
// and (b) file size, writing to stdout if either has changed.

func (p *FileMeta) Refresh() {
        pp := NewFileMeta(p.path)
	if p.Exists() != pp.Exists() {
	   fmt.Fprintf(os.Stderr, "Existence changed! (%s)", p.path) 
	}
	if p.Size() != pp.Size() {
	   fmt.Printf("size changed: %d => %d \n", p.Size(), pp.Size())
	}
	*p = *pp
}

// Exists is a convenience function.
func (p *FileMeta) Exists() bool {
	return p.exists
}

// IsEmpty is a convenience function
// for files (and directories too?).
// It can be overwritten when the file
// contents are loaded (and modifiable).
// .
func (p *FileMeta) IsEmpty() bool {
	return p.Size() == 0
}

// HasContents is the opposite of [IsEmpty].
func (p *FileMeta) HasContents() bool {
	return p.Size() != 0
}

// IsFile is a (somewhat foolproofed) convenience function.
func (p *FileMeta) IsFile() bool {
	return p.exists && p.isFile() && !p.IsDir() && !p.isSymlink()
}

// IsSymlink is a (somewhat foolproofed) convenience function.
func (p *FileMeta) IsSymlink() bool {
	return p.exists && !p.isFile() && !p.IsDir() && p.isSymlink()
}

// IsDirlike is, well, documented elsewhere.
func (p *FileMeta) IsDirlike() bool {
        return p.exists && !p.isFile()
}

func (p *FileMeta) isFile() bool {
	return p.Mode().IsRegular() && !p.IsDir()
}

func (p *FileMeta) isSymlink() bool {
	return (0 != (p.Mode() & os.ModeSymlink))
}

/*
// IsOkayDir is a (somewhat foolproofed) convenience function.
func (p *FileMeta) IsOkayDir() bool {
	return p.exists && !p.isFile() && p.IsDir() && !p.isSymL()
}
*/

