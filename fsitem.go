package fileutils

import (
	"fmt"
	"errors"
	CT "github.com/fbaube/ctoken"
	L "github.com/fbaube/mlog"
	// SU "github.com/fbaube/stringutils"
	"io"
	"io/fs"
	"os"
	FP "path/filepath"
)

// MAX_FILE_SIZE is set (arbitrarily) to 4 megabytes
const MAX_FILE_SIZE = 4000000

// FSItem is an item identified by a filepath (plus its contents) that
// we have redd, will read, or will create. It might be a directory or
// symlink, either of which requires further processing elsewhere. In
// the most common usage, it is a file. Its filepath(s) can be empty
// ("") if (for example) its content was created interactively. 
//
// NOTE that the file name (the part of the full path after the last
// directory separator) is not stored separately: it is stored in the
// AbsFP *and* the RelFP. Note also that this path & name information
// duplicates what is stored in an instance of orderednodes.Nord . 
//
// NOTE that the embedded field [FileMeta] embeds an [os.FileInfo].
// 
// FSItem is embedded in struct [datarepo/rowmodels/ContentityRow].
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
// .
type FSItem struct { // this has (Typed) Raw
	FileMeta
	CT.TypedRaw
	FPs Filepaths 
}

func (p *FSItem) IsDirlike() bool {
     if p.IsFile() { return false }
     if p.IsDir()  { return true  }
     return p.FileMeta.IsDirlike()
}

// IsWhat is for use with functions from github.com/samber/lo .
// If the item does not exists, it returns "".
func (p *FSItem) IsWhat() string {
	if p.IsFile() {
		return "FILE"
	}
	if p.IsDir() {
		return "DIR"
	}
	if p.IsSymlink() {
		return "SYMLINK"
	}
	if p.Exists() {
		return "UnknownType"
	}
	return "" // "Non-existent"
}

// ResolveSymlinks will follow links until it finds
// something else. NOTE that this is a SECURITY HOLE. 
func (p *FSItem) ResolveSymlinks() *FSItem {
	if !p.IsSymlink() {
		return nil
	}
	var newPath string
	var e error
	for p.IsSymlink() {
		// func os.Readlink(pathname string) (string, error)
		// func FP.EvalSymlinks(path string) (string, error)
		newPath, e = FP.EvalSymlinks(p.FPs.AbsFP.S())
		if e != nil {
			L.L.Error("fu.RslvSymLx <%s>: %s", p.FPs.AbsFP, e.Error())
			// p.SetError(fmt.Errorf("fu.RslvSymLx <%s>: %w", p.FPs.AbsFP, e))
			return nil
		}
		println("--> Symlink from:", p.FPs.AbsFP)
		println("     resolved to:", newPath)
		p.FPs.AbsFP = AbsFilePath(newPath)
		var e error
		p, e = NewFSItem(newPath)
		if e != nil {
			panic(e)
			return nil
		}
		// CHECK IT
	}
	return p
}

// GoGetFileContents reads in the file (assuming it is a file)
// into the field [Raw] and does a quick check for XML and HTML5
// declarations.
//
// It assumes that [LStat] has been called, and that the size
// of the file is known. Therefore this func is a no-op if func
// [BasicInfo.Size] returns 0, its zero value. Therefore do not
// call this if the argument's [BasicInfo] is uninitialized.
//
// It is tolerant about non-files and empty files,
// returning nil for error.
//
// The call it makes to [os.Open] defaults to R/W mode,
// altho R/O would probably suffice.
// .
func (p *FSItem) GoGetFileContents() error { // *fs.PathError {
	if p.Size() == 0 {
		// No-op
		return nil
	}
	if !p.IsFile() {
		// No-op
		return nil
	}
	var shortFP = p.FPs.ShortFP
	if p.Raw != "" {
		// No-op with warning
		L.L.Warning("pp.GoGetFileContents: already "+
			"loaded [%d]: %s", len(p.Raw), shortFP)
		return nil
	}
	// Suspiciously tiny ?
	if p.Size() < 6 {
		L.L.Warning("pp.GoGetFileContents: tiny "+
			"file [%d]: %s", p.Size(), shortFP)
	}
	// If it's too big, BARF!
	if p.Size() > MAX_FILE_SIZE {
		return &fs.PathError{Op:"FSI.GoGetFileContents",
		       Err:errors.New(fmt.Sprintf(
		       "file too large: %d", p.Size())),Path:shortFP}
	}
	// Open it, just to check (and then immediately close it)
	var pF *os.File
	var e error
	pF, e = os.Open(p.FPs.AbsFP.S())
	// Note that this defer'd Close() (i.e. file is left open)
	// is not a problem for the call to io.Readall].
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
	// pPI.Raw = S.TrimSpace(pPI.TypedRaw.S() + "\n")
	// pPI.size = len(pPI.Raw)

	// This is not supposed to happen,
	// cos we checked for Size()==0 at entry
	if len(bb) == 0 {
		panic("==> empty file?!: " + shortFP)
	}
	p.Raw = CT.Raw(string(bb))
	return nil
}
