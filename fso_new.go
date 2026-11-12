package fileutils

import (
	"io/fs"
	"os"
	"fmt"
	"errors"
	"syscall"
    FP "path/filepath"
)

// NewFSObject takes a filepath (absolute or relative) and an
// optional [os.Root], and analyzes the object at the path (if 
// it exists). It does not load and analyse the content.
//
// A relative path is either appended to the CWD or resolved
// by a non-nil Root. Using the CWD may not be the desired
// behavior, in which case, use NewFSObjectRelativeTo. 
//
// # Return value
// 
// There is only one return value, a pointer, always non-nil. 
// If there is an error to be returned, it is in the embedded 
// struct [Errer], and the rest of the returned struct may be
// empty and invalid, except (maybe!) embedded struct [FPs]
// (a [Filepaths]). 
//
// Note that passing in an empty path is not OK; instead
// create ("by hand") a new pathless FSObject from the content.
//
// If you wish to create a blank FSObject that has no path,
// simply use a nil ptr instead of calling this func. 
// Passing an empty path to this func is not OK.
//
// " Root is nil => CWD is Used ; also Lstat ???
// 
// When Root is nil, the security of this func is not known.
// However this func does not follow symlinks: it returns
// information about the symbolic link itself. 
//
// # Root is not nil
// 
// When Root is not nil, a relative path is used w.r.t. the
// input [os.Root]. If the path fails to be Valid or Local,
// a loud warning ss issued.
//
// [os.Root] is relied upon, so symlinks are followed:
// [os.Root.Stat] is called rather than [os.Root.LStat].
// Thus it returns information about the symlink's target,
// not the symlink itself. If the symlink points to a non-
// existent item, the result should be [fs.ErrNotExist].
//
// Methods on os.Root will follow symlinks, but a symlink
// may not reference a location outside the root, and a
// symlink must be a relative path, not ever an absolute.
//
// Methods on Root allow traversal of filesystem boundaries,
// Linux bind mounts, /proc special files, and access to
// Unix device files.
//
// func (r *Root) Name() string
//
// Name returns the name of the directory presented to OpenRoot.
// It is safe to call Name after [Close].
// 
// func os.OpenInRoot(dir, name string) (*File, error)
//
// This opens the file name in the directory dir. It is
// equal to OpenRoot(dir) followed by opening the file in
// the root. OpenInRoot returns an error if any component
// of the name references a location outside of dir.
//
// func OpenRoot(name string) (*Root, error)
//
// OpenRoot opens the named directory. It follows symbolic
// links in the directory name. If there is an error, it
// will be of type *PathError.
// .
func NewFSObject(anFP string, aRoot *os.Root) *FSObject {
    var e error
	var pEmpty = new(FSObject)
    // Check the path
    if anFP == "" {
	   pEmpty.SetError(errors.New("newfso: empty path"))
	   return pEmpty
	}
    checkFP(anFP)
    
	var pFPs *Filepaths
    // We need a filepath based on anFP: 
    //  - If it's abs, that's great
    //  - If it's rel & no Root, use CWD
    //  - If it's rel & Root, use the Root's "CWD"
    if aRoot != nil && !FP.IsAbs(anFP) {
       anFP = FP.Join(aRoot.Name(), anFP)
       }
    pFPs = NewFilepaths(anFP)
	var pPE  = new(os.PathError { Path: anFP })
	pEmpty.FPs = *pFPs
	
	if pFPs.HasError() {
	   pPE.Op = "newFPs"
	   pPE.Err = e
	   pEmpty.SetError(pPE)
	   return pEmpty
	}

    // Stat/Lstat discussion: We use Lstat only: 
    // func (r *Root) Lstat(name string) (FileInfo, error)
    //      If the target is a symlink, the FileInfo des cribes
    //      the symlink. See Lstat for more details.
    // func Lstat(name string) (FileInfo, error)
    //      If the file is a symlink, the FileInfo describes
    //      the symlink; Lstat does not follow the link.
    //      If error, it is a *PathError.
    // func (r *Root) Stat(name string) (FileInfo, error)
    //      See Stat for more details.
    // func Stat(name string) (FileInfo, error)
    //      If error, it is a *PathError.

	// L.L.Dbg("NewFilepaths: %#v", *pFPs)
	// To call [os.Lstat], we strip off any trailing slash
    // (or OS sep), cos it might make Lstat follow a symlink
    // (which kinda defeats the whole purpose of defining it
    // in opposition to os.Stat)
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
    // ----------------------------------------------
    // If it's a symlink, (for now) issue a big note.
    if pFSI.FSO_type == FSO_type_SYML {
    // func (r *Root) os.Readlink(name string) (string, error)
       slS, slE := os.Readlink(pFSI.FPs.AbsFP)
       fmt.Fprintf(os.Stderr, "SYMLINK: src|%s| tgt|%s| err|%s| \n",
        pFSI.FPs.AbsFP, slS, slE)
    }
    // -----------------------------------------------
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

