package fileutils

// TODO maybe generalize
// it to have the funcs
//  - MustNoContent
//  - MustBeLeaf 

import (
	"io/fs"
	"os"
	"fmt"
	"errors"
	"time"
	"syscall"
)

/* REF: ifc fs.FileInfo
Name() string       // base name of the item
Size() int64        // length in bytes for regular files; else TBS
Mode() FileMode     // file mode bits
ModTime() time.Time // modification time
IsDir() bool        // abbreviation for Mode().IsDir()
Sys()   any

REF: ifc fs.DirEntry
Name() string // base name of the item
IsDir() bool
// Type returns a subset of the usual FileMode bits returned by [FileMode.Type] 
Type() FileMode
// Info may be from the time of either (a) the original directory read, 
// or (b) the call to Info. If the file has been removed or renamed since 
// the directory read, Info may return an error satisfying  errors.Is(err,
// ErrNotExist). If the entry denotes a symlink, Info reports information
// about the link itself, like [Lstat] does, and not the symlink's target.
Info() (FileInfo, error)
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

// FSItemMeta implements [FSItemer] (and also [fs.FileInfo] and
// [fs.DirEntry]) and is basic file system metadata: the results 
// of a call to [fs.LStat] (or the contents of a record in sqlar
// or zip), lightly parsed.
// 
// This struct is "sorta" applicable to non-file FS nodes and
// also other hierarchical structures (like XML). For example:
//  - for directories, Size() can be the number 
//    of files in it, and permissions can apply
//  - for XML elements, Size() can apply
//
// [IsDir] and [ModTime] are pass-thru. If the item is a directory, 
// its name should (and: MUST) end in a path separator (tipicly "/"). 
// 
// TODO: Size() is now pass-thru, but it could be overridden 
// for directories (to return child item count), and might be
// overridden for a file that is modifiable/dynamic in memory. 
// .
type FSItemMeta struct {
	fs.FileInfo
	// path is used only for [Refresh] and is not exported. 
	path string
	exists bool
	// store this to hide OS dependencies
	modTime time.Time
	// hard link detection 
	inode, nlinks int // uint64
	// NPE-proof error
	Errer
}

// compile-time interface check 
func init() {
     var pfm *FSItemMeta
     var fsir FSItemer 
     // _, ok := fm.(FSItemer)
     fsir = pfm
     // if !ok { panic("fileutils: FSItemMeta not implem FSItemer") }
     _ = fmt.Sprintf("pfm %T fsir %T \n", pfm, fsir)
}

// NewFSItemMeta replaces a call to [os.LStat]. This is necessary because
// a call of the form NewFSItemMeta(FileInfo) won't work because an error
// return from [os.LStat] indicates whether the file or dir (or symlink)
// exists. However no further analysis of the path is performed in this
// func, because that is more properly done by the caller.
//
// NOTE that if the file/dir does not exist, this returns (nil, nil).
//
// NOTE that if some fields are unavailable due to portability issues,
// this returns (non-nil, non-nil), so the error should not be fatal.
//
// NOTE that by convention, directories should (welll, MUST) 
// have a trailing path separator, and it is enforced here. 
//
// NOTE not 100% sure how it behaves with relative filepaths. 
// .
func NewFSItemMeta(inpath string) (*FSItemMeta, error) {
     	if inpath == "" {
	   println("NewFSItemMeta GOT EMPTY PATH")
	   return nil, errors.New("Empty path")
	   }
	var p *FSItemMeta
	var e error
	p = new(FSItemMeta)
	/*
	Fields to set:
	fs.FileInfo
	path string
	exists bool
	modTime time.Time
	inode, nlinks int64
	Errer
	*/
	p.FileInfo, e = os.Lstat(inpath)
	// There's a potential problem here, that FileInfo.Name()
	// might not be returning a trailing path sep, and it
	// might not be legal there either. So we want to rely
	// instead on the FSItemMeta.path
	if p.IsDir() { inpath = EnsureTrailingPathSep(inpath) }
	p.path = inpath
	p.SetError(e)
	if e != nil {
	     	if errors.Is(e, fs.ErrNotExist) {
		   	// Does not exist !
			return nil, nil
			}
		p.exists = false // redundant, but let's 
		        // be clear: we really don't know
		return nil, fmt.Errorf("NewFSItemMeta<%s>: %w", inpath, e)
	}
	p.exists = true
	p.modTime = p.FileInfo.ModTime()
	s, ok := p.FileInfo.Sys().(*syscall.Stat_t)
	if !ok {
               return p, fmt.Errorf("NewFSItemMeta: " +
	       	      "can't convert Stat.Sys() to syscall.Stat_t: %s", inpath)
	}
	var nlinks int 
	nlinks = int(s.Nlink) 
	if nlinks > 1 && (p.FileInfo.Mode()&fs.ModeSymlink == 0) {
 	   	// The index number of this file's inode:
		p.inode = int(s.Ino)
		p.nlinks = int(s.Nlink)
	}		
	// inode, nlinks int64
	if p.FileInfo.Name() != inpath {
	   // NOTE false warning if they differ on trailing slash 
	   println(fmt.Sprintf("NewFSItemMeta: path mismatch: " +
	   	"inpath<%s> FSitemmetapath<%s>",
		 inpath, p.FileInfo.Name()))
		 panic("FSItemMeta Problem")
	}
	return p, nil
}

