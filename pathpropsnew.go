package fileutils

import (
	"os"
	FP "path/filepath"
)

// NewPathProps requires an absolute or relative filepath, and analyzes 
// the object (if one exists) at the path. This func does not load and 
// analyse the content.
// 
// Note that a relative path is appended to the CWD, which may not be the
// desired behavior; in such a case, use NewPathPropsRelativeTo (below).
// 
// This func has been changed from a ptr return to value return because 
// of a bug in Go 1.18 generics. Also, a value return can be considered 
// a flag that this struct is mostly read-onlny after creation.
// 
func NewPathProps(fp string) (*PathProps, error) {
	var e error
	pp := new(PathProps)
	afp, e := FP.Abs(fp)
	if e != nil {
		return nil, WrapAsPathPropsError(
			e, "FP.Abs(..) (fu.PPnew.L20)", nil)
	}
	pp.AbsFP = AbsFilePath(afp)
	var FI os.FileInfo
	FI, e = os.Lstat(afp)
	if e != nil {
		// fmt.Println("fu.newPP: os.Lstat<%s> failed: %w", afp, e)
		if os.IsNotExist(e) {
			// File or directory does not exist
			return nil, WrapAsPathPropsError(
				e, "does not exist (fu.PPnew.L30)", pp) 
			}
		if os.IsExist(e) {
			pp.exists = true 
			return nil, WrapAsPathPropsError(
				e, "exists but !os.Lstat(..) (fu.PPnew.L35)", pp) 
			}
		panic("exists+not in fu.newPP.L34")
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
	var e error
	pp := new(PathProps)
	if !FP.IsAbs(relTo) {
		return nil, NewPathPropsError(
			"not an abs FP", "FP.IsAbs(..) (fu.PPnew.L50)", nil) 
		// panic("newPPrelTo: not an abs.FP: " + relTo)
	}
	pp.RelFP = rfp
	afp := FP.Join(relTo, rfp)
	pp.AbsFP = AbsFP(afp)
	var FI os.FileInfo
	FI, e = os.Lstat(afp)
	if e != nil {
		// fmt.Println("fu.newPP: os.Lstat<%s> failed: %w", Tildotted(afp), e)
		// The file or directory does not exist. DON'T PANIC.
		// Just return before any flags are set, such as Exists.
		return pp, WrapAsPathPropsError(
			e, "os.Lstat(..) (fu.PPnewrelto.L62)", pp) 
	}
	pp.isDir = FI.IsDir()
	pp.isFile = FI.Mode().IsRegular()
	pp.isSymL = (0 != (FI.Mode() & os.ModeSymlink))
	pp.exists = pp.isDir || pp.isFile || pp.isSymL
	if pp.isFile {
		pp.size = int(FI.Size())
	}
	// println("==>", SU.Gbg(" "+pi.String()+" "))
	return pp, nil 
}
