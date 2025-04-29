package fileutils

import (
	"io/fs"
	"os"
	"fmt"
	"time"
	"errors"
	"syscall"
	FP "path/filepath"
)

/*
Methods on os.Root will follow symlinks, but
a symlink may not reference a location outside
the root. A symlink must not be absolute.

Methods on Root allow traversal of filesystem
boundaries, Linux bind mounts, /proc special
files, and access to Unix device files.

func os.OpenInRoot(dir, name string) (*File, error)

This opens the file name in the directory dir. It is
equal to OpenRoot(dir) followed by opening the file in
the root. OpenInRoot returns an error if any component
of the name references a location outside of dir.
*/

// NewFSItem takes a filepath (absolute or relative) and
// analyzes the object (assuming one exists) at the path.
// This func does not load and analyse the content.
//
// A relative path is used w.r.t. the input [os.Root].
// It should therefore be secure. If it can fetch a
// [fs.FileInfo] but the path fails to be Valid or
// Local, a loud warning ss issued.
//
// [os.Root] is relied upon, so symlinks are followed:
// [os.Root.Stat] is called rather than [os.Root.LStat].
// Thus it returns information about the symlink's target, 
// not the symlink itself. If the symlink points to a 
// non-existent item, the result should be File Not Exist.
//
// There is only one return value, a pointer, always non-nil. 
// If there is an error to be returned, it is in embedded 
// struct [Errer], and the rest of the returned struct may 
// be empty and invalid, except (probably) embedded struct
// [FPs] (a [Filepaths]). 
//
// Note that passing in an empty path is not OK; instead 
// create (by hand) a new pathless FSItem from the content. 
// .
func NewFSItemSandboxed(fp string, rt os.Root) *FSItem {
     	var e error
        var Empty  *FSItem
	Empty = new(FSItem)
	
     	// Check the path
     	if fp == "" {
	   Empty.SetError(errors.New("newfsitem: empty path"))
	   return Empty
	   }
	isValid := fs.ValidPath(fp)
	isLocal := FP.IsLocal(fp)
	isAbsol := FP.IsAbs(fp)
	if isAbsol  { println("******** os.Root v AbsFP: " + fp) }
	if isValid != isLocal {
	   	      println("******** os.Root: isValid != isLocal: " + fp) }
	if !isValid { println("******** os.Root v !Valid: " + fp) }
	if !isLocal { println("******** os.Root v !Local: " + fp) }

	var pFPs *Filepaths
	pFPs, e = NewFilepaths(fp) // probably a PathError
	Empty.FPs = pFPs
	if e != nil {
	   Empty.SetError(&fs.PathError{ Op:"newfilepaths", Path:fp, Err:e })
	   return Empty
	}
	// L.L.Dbg("NewFilepaths: %#v", *pFPs)
	// Because we use Stat and not LStat, we do not 
	// have to strip off any trailing slash (or OS sep). 
	// pFPs.TrimPathSepSuffixes()

	// Now we can proceed 
	var fi fs.FileInfo
	//     os.Lstat(pFPs.AbsFP)
	fi, e = rt.Stat(pFPs.AbsFP)
	if fi == nil || e != nil {
		if e != nil && errors.Is(e, fs.ErrNotExist) {
		   Empty.SetError(&fs.PathError{ Op:"os.root.stat",
		   	 Path:pFPs.AbsFP, Err:errors.New("does not exist")})
		} else {
		   Empty.SetError(&fs.PathError{
			 Op:"os.root.stat", Path:pFPs.AbsFP, Err:e })
        	}
		return Empty
	}
	// Now we have a valid FileInfo. From here, on we 
	// can return  a valid FSItem rather than var Empty. 
	var pFSI  *FSItem
	pFSI = new(FSItem)
	pFSI.FPs = pFPs
	pFSI.FileInfo = fi
	pFSI.Exists = true
	// Also set the time of access
	pFSI.LastCheckTime = time.Now()

	// Set the FSItem_type, important for calling code. 
	pFSI.setFSItemType()
	
	// (Now we can) Check for a directory, and if
	// it is, add the trailing slashes back in. 
	if fi.IsDir() {
	   pFSI.FPs.EnsurePathSepSuffixes()
	   }
	// If it's a symbolic link, (for now) issue a big note.
	if pFSI.FSItem_type == FSItem_type_SYML {
	// func (r *Root) os.Readlink(name string) (string, error)
	   slS, slE := os.Readlink(pFSI.FPs.AbsFP)
	   fmt.Fprintf(os.Stderr, "SYMLINK: src|%s| tgt|%s| err|%s| \n",
	   	pFSI.FPs.AbsFP, slS, slE)
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

	var perms, world, group, yuser int 
	perms = int(fi.Mode().Perm()) // 0777 or 0x1ff
	world =  perms & 7
	group = (perms >> 3) & 7
	yuser = (perms >> 6) & 7
	var ww, gg, yu string
	ww = permStr(world)
	gg = permStr(group)
	yu = permStr(yuser)
	pFSI.Perms = fmt.Sprintf("%s,%s,%s", yu, gg, ww) 
        return pFSI 
}

