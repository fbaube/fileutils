package fileutils

import (
	"fmt"
	"errors"
	"io"
	"os"
	"io/fs"
	"crypto/md5"
	CT "github.com/fbaube/ctoken"
	SU "github.com/fbaube/stringutils"
	L "github.com/fbaube/mlog"
)

// MAX_FILE_SIZE is set (arbitrarily) to 100 megabytes.
const MAX_FILE_SIZE = 100_000_000
// MIN_FILE_SIZE is the minimum that can be analysed for
// MIME/content type, and is set (arbitrarily) to 6 bytes.
const MIN_FILE_SIZE = 6

func init() {
     var fso *FSObject
     fso = &(FSObject {})
     var de fs.DirEntry  // ifc
     var sr SU.Stringser // ifc
     // var okde, oksr bool 
     de /* ,okde */ = fs.DirEntry(fso)
     sr /* ,oksr */ = SU.Stringser(fso)
     // if ! (okde && oksr) { panic("FSObject ifc's") }
     fmt.Printf("DirrEntry: %v \n", de)
     fmt.Printf("Stringser: %v \n", sr)
}

// func (p *FSObject) Echo()  string { return "ECHO" }
// func (p *FSObject) Infos() string { return "INFOS" }
// func (p *FSObject) Debug() string { return "DEBUG" }

// FSObject is an item identified by a filepath (plus its contents) 
// that we have tried to or will try to read, write, or create. It 
// might be a directory or symlink, either of which requires further
// processing elsewhere. In the most common usage, it is a file.
//
// CONTAINS RAW, in *CT.TypedRaw 
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
// NOTE that it embeds an [fs.FileInfo], and implements interfaces [FSObjecter],
// [fs.FileInfo], and [fs.DirEntry]), and contains basic file system metadata
// PLUS the path to the item (whicih FIleInfo does not contain) AND the item
// contents (but only after lazy loading). The `FileInfo` is the results of
// a call to [os.LStat]/[fs.Lstat] (or perhaps alternatively the contents
// of a record in sqlar or zip), parsed.
// 
// FSObject is embedded in struct [datarepo/rowmodels/ContentityRow].
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
type FSObject struct { // this has (Typed) Raw
     	// LastCheckTime is TBS.
	// LastCheckTime time.Time

	// TypedRaw is an outlier, because it is the only 
	// field in this struct that cannot be deduced from 
	// the FileInfo. Therefore it is a candidate to be
	// removed. It is a ptr, to allow for lazy loading.
	*CT.TypedRaw
	
     	// FileInfo implements interfaces [os.FileInfo] and [fs.DirEntry].
	// It should probably be an unexported, lower case "fi", because
	// its integrity is relied on heavily. We use it as a struct and 
	// not as a ptr-to-struct, so that (a) it is not shared-writable,
	// and (b) there is no chance of a NPE. Each path includes the
	// [FP.Base].
	fs.FileInfo
	// FSO_type is closely linked to FileInfo and 
	// they should always be updated in lockstep.
	FSO_type
	// FPs has abs & rel paths, and an indicator of which was used to
	// instantiate it. We use it as a struct and not as a ptr-to-struct,
	// so that it is not shared-writable, and so that there is no chance
	// of a NPE.
	// 
	// Paths follow our rules:
	//  - a directory MUST end in a slash (or OS sep)
	//  - a symlink MUST NOT end in a slash (or OS sep)
	// 
	// Note that an [fs.FileInfo] does not preserve or provide path 
	// info, which is part of the motivstion for this too-large struct. 
	FPs Filepaths

	// Perms is UNIX-style "rwx" user/group/world
	Perms string	
	// Inode and NLinks are for hard link detection. 
	Inode, NLinks int // uint64
	// Errer provides an NPE-proof error field
	Errer
}

