package fileutils

import (
	SU "github.com/fbaube/stringutils"
	// "os"
	"fmt"
	FP "path/filepath"
)

func NewContentedFSItem(fp string) (*FSItem, error) {
	var e error
	var pfsi *FSItem
	pfsi, e = NewFSItem(fp)
	if e != nil {
		return nil, fmt.Errorf(
			"pfsi.NewContentedFSItem.newPFSI<%s>: %w", fp, e)
	}
	e = pfsi.GoGetFileContents()
	if e != nil {
		return nil, fmt.Errorf(
			"pfsi.NewContentedFSItem.goGetFC<%s>: %w", fp, e)
	}
	return pfsi, nil
}

// NewFSItem requires an absolute or relative filepath,
// and analyzes the object (if one exists) at the path.
// This func does not load and analyse the content.
//
// Note that a relative path is appended to the CWD,
// which may not be the desired behavior; in such a
// case, use NewFSItemRelativeTo (below).
//
// (OBS) This func had been changed from a ptr return to
// value return cos of a bug (I assumed) in Go 1.18 generics.
// Also, a value return could be considered a flag that this
// struct is mostly read-onlny after creation.
// .
func NewFSItem(fp string) (*FSItem, error) {
	var e error
	var pfsi *FSItem
	pfsi = new(FSItem)
	pfsi.RelFP = fp
	afp, e := FP.Abs(fp)
	pfsi.AbsFP = AbsFilePath(afp)
	pfsi.ShortFP = SU.Tildotted(afp)
	if e != nil {
		return pfsi, WrapAsFSItemError(
			e, "FP.Abs(..) (fu.PFSInew.L32)", nil)
	}
	pfsi.FileMeta = *NewFileMeta(afp)
	return pfsi, pfsi.GetError()
}

// NewFSItemRelativeTo requires a relative filepath plus an absolute
// filepath being referenced. This func does not load & analyse the content.
func NewFSItemRelativeTo(rfp, relTo string) (*FSItem, error) {
	pfsi := new(FSItem)
	if !FP.IsAbs(relTo) {
		return pfsi, NewFSItemError(
			"not an abs FP", "FP.IsAbs(..) (fu.PFSInew.L76)", nil)
		// panic("newPFSIrelTo: not an abs.FP: " + relTo)
	}
	// pfsi.RelFP = rfp // this looks pretty dodgy
	afp := FP.Join(relTo, rfp)
	return NewFSItem(afp)
}
