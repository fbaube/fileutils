package fileutils

import (
	"io/fs"
	"os"
	// "fmt"
	"errors"
	"syscall"
	// FP "path/filepath"
)

// NewFSObject takes a filepath (absolute or relative) and
// analyzes the object (assuming one exists) at the path.
// This func does not load and analyse the content.
//
// A relative path is appended to the CWD,
// which may not be the desired behavior; in 
// such case, use NewFSObjectRelativeTo (below).
//
// This func does not use [os.Root], ao its security is
// not known. However this func does not follow symlinks:
// it returns information about the symbolic link itself. 
//
// There is only one return value, a pointer, always non-nil. 
// If there is an error to be returned, it is in embedded 
// struct [Errer], and the rest of the returned struct may 
// be empty and invalid, except (maybe!) embedded struct
// [FPs] (a [Filepaths]). 
//
// Note that passing in an empty path is not OK; instead
// create (by hand) a new pathless FSObject from the content.
//
// If you wish to create a blank FSObject that has no path,
// simply use a nil ptr instead of calling this func. 
// Passing an empty path to this func is not OK.
// .
func NewFSObject(anFP string) *FSObject {
     	var e error
	var pEmpty = new(FSObject)
     	// Check the path
     	if anFP == "" {
	   pEmpty.SetError(errors.New("newfsitem: empty path"))
	   return pEmpty
	   }
	   
	var pFPs = newFilepaths(anFP)
	var pPE  = new(os.PathError { Path: anFP })
	pEmpty.FPs = *pFPs
	
	if pFPs.HasError() {
	   pPE.Op = "newfilepaths"
	   pPE.Err = e
	   pEmpty.SetError(pPE)
	   return pEmpty
	}
	// L.L.Dbg("NewFilepaths: %#v", *pFPs)
	// Before we can call os.Lstat, we have to strip off any trailing
	// slash (or OS sep), cos it would make Lstat follow a symlink
	// (which kind of defeats the whole purpose of defining it in
	// opposition to os.Stat) 
	pFPs.TrimPathSepSuffixes()

	// --------------------
	//  Now we can proceed
	// --------------------	
	var fi fs.FileInfo
	fi, e = os.Lstat(pFPs.AbsFP)
	// But maybe the path does not exist !
	//  We mark this as an error. 
	if fi == nil || e != nil {
		pPE.Op = "os.lstat"
		pPE.Err = e
		pEmpty.SetError(pPE)
                if e != nil && errors.Is(e, fs.ErrNotExist) {
 		     pFPs.DoesNotExist = true 
		} 
		return pEmpty
	}
	// Now we have a valid FileInfo. From here, on we
        // can return  a valid FSObject rather than var Empty.
	var pFSI  *FSObject
	pFSI = new(FSObject)
	pFSI.FPs = *pFPs
	pFSI.FileInfo = fi
	// pFSI.Exists = true
	// Also set the time of access
	// pFSI.LastCheckTime = time.Now()

	// (Now we can) Check for a directory, and if
	// it is, add the trailing slashes back in.
	if fi.IsDir() {
	   pFSI.FPs.EnsurePathSepSuffixes()
	   }
	// Now we try to fetch the fields that might be OS-dependent
	s, ok := fi.Sys().(*syscall.Stat_t)
        if !ok {
	       // Non-fatal error
	       // FIXME: This might be difficult to debug 
	       pe := &fs.PathError{ Op:"fs.fileinfo.sys", 
	       	      Path:pFSI.FPs.AbsFP, Err:errors.New(
		     "cannot convert Stat.Sys() to syscall.Stat_t " +
		     "(should NOT be fatal!)") }
		pFSI.SetError(pe)
		// Do not return, from here 
	       }
        var nlinks int
        nlinks = int(s.Nlink)
        if nlinks > 1 && (fi.Mode()&fs.ModeSymlink == 0) {
                // The index number of this file's inode:
                pFSI.Inode  = int(s.Ino)
                pFSI.NLinks = int(s.Nlink)
        }
        // TODO: FILE PERMS
        // inode, nlinks int64

	pFSI.Perms = permString(fi)
	
        return pFSI 
}

func permStr(i int) string {
     var s string
     if 0 != (i&4) { s  = "r" } else { s  = "-" }
     if 0 != (i&2) { s += "w" } else { s += "-" }
     if 0 != (i&1) { s += "x" } else { s += "-" }
     return s
}

/*
func NewFSObjectFromContent(s string) (*FSObject, error) {
     return nil, nil
}
*/

// ==================

// TODO: Maybe generalize it to have the funcs MustNoContent & MustBeLeaf 

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

/*
TODO
Change this to FSO
Make it a func that uses the FileInfo
Delete the field from the struct
*/

// FSObjectType examines the embedded [os.FileInfo] 
// to return a value in the set of FSO_type_* 
func (p *FSObject) FSObjectType() FSO_type {
     if p.Mode().IsRegular() { return FSO_type_FILE } 
     if p.IsDir() { return FSO_type_DIRR } 
     if 0 != (p.Mode() & fs.ModeSymlink) {
     	      return FSO_type_SYML } 
     return FSO_type_OTHR 
}