// Contents returns a file's contents. It first 
// lazy-loads the file into field [TypedRaw] IFF
//  - it IS a file, and 
//  - it has not been read in yet
// and then it
//  - calculates & stores file file's hash, and 
//  - quickly checks for XML and HTML5 declarations
//
// Contents should always be fresh, even when files
// are active, so we first fetch a new [os.FileInfo]
// to check for changes. If no changes are indicated,
// and the content has not been changed programmatically
// (check flag [ContentIsDirty]), do a fast return.
// 
// It is tolerant about non-files, non-existent 
// objects, and empty files, returning nil error.
//
// NOTE The call to [os.Open] defaults to R/W mode,
// even tho R/O might often suffice.
// .
func (p *FSObject) Contents() (string, error) {
	// Exists ?
	if p.FPs.DoesNotExist {
	   return "", fmt.Errorf("fso.contents(%s): %w", os.ErrNotExist) 
	}
	// Not a file ? 
	if !p.IsFile() {
	   	// p.TypedRaw.Raw_type =  SU.Raw_type_DIRLIKE
		// if p.IsDir() { p.TypedRaw.Raw_type = SU.Raw_type_DIR }
		return "", nil
	}
     	var e error
	var hasChanged bool 
	var newFI os.FileInfo
	var shortFP = p.FPs.ShortFP

	// Get a fresh FileInfo
	newFI, e = os.Lstat(p.FPs.AbsFP)
	if e != nil {
	   return "", fmt.Errorf("fso.contents(%s): %w", p.FPs.AbsFP, e)
	}
	// The content has previously been fetched
	// but the object has changed somehow ? 
	if p.FPs.ContentInMemoryIsDirty ||
	  (p.FileInfo.ModTime() != newFI.ModTime() &&
	  !p.FileInfo.ModTime().IsZero()) {
	   p.FileInfo = newFI
	   hasChanged = true	   
	}
	// If the object hasn't changed and we
	// already have the contents, return now
	if p.TypedRaw != nil && !hasChanged {
		return p.TypedRaw.S(), nil
	}
	// Allocate this now to prevent NPEs
	p.TypedRaw = new(CT.TypedRaw)
	
	if p.Size() == 0 {
	   // No-op
	   // This might be repetitive
	   p.TypedRaw.Raw_type = SU.Raw_type_NIL
	   return "", nil
	} else if // Suspiciously tiny ?
	        p.Size() < MIN_FILE_SIZE { 
		L.L.Warning("fso.contents: tiny "+
			"file [%d]: %s", p.Size(), shortFP)
		p.TypedRaw.Raw_type = SU.Raw_type_NIL
	}
	// println("LoadContents: chkpt 2")
	// If it's too big, BARF!
	if p.Size() > MAX_FILE_SIZE {
		return "[TOO BIG]", &fs.PathError {
		        Op:"fso.contents", Err:errors.New(fmt.Sprintf(
		       "file too large: %d", p.Size())), Path:shortFP}
	}
	// println("LoadContents: chkpt 3")
	// Open it, just to check (and then immediately close it)
	var pF *os.File
	// NOTE FIXME This might fail in a RootFS
	pF, e = os.Open(p.FPs.AbsFP)
	defer pF.Close()
	if e != nil {
		// We could check for file non-existence here.
		// And we could panic if it happens, altho a race
		// for a just-deleted file is also conceivable.
		return "", &fs.PathError{Op:"fso.content:os.open",
		       Err:e, Path:shortFP }
	}
	// -------------------
	//  NOW READ THE FILE
	// -------------------
	var bb []byte
	bb, e = io.ReadAll(pF)
	if e != nil {
		return "", &fs.PathError{
		       Op:"fso.contents:io.readall", Err:e,Path:shortFP }
	}	
	// NOTE: 2023.03 Trimming leading whitespace and ensuring
	// that there is a trailing newline are probably unnecessary
	// AND unhelpful - they violate the Principle of Least Surprise
	// - and they might also conflict with digital signings. 
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

	// TODO: Try to set CT.RawMT?
	
	return p.TypedRaw.S(), nil
}

