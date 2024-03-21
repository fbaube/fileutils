package fileutils

import (
	// SU "github.com/fbaube/stringutils"
	"io/fs"
	"os"
	"fmt"
	"errors"
	FP "path/filepath"
	L "github.com/fbaube/mlog"
)

func NewFSItemWithContent(fp string) (*FSItem, *fs.PathError) {
	var e error // *fs.PathError
	var pfsi *FSItem
	pfsi, e = NewFSItem(fp)
	if e != nil {
		return nil, &fs.PathError{Op:"NewFSItem",Err:e,Path:fp}
	}
	e = pfsi.GoGetFileContents()
	if e != nil {
		return nil, &fs.PathError{
		       Op:"FSI.GoGetFileContents", Err:e,Path:fp}
	}
	return pfsi, nil
}

// NewFSItem takes a filepath (absolute or relative) and
// analyzes the object (assuming one exists) at the path.
// This func does not load and analyse the content.
//
// Note that a relative path is appended to the CWD,
// which may not be the desired behavior; in such a
// case, use NewFSItemRelativeTo (below).
// .
func NewFSItem(fp string) (*FSItem, *fs.PathError) {
     	if fp == "" {
	   println("NewFSItem GOT NIL PATH")
	   return nil, &os.PathError{Op:"NewFSItem",
	   	  Err:errors.New("Empty path arg"),Path:""}
	   }
	var pfsi *FSItem
	pfsi = new(FSItem)
	pfps, e := NewFilepaths(fp)
	if e != nil {
	     	pfsi.SetError(e)
		return nil, &fs.PathError{Op:"FSI.NewFPs",Err:e,Path:fp}
	}
	// L.L.Dbg("NewFilepaths: %#v", *pfps)
	pfsi.FPs = *pfps
	pfsi.FileMeta = *NewFileMeta(pfps.AbsFP.S())
	var fmError error 
	if fmError = pfsi.GetError(); fmError == nil {
	   return pfsi, nil
	   }
	L.L.Info("fmError %T %#v", fmError, fmError)
	var q *os.PathError
	var ok bool
	q, ok = fmError.(*fs.PathError)
	if !ok {
	   q = &os.PathError{Op:"NewFileMeta",Err:fmError,Path:fp}
	   }
	return pfsi, q
}

// NewFSItemRelativeTo takes a relative filepath
// plus an absolute filepath being referenced. 
// This func does not load & analyse the content.
func NewFSItemRelativeTo(rfp, relTo string) (*FSItem, *fs.PathError) {
	if !FP.IsAbs(relTo) {
		return nil, &fs.PathError{Op:"fp.isRelTo.notAbs",
		Err:errors.New("relFP must be rel to an absFP"),
		Path:fmt.Sprintf("relFP<%s>.relTo.nonAbsFP<%s>",rfp,relTo)}
	}
	afp := FP.Join(relTo, rfp)
	return NewFSItem(afp)
}
