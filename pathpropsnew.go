package fileutils

import (
	SU "github.com/fbaube/stringutils"
	"os"
	FP "path/filepath"
)

// NewPathProps requires an absolute or relative filepath,
// and analyzes the object (if one exists) at the path.
// This func does not load and analyse the content.
//
// Note that a relative path is appended to the CWD,
// which may not be the desired behavior; in such a
// case, use NewPathPropsRelativeTo (below).
//
// This func has been changed from a ptr return to value
// return because of a bug (I assume) in Go 1.18 generics.
// Also, a value return can be considered a flag that this
// struct is mostly read-onlny after creation.
func NewPathProps(fp string) (*PathProps, error) {
	var e error
	pp := new(PathProps)
	pp.RelFP = fp
	afp, e := FP.Abs(fp)
	pp.AbsFP = AbsFilePath(afp)
	pp.ShortFP = SU.Tildotted(afp)
	if e != nil {
		return pp, WrapAsPathPropsError(
			e, "FP.Abs(..) (fu.PPnew.L30)", nil)
	}
	var FI os.FileInfo
	FI, e = os.Lstat(afp)
	if e != nil {
		// fmt.Println("fu.newPP: os.Lstat<%s> failed: %w", afp, e)
		if os.IsNotExist(e) {
			// File or directory does not exist
			return pp, WrapAsPathPropsError(
				e, "does not exist", pp)
		}
		if os.IsExist(e) {
			pp.exists = true
			return nil, WrapAsPathPropsError(
				e, "exists but !os.Lstat(..) (fu.PPnew.L44)", pp)
		}
		panic("exists+not in fu.newPP.L46")
	}
	pp.isDir = FI.IsDir()
	pp.isFile = FI.Mode().IsRegular()
	pp.isSymL = (0 != (FI.Mode() & os.ModeSymlink))
	pp.exists = pp.isDir || pp.isFile || pp.isSymL
	if pp.isFile {
		pp.size = int(FI.Size())
	}
	// println("==> new fu.pathprops:", pi.String())
	return pp, nil
}

// NewPathPropsRelativeTo requires a relative filepath plus an absolute
// filepath being referenced. This func does not load & analyse the content.
func NewPathPropsRelativeTo(rfp, relTo string) (*PathProps, error) {
	pp := new(PathProps)
	if !FP.IsAbs(relTo) {
		return pp, NewPathPropsError(
			"not an abs FP", "FP.IsAbs(..) (fu.PPnew.L66)", nil)
		// panic("newPPrelTo: not an abs.FP: " + relTo)
	}
	// pp.RelFP = rfp // this looks pretty dodgy
	afp := FP.Join(relTo, rfp)
	return NewPathProps(afp)
}
