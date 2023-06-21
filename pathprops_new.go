package fileutils

import (
	SU "github.com/fbaube/stringutils"
	// "os"
	"fmt"
	FP "path/filepath"
)

func NewContentedPathProps(fp string) (*PathProps, error) {
	var e error
	var pp *PathProps
	pp, e = NewPathProps(fp)
	if e != nil {
		return nil, fmt.Errorf(
			"pp.NewContentedPathProps.newPP<%s>: %w", fp, e)
	}
	e = pp.GoGetFileContents()
	if e != nil {
		return nil, fmt.Errorf(
			"pp.NewContentedPathProps.goGetFC<%s>: %w", fp, e)
	}
	return pp, nil
}

// NewPathProps requires an absolute or relative filepath,
// and analyzes the object (if one exists) at the path.
// This func does not load and analyse the content.
//
// Note that a relative path is appended to the CWD,
// which may not be the desired behavior; in such a
// case, use NewPathPropsRelativeTo (below).
//
// (OBS) This func had been changed from a ptr return to
// value return cos of a bug (I assumed) in Go 1.18 generics.
// Also, a value return could be considered a flag that this
// struct is mostly read-onlny after creation.
// .
func NewPathProps(fp string) (*PathProps, error) {
	var e error
	var pp *PathProps
	pp = new(PathProps)
	pp.RelFP = fp
	afp, e := FP.Abs(fp)
	pp.AbsFP = AbsFilePath(afp)
	pp.ShortFP = SU.Tildotted(afp)
	if e != nil {
		return pp, WrapAsPathPropsError(
			e, "FP.Abs(..) (fu.PPnew.L32)", nil)
	}
	pp.FileMeta = *NewFileMeta(afp)
	return pp, pp.GetError()
}

// NewPathPropsRelativeTo requires a relative filepath plus an absolute
// filepath being referenced. This func does not load & analyse the content.
func NewPathPropsRelativeTo(rfp, relTo string) (*PathProps, error) {
	pp := new(PathProps)
	if !FP.IsAbs(relTo) {
		return pp, NewPathPropsError(
			"not an abs FP", "FP.IsAbs(..) (fu.PPnew.L76)", nil)
		// panic("newPPrelTo: not an abs.FP: " + relTo)
	}
	// pp.RelFP = rfp // this looks pretty dodgy
	afp := FP.Join(relTo, rfp)
	return NewPathProps(afp)
}
