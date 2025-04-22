package fileutils

import (
	"fmt"
	"errors"
	"io"
	"os"
	"time"
	"io/fs"
	"crypto/md5"
	FP "path/filepath"
	CT "github.com/fbaube/ctoken"
	SU "github.com/fbaube/stringutils"
	L "github.com/fbaube/mlog"
)

// MAX_FILE_SIZE is set (arbitrarily) to 100 megabytes
const MAX_FILE_SIZE = 100_000_000

func init() {
     var fsi *FSItem
     fsi = &(FSItem {})
     var de fs.DirEntry  // ifc
     var sr SU.Stringser // ifc
     // var okde, oksr bool 
     de /* ,okde */ = fs.DirEntry(fsi)
     sr /* ,oksr */ = SU.Stringser(fsi)
     // if ! (okde && oksr) { panic("FSItem ifc's") }
     fmt.Printf("DirrEntry: %v \n", de)
     fmt.Printf("Stringser: %v \n", sr)
}

// FSItem is an item identified by a filepath (plus its contents) 
// that we have tried to or will try to read, write, or create. It 
// might be a directory or symlink, either of which requires further
// processing elsewhere. In the most common usage, it is a file.
//
// It implements four interfaces:
//  - [fs.FileInfo]
//  - [fs.DirEntry]
//  - [Errer] (actually, via an embed) 
//  - [stringutils.Stringser] (Echo, Infos, Debug)
//
// It might be just a path where nothing exists but we intend to do
// something. Its filepath(s) can be empty ("") if (for example) its
// content was created interactively or it so far lives only in memory.
//
// NOTE basically all fields are exported. This will change in the 
// future when the handlng of modifications is tightened up. 
//
// NOTE that the file name (aka [FP.Base], the part of the full path after
// the last directory separator) is not stored separately: it is stored in
// the AbsFP *and* the RelFP. Note also that this path & name information
// duplicates what is stored in an instance of orderednodes.Nord . 
//
// NOTE that it embeds an [fs.FileInfo], and implements interfaces [FSItemer],
// [fs.FileInfo], and [fs.DirEntry]), and contains basic file system metadata
// PLUS the path to the item (whicih FIleInfo does not contain) AND the item
// contents (but only after lazy loading). The `FileInfo` is the results of
// a call to [os.LStat]/[fs.Lstat] (or perhaps alternatively the contents
// of a record in sqlar or zip), parsed.
// 
// FSItem is embedded in struct [datarepo/rowmodels/ContentityRow].
//
// This struct is rather large and all-encompassing, but this follows from
// certain design decisions and certain behavior in the stahdard library.
// 
// It might seem odd to include a [TypedRaw] rather than a plain [Raw].
// But in general when we are working with serializing and deserializing
// content ASTs, it is important to know what we are working with, cos
// sometimes we can - or want to - have to - do things like include
// HTML in Markdown, or permit HTML tags in LwDITA.
//
// It might also seem odd that MU_type_DIRLIKE is a "markup type",
// but this avoids many practival problems encountered in trying 
// to process file system trees.
//
// NOTE that RelFP and AbsFP must be exported to be persisted to the DB.
//
// This struct might be somehow applicable to non-file FS nodes and also
// other hierarchical structures (like XML), but this is not explored yet.
// .
type FSItem struct { // this has (Typed) Raw
     	// LastCheckTime is TBS.
	LastCheckTime time.Time 
     	// FileInfo should be an unexported, lower case "fi", 
	// because it is relied on heavily and updated often 
	// and carefully; also it is implementing interfaces
	// FileInfo, DirEntry, FSItemer.
	fs.FileInfo
	// FSItem_type is closely linked to FileInfo and 
	// they should always be updated in lockstep.
	FSItem_type
	// TypedRaw is a ptr, to allow for lazy loading.
	*CT.TypedRaw
	// FPs is a ptr, to allow for items that are not (yet) on disk 
	// or are kept only in memory. Each path includes the [FP.Base].
	// Paths are used mainly for func [Refresh] and for reproducing
	// the tree structure of import batches; other uses TBD. 
	// 
	// Paths follow our rules:
	//  - a directory MUST end in a slash (or OS sep)
	//  - a symlink MUST NOT end in a slash (or OS sep)
	// 
	// Note that an [fs.FileInfo] does not preserve or provide path
	// info, which is part of the motivstion for this large struct. 
	FPs *Filepaths
	// Exists is false when [os.Lstat] returns ´(nil, nil)´. 
	Exists bool
	// Dirty has semantics TBD.
	Dirty bool
	// Perms is UNIX-style "rwx" user/group/world
	Perms string 
	// Inode and NLinks are for hard link detection. 
	Inode, NLinks int // uint64
	// Errer provides an NPE-proof error field
	Errer
}

