package fileutils

// This file implements interface [FSObjectInfo]
// defined in file fsiteminfo.go 

// NOTE maybe generalize
// it to have the funcs
//  - MustNoContent
//  - MustBeLeaf 

import (
	"time"
	"os"
	"io/fs"
	FP "path/filepath"
	L "github.com/fbaube/mlog"
)


// ===========
//  TYPE INFO 
// ===========

// IsFile is a convenience function.
func (p *FSObject) IsFile() bool {
     	return p.FileInfo.Mode().IsRegular()
}

// IsDirlike is, well, documented elsewhere.
func (p *FSObject) IsDirlike() bool {
	return p.IsDir() || (p.FSO_type == FSO_type_SYML) //p.IsSymlink
}

func (p *FSObject) IsSymlink() bool {
     	return (0 != (p.FileInfo.Mode() & os.ModeSymlink))
}

func (p *FSObject) HasMultiHardlinks() bool {
	return (p.NLinks > 1)
}

// =====================
//  EMBEDDED INTERFACES
// =====================

// Type implements [fs.DirEntry] by returning the [fs.FileMode].
func (p *FSObject) Type() fs.FileMode {
     return p.Mode()
}

/*
// DirEntryInfo implements [fs.DirEntry] by returning interface [fs.FileInfo].
// This should be named Info but it collides with interface [Stringser).
func (p *FSObject) DirEntryInfo() fs.FileInfo {
     return p.FileInfo
}
*/

func (p *FSObject) IsEmpty() bool {
	return p.IsFile() && p.Size() == 0 && !p.FPs.DoesNotExist
}

func (p *FSObject) ModTime() time.Time {
     return p.ModTime()
}

func (p *FSObject) Info() (fs.FileInfo, error) {
     return p.Info()
}

// ResolveSymlinks will follow links until it finds
// something else. NOTE that this can be a SECURITY HOLE. 
func (p *FSObject) ResolveSymlinks() *FSObject {
	if !p.IsSymlink() {
		return nil
	}
	var newPath string
	var e error
	for p.IsSymlink() {
		// func os.Readlink(pathname string) (string, error)
		// func FP.EvalSymlinks(path string) (string, error)
		newPath, e = FP.EvalSymlinks(p.FPs.AbsFP)
		if e != nil {
			L.L.Error("fu.RslvSymLx <%s>: %s", p.FPs.AbsFP, e.Error())
			// p.SetError(fmt.Errorf("fu.RslvSymLx <%s>: %w", p.FPs.AbsFP, e))
			return nil
		}
		println("--> Symlink from:", p.FPs.AbsFP)
		println("     resolved to:", newPath)
		p.FPs.AbsFP = newPath
		p = NewFSObject(newPath)
		if p.HasError() {
			panic(p.GetError())
			return nil
		}
		// CHECK IT
	}
	return p
}

/*
func FileInfoString(p fs.FileInfo) string {
     if p == nil { return "<FI:NIL>" }
     return fmt.Sprintf("%s<%s>%d", p.Name(), p.FSObject_type, p.Size())
}
*/

func (p *FSObject) IsDir() bool {
     if p == nil { return false } // "should not happen", but does 
     if p.FileInfo == nil { println("IsDir got a nil ptr") ; return false } 
     return p.FileInfo.IsDir()
}

