package fileutils

import (
	"io/fs"
	"os"
	"errors"
	"syscall"
	// FP "path/filepath"
)

/*
// TODO replace this with New + LoadContent 
func NewFSItemWithContent(fp string) (*FSItem, error) {
	var e error 
	var pfsi *FSItem
	pfsi, e = NewFSItem(fp)
	if e != nil {
		return nil, &fs.PathError{Op:"NewFSItem",Err:e,Path:fp}
	}
	e = pfsi.LoadContents()
	if e != nil {
		return nil, &fs.PathError{
		       Op:"FSI.LoadContents", Err:e,Path:fp}
	}
	return pfsi, nil
}
*/

// NewFSItem takes a filepath (absolute or relative) and
// analyzes the object (assuming one exists) at the path.
// This func does not load and analyse the content.
//
// Note that a relative path is appended to the CWD,
// which may not be the desired behavior; in such a
// case, use NewFSItemRelativeTo (below).
//
// NOTE if no item exists at fp, this might be flakey.
//
// Note that an empty path is not OK; instead create
// an pathless FSItem from the content. 
// .
func NewFSItem(fp string) (*FSItem, error) {
     	var e error 
     	// Check the path
     	if fp == "" {
	   return nil, errors.New("NewFSItem: empty path")
	   }
	var pFPs *Filepaths 
	pFPs, e = NewFilepaths(fp)
	if e != nil {
		return nil, &fs.PathError{ Op:"NewFilepaths", Path:fp, Err:e }
	}
	var pI *FSItem
	pI = new(FSItem)
	// L.L.Dbg("NewFilepaths: %#v", *pFPs)
	pI.FPs = pFPs 
	var FI fs.FileInfo
	// Before we can call os.Lstat, we have to strip off any trailing
	// slash (or OS sep), cos it would make Lstat follow a symlink
	// (which kind of defeats the whole purpose of defining it in
	// opposition to os.Stat) 
	pFPs.TrimPathSepSuffixes()
	// Now we can proceed 
	FI, e = os.Lstat(pFPs.AbsFP)
	if e != nil {
                if errors.Is(e, fs.ErrNotExist) {
                        // Does not exist!
                        return nil, nil
                        }
                return nil, &os.PathError{
		       Op:"os.Lstat", Path:pFPs.AbsFP, Err:e }
        }
	pI.Exists = true 
	// Now we can check for a directory, and if
	// it is, add the trailing slashes back in
	if FI.IsDir() {
	   pI.FPs.EnsurePathSepSuffixes()
	   }
	// Now we try to fetch the fields that might be OS-dependent
	s, ok := FI.Sys().(*syscall.Stat_t)
        if !ok {
	       // Non-fatal error 
	       pe := &os.PathError{ Op:"parse fs.FileInfo", 
	       	      Path:pI.FPs.AbsFP, Err:errors.New(
		     "cannot convert Stat.Sys() to syscall.Stat_t") }
		pI.SetError(pe)
	       }
        var nlinks int
        nlinks = int(s.Nlink)
        if nlinks > 1 && (FI.Mode()&fs.ModeSymlink == 0) {
                // The index number of this file's inode:
                pI.Inode = int(s.Ino)
                pI.NLinks = int(s.Nlink)
        }
        // TODO: FILE PERMS
        // inode, nlinks int64
	
	
        return pI, nil
}

/*
// NewFSItemRelativeTo simply appends a relative filepath 
// to an absolute filepath being referenced, and then uses 
// it to create a new FSItem. So, it's pretty dumb.
// This func does not load & analyse the content.
func NewFSItemRelativeTo(rfp, relTo string) (*FSItem, error) {
	if !FP.IsAbs(relTo) {
		return nil, &fs.PathError{Op:"fp.isRelTo.notAbs",
		Err:errors.New("relFP must be rel to an absFP"),
		Path:fmt.Sprintf("relFP<%s>.relTo.nonAbsFP<%s>",rfp,relTo)}
	}
	afp := FP.Join(relTo, rfp)
	return NewFSItem(afp)
}
*/

func NewFSItemFromContent(s string) (*FSItem, error) {
     return nil, nil
}

// ==================

// TODO maybe generalize
// it to have the funcs
//  - MustNoContent
//  - MustBeLeaf 

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

/*
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
func NewFSItemMeta(inpath string) (*FSItem, error) {
     	if inpath == "" {
	   return nil, errors.New("NewFSItemMeta: empty path")
	   }
	// TODO: CHECK THAT PATH IS VALID 
	var fi fs.FileInfo
	var e error
	fi, e = os.Lstat(inpath)
	if e != nil {
	     	if errors.Is(e, fs.ErrNotExist) {
		   	// Does not exist!
			return nil, nil
			}
		return nil, fmt.Errorf("NewFSItemMeta<%s>: %w", inpath, e)
	}
	return NewFSItemMetaFromFileInfo(fi)
}

func NewFSItemMetaFromFileInfo(fi fs.FileInfo) (*FSItem, error) {
     	var e error 
        if fi == nil {
	   return nil, errors.New("NewFSItemMeta: empty path")
	   }
	p := new(FSItem)
	p.fi = fi 
	/*
	Fields to set:
	fs.FileInfo
	path string
	exists bool
	modTime time.Time
	inode, nlinks int64
	Errer
	* /
	var inpath string 
	inpath, e = FP.Abs(fi.Name())
	// If we got this far, we now assume that the item's path is valid. 
	// If this assumption does not hold, we be in a heap o'trouble. 
	if e != nil {
	     	return nil, fmt.Errorf("NewFSItemMetaFromFileInfo<%s>: %w",
		       inpath, e)
		}
	// CHECK FOR "/" symlink !!
	// Field inpath is only used internally, for func `Refresh`.
	// Still tho, we want to enforce trailing slashes on dirs.
	if fi.IsDir() { inpath = EnsureTrailingPathSep(inpath) }
	
	p.FPs, e = NewFilepaths(inpath) 
	// If we got this far, we now assume that the item exists.
	// If this assumption does not hold, we be in a heap o'trouble. 
	p.Exists = true
	s, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
               return p, fmt.Errorf("NewFSItemMeta: " +
	       	      "can't convert Stat.Sys() to syscall.Stat_t: %s", inpath)
	}
	var nlinks int 
	nlinks = int(s.Nlink) 
	if nlinks > 1 && (fi.Mode()&fs.ModeSymlink == 0) {
 	   	// The index number of this file's inode:
		p.Inode = int(s.Ino)
		p.NLinks = int(s.Nlink)
	}
	// TODO: FILE PERMS 
	// inode, nlinks int64
	/*
	if fi.Name() != inpath {
	   // NOTE false warning if they differ on trailing slash 
	   println(fmt.Sprintf("NewFSItemMeta: path mismatch: " +
	   	"FSitemMeta.path<%s> inpath<%s>", 
		 fi.Name(), inpath))
	} * /
	return p, nil
}

func NewFSItemMetaFromDirEntry(de fs.DirEntry) (*FSItem, error) {
     var fi fs.FileInfo
     var e error 
     fi, e = de.Info()
     if e != nil {
     	  return nil, fmt.Errorf("NewFSItemMetaFromDirEntry<%s>: %w",
	  	 de.Name(), e)
	  }     
     return NewFSItemMetaFromFileInfo(fi)
}

*/

