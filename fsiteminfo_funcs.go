package fileutils

// This file implements interface [FSItemInfo]
// defined in file fsiteminfo.go 

// NOTE maybe generalize
// it to have the funcs
//  - MustNoContent
//  - MustBeLeaf 

import (
	"io/fs"
	"os"
	"fmt"
)

/* REF: re. interface [fs.DirEntry]:
// Info may be from the time of either the original directory read
// OR the call to Info.
// If the file has been removed or renamed since the directory read,
// Info may return an error satisfying `errors.Is(err, ErrNotExist)`.
// If the entry denotes a symlink, Info reports information about
// the link itself, like [os.Lstat] does, and not the symlink's
// target, ALTHO this assume that the call to [os.Lstat] did not
// have a trailling slash (or OS sep).
Info() (FileInfo, error)
*/

/*
// init is a compile-time interface check FIXME
func init() {
     var pfm *FSItemMeta
     var fsir FSItemer 
     // _, ok := fm.(FSItemer)
     fsir = pfm
     // if !ok { panic("fileutils: FSItemMeta not implem FSItemer") }
     _ = fmt.Sprintf("pfm %T fsir %T \n", pfm, fsir)
}

func (p *FSItemMeta) String() string {
     if !p.Exists() { return p.Name() + ":notExist" }
     var e, sz string
     if p.HasError() { e = "<err:" + p.Error() + ">" }
     if FileInfoTLC(p.FileInfo) == "FIL" { sz = fmt.Sprintf("[%d]", p.Size()) }
     return fmt.Sprintf("%s: %s %s %s",
     	    p.Name(), FileInfoTLC(p.FileInfo), sz, e)
}

func (p *FSItemMeta) StringWithPermissions() string {
     if !p.Exists() { return p.Name() + ":notExist" }
     var e, sz string
     if p.HasError() { e = "<err:" + p.Error() + ">" }
     if FileInfoTLC(p.FileInfo) == "FIL" { sz = fmt.Sprintf("[%d]", p.Size()) }
     perms := fmt.Sprintf("%09b", p.FileInfo.Mode() & 0x1ff)
     return fmt.Sprintf("%s: %s %s %s %s",
     	    p.Name(), FileInfoTLC(p.FileInfo), sz, e, perms)
}
*/

// ========
//  BASICS
// ========

// IsExist is updated by [Refresh] but any
// error returned by Refresh is ignored. 
func (p *FSItem) IsExist() bool {
     	p.Refresh() // ignore error return 
	return p.Exists
}

// IsDirty is TBD. 
func (p *FSItem) IsDirty() bool {
	return p.Dirty
}

// Refresh updates the embedded [fs.FileInfo] and checks four things: 
// existence, item type, file size, and modification time. Details:
//  - A file coming into existence or a file being appended to might 
//    be common use cases.
//  - In general, if any of the four things has changed, it writes a
//    warning to stdout and in some cases returns an [os.PathError].
//  - If [Dirty] is set, some warnings do not apply.
//  - If there is already an error, this call is ignored.
// . 
func (p *FSItem) Refresh() error {
     	// Refreshable ? 
     	if !(p.Exists && (nil != p.FPs)) { return nil }
     	crePath := p.FPs.CreationPath()
        pp, e := NewFSItem(crePath)
	if pp == nil && e != nil {
	   fmt.Fprintf(os.Stderr, "FSItem.Refresh<%s> failed: %w \n",
	   	p.fi.Name(), e)
	}
	if p.IsExist() != pp.IsExist() {
	   fmt.Fprintf(os.Stderr, "Existence changed! (%s) \n", crePath)
	}
	if p.fi.Size() != pp.fi.Size() {
	   fmt.Printf("Size changed! (%s) %d => %d \n",
	   	crePath, p.fi.Size(), pp.fi.Size())
	}
	if !p.fi.ModTime().Equal(pp.fi.ModTime()) {
	   fmt.Printf("ModTime changed! (%s) %s => %s \n",
	   	crePath, p.fi.ModTime(), pp.fi.ModTime())
	}
	*p = *pp
	return nil
}

// ===========
//  TYPE INFO 
// ===========

// FICode4L returns a 4-Letter code: FILE, DIRR, SYML, OTHR.
// A FileInfo that is one of multiple hard links to a
// particular inode is not distinguished as such here. 
func FICode4L(p fs.FileInfo) string {
     if p == nil  { return "_NIL" }
     if p.IsDir() { return "DIRR" }
     // func (m FileMode) IsRegular() bool
     if p.Mode().IsRegular() { return "FILE" }
     if 0 != (p.Mode() & fs.ModeSymlink) { return "SYML" }
     return "OTHR"
}

// IsFile is a convenience function.
func (p *FSItem) IsFile() bool {
	return p.Exists && p.isFile() && !p.fi.IsDir() && !p.isSymlink()
}

// IsDirlike is, well, documented elsewhere.
func (p *FSItem) IsDirlike() bool {
        if !p.Exists { return false }
	return p.fi.IsDir() || p.IsSymlink()
}

// IsSymlink is a convenience function.
func (p *FSItem) IsSymlink() bool {
	return p.Exists && !p.isFile() && !p.fi.IsDir() && p.isSymlink()
}

func (p *FSItem) HasMultiHardlinks() bool {
	return (p.NLinks > 1)
}

// -----------------------------
//  TYPE INFO utility functions
// -----------------------------

func (p *FSItem) isFile() bool {
	return p.fi.Mode().IsRegular() && !p.fi.IsDir()
}

func (p *FSItem) isSymlink() bool {
	return (0 != (p.fi.Mode() & os.ModeSymlink))
}

// =====================
//  EMBEDDED INTERFACES
// =====================

// Type implements [fs.DirEntry] by returning the [fs.FileMode].
func (p *FSItem) Type() fs.FileMode {
     return p.fi.Mode()
}

// DirEntryInfo implements [fs.DirEntry] by returning interface [fs.FileInfo].
// This should be named Info but it collides with interface [Stringser).
func (p *FSItem) DirEntryInfo() fs.FileInfo {
     return p.fi
}

// IsEmpty is a convenience function for files (and directories too?).
func (p *FSItem) IsEmpty() bool {
	return p.fi.Size() == 0 || !p.IsFile()
}

// HasContents is the opposite of [IsEmpty].
func (p *FSItem) HasContents() bool {
	return p.isFile() && p.fi.Size() > 0
}

