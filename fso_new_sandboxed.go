package fileutils

import (
	"io/fs"
	"os"
	"fmt"
	"errors"
	"syscall"
	FP "path/filepath"
)

/*
Methods on os.Root will follow symlinks, but
a symlink may not reference a location outside
the root, and a symlink must not be absolute.

Methods on Root allow traversal of filesystem
boundaries, Linux bind mounts, /proc special
files, and access to Unix device files.

func os.OpenInRoot(dir, name string) (*File, error)

This opens the file name in the directory dir. It is
equal to OpenRoot(dir) followed by opening the file in
the root. OpenInRoot returns an error if any component
of the name references a location outside of dir.

func OpenRoot(name string) (*Root, error)

OpenRoot opens the named directory. It follows symbolic
links in the directory name. If there is an error, it
will be of type *PathError.
*/

// NewFSObjectInRoot takes a filepath (absolute or relative) 
// and analyzes the object (assuming one exists) at the 
// path. This func does not load and analyse the content.
//
// A relative path is used w.r.t. the input [os.Root],
// so the func should be secure. If it can fetch a
// [fs.FileInfo] but the path fails to be Valid or
// Local, a loud warning ss issued.
//
// [os.Root] is relied upon, so symlinks are followed:
// [os.Root.Stat] is called rather than [os.Root.LStat].
// Thus it returns information about the symlink's target, 
// not the symlink itself. If the symlink points to a non-
// existent item, the result should be [fs.ErrNotExist].
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
func NewFSObjectSandboxed(anFP string, aRoot os.Root) *FSObject {
     	var e error
        var pEmpty = new(FSObject)
     	// Check the path
     	if anFP == "" {
	   pEmpty.SetError(errors.New("newfsitem: empty path"))
	   return pEmpty
	   }
	checkFP(anFP)

	var pFPs = NewFilepaths(anFP)
	var pPE  = new(os.PathError { Path: anFP })
	pEmpty.FPs = *pFPs
	
	if pFPs.HasError() {
	   pPE.Op = "newfilepaths"
           pPE.Err = e
           pEmpty.SetError(pPE)
           return pEmpty
	}
	// L.L.Dbg("NewFilepaths: %#v", *pFPs)
	// Because we use Stat and not LStat, we do not 
	// have to strip off any trailing slash (or OS sep). 
	// pFPs.TrimPathSepSuffixes()

	// --------------------
	//  Now we can proceed
	// --------------------
	var fi fs.FileInfo
	//     os.Lstat(pFPs.AbsFP)
	fi, e = aRoot.Stat(pFPs.AbsFP)
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

func checkFP(anFP string) {
	// *** START FOR ROOT ***
	isValid := fs.ValidPath(anFP)
	isLocal := FP.IsLocal(anFP)
	isAbsol := FP.IsAbs(anFP)
	if isAbsol  { println("******** os.Root v AbsFP: " + anFP) }
	if isValid != isLocal {
	   	      println("******** os.Root: isValid != isLocal: "+anFP) }
	if !isValid { println("******** os.Root v !Valid: " + anFP) }
	if !isLocal { println("******** os.Root v !Local: " + anFP) }
	// *** END FOR ROOT ***
}

func permString(pFI fs.FileInfo) string { 
	var perms, world, group, yuser int 
	perms = int(pFI.Mode().Perm()) // 0777 or 0x1ff
	world =  perms & 7
	group = (perms >> 3) & 7
	yuser = (perms >> 6) & 7
	var ww, gg, yu string
	ww = permStr(world)
	gg = permStr(group)
	yu = permStr(yuser)
	return fmt.Sprintf("%s,%s,%s", yu, gg, ww)
}