// Refresh does not check for changed type, it only checks (a) existence,
// and (b) file size, writing to stdout if either has changed.

func (p *FSItemMeta) Refresh() {
        pp, e := NewFSItemMeta(p.path)
	if pp == nil && e != nil {
	   fmt.Fprintf(os.Stderr, "FSItemMeta.Refresh<%s> failed: %w",
	   	p.Name(), e)
	}
	if p.Exists() != pp.Exists() {
	   fmt.Fprintf(os.Stderr, "Existence changed! (%s)", p.path) 
	}
	if p.Size() != pp.Size() {
	   fmt.Printf("Size changed! (%s) %d => %d \n",
	   	p.path, p.Size(), pp.Size())
	}
	*p = *pp
}

// Exists is a convenience function.
func (p *FSItemMeta) Exists() bool {
	return p.exists
}

// IsEmpty is a convenience function
// for files (and directories too?).
// It can be overwritten when the file
// contents are loaded (and modifiable).
// .
func (p *FSItemMeta) IsEmpty() bool {
	return p.Size() == 0
}

// HasContents is the opposite of [IsEmpty].
func (p *FSItemMeta) HasContents() bool {
	return p.Size() != 0
}

// IsFile is a (somewhat foolproofed) convenience function.
func (p *FSItemMeta) IsFile() bool {
	return p.exists && p.isFile() && !p.IsDir() && !p.isSymlink()
}

// IsSymlink is a (somewhat foolproofed) convenience function.
func (p *FSItemMeta) IsSymlink() bool {
	return p.exists && !p.isFile() && !p.IsDir() && p.isSymlink()
}

// IsDirlike is, well, documented elsewhere.
func (p *FSItemMeta) IsDirlike() bool {
        return p.exists && !p.isFile()
}

func (p *FSItemMeta) isFile() bool {
	return p.Mode().IsRegular() && !p.IsDir()
}

func (p *FSItemMeta) isSymlink() bool {
	return (0 != (p.Mode() & os.ModeSymlink))
}

func (p *FSItemMeta) HasMultiHardlinks() bool {
	return (p.nlinks > 1)
}

// Type implements [fs.DirEntry] by returning the [fs.FileMode].
func (p *FSItemMeta) Type() fs.FileMode {
     return p.Mode()
}

// Info implements [fs.DirEntry] by returning the interface [fs.FileInfo].
func (p *FSItemMeta) Info() fs.FileInfo {
     return p.FileInfo
}

/*
// IsOkayDir is a (somewhat foolproofed) convenience function.
func (p *FSItemMeta) IsOkayDir() bool {
	return p.exists && !p.isFile() && p.IsDir() && !p.isSymL()
}
*/