func (p *FSItem) IsDir() bool {
     if p == nil { return false } // "should not happen", but does 
     if p.FileInfo == nil { println("IsDir got a nil ptr") ; return false } 
     return p.FileInfo.IsDir()
}

// ResolveSymlinks will follow links until it finds
// something else. NOTE that this can be a SECURITY HOLE. 
func (p *FSItem) ResolveSymlinks() *FSItem {
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
		p = NewFSItem(newPath)
		if p.HasError() {
			panic(p.GetError())
			return nil
		}
		// CHECK IT
	}
	return p
}

// LoadContents reads the file (assuming it is a file) into the field
// [TypedRaw], takes the hash, and quickly checks for XML and HTML5
// declarations.
//
// Before proceeding it calls [Refresh], just in case.
//
// It is tolerant about non-files, and empty files,returning nil for error.
//
// NOTE the call to [os.Open] defaults to R/W mode, altho R/O might suffice.
// .
func (p *FSItem) LoadContents() error {
     	var e error 
	/*
	// println("LoadContents: Entering!")
     	// Update the metadata (fs.FileInfo)
	// OOPS Causes infinite recursion !!
	e = p.Refresh()
	if e != nil {
	     p.SetError(e)
	     return &fs.PathError{
	     	    Op:"LoadContents.Refresh", Path:p.FPs.AbsFP, Err:e }
	}
	*/
	// println("LoadContents: chkpt 1")
	var shortFP = p.FPs.ShortFP
	if p.TypedRaw != nil {
		L.L.Warning("pp.LoadContents: already "+
			"loaded [%d]: %s", len(p.Raw), shortFP)
		// println("LoadContents: already loaded")
		// >> return nil
	}
	// Allocate this to prevent NPEs
	p.TypedRaw = new(CT.TypedRaw)
	if !p.IsFile() {
		// No-op
		// println("LoadContents: not a file")
		p.TypedRaw.Raw_type = SU.Raw_type_DIRLIKE 
		return nil
	}
	if p.Size() == 0 {
		// No-op
		// println("LoadContents: file size zero")
		p.TypedRaw.Raw_type = SU.Raw_type_NIL
		return nil
	} else if p.Size() < 6 { // Suspiciously tiny ?
		L.L.Warning("pp.LoadContents: tiny "+
			"file [%d]: %s", p.Size(), shortFP)
		p.TypedRaw.Raw_type = SU.Raw_type_NIL
	}
	// println("LoadContents: chkpt 2")
	// If it's too big, BARF!
	if p.Size() > MAX_FILE_SIZE {
		return &fs.PathError{Op:"FSI.LoadContents",
		       Err:errors.New(fmt.Sprintf(
		       "file too large: %d", p.Size())), Path:shortFP}
	}
	// println("LoadContents: chkpt 3")
	// Open it, just to check (and then immediately close it)
	var pF *os.File
	pF, e = os.Open(p.FPs.AbsFP)
	// Note that this defer'd Close() (i.e. the file is left open)
	// is not a problem for the call to [io.Readall].
	defer pF.Close()
	if e != nil {
		// We could check for file non-existence here.
		// And we could panic if it happens, altho a race
		// for a just-deleted file is also conceivable.
		return &fs.PathError{Op:"os.Open",Err:e,Path:shortFP}
	}
	var bb []byte
	bb, e = io.ReadAll(pF)
	if e != nil {
		return &fs.PathError{Op:"io.ReadAll",Err:e,Path:shortFP}
	}	
	// NOTE: 2023.03 Trimming leading whitespace and ensuring
	// that there is a trailing newline are probably unnecessary
	// AND unhelpful - they violate the Principle of Least Surprise.
	// They might also conflict with digital signings. 
	// pPI.Raw = S.TrimSpace(pPI.TypedRaw.S() + "\n")
	// pPI.size = len(pPI.Raw)

	// This is not supposed to happen,
	// cos we checked for Size()==0 at entry
	if len(bb) == 0 {
		panic("==> empty file?!: " + shortFP)
	}
	// println("LoadContents: Allocating!")
	p.TypedRaw = new(CT.TypedRaw)
	p.Raw = CT.Raw(string(bb))
	// Take the hash and set the field.
	// p.Hash = *new([16]byte)
        p.Hash = md5.Sum(bb)

	// TODO try to set CT.RawMT?
	
	return nil
}

/*
func FileInfoString(p fs.FileInfo) string {
     if p == nil { return "<FI:NIL>" }
     return fmt.Sprintf("%s<%s>%d", p.Name(), p.FSItem_type, p.Size())
}
*/

