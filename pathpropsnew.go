package fileutils

import (
	"os"
	FP "path/filepath"
)

// NewPathProps requires an absolute or relative filpath, and analyzes it.
// Note that a relative path is appended to the CWD, which may not be correct.
// In such a case, use NewPathPropsRelativeTo (below).
// This func does not load & analyse the content.
func NewPathProps(fp string) (*PathProps, error) {
	var e error
	pp := new(PathProps)
	afp, e := FP.Abs(fp)
	if e != nil {
		// (ermsg string, op string, pp *PathProps, srcLoc string) 
		return nil, NewPathPropsError(
			"FP.Abs(..) failed", "Abs(..)", nil, "fu.PPnew.L20")
	}
	pp.AbsFP = AbsFilePath(afp)
	var FI os.FileInfo
	FI, e = os.Lstat(afp)
	if e != nil {
		// fmt.Println("fu.newPP: os.Lstat<%s> failed: %w", afp, e)
		// The file or directory does not exist. DON'T PANIC.
		// Just return before any flags are set, such as Exists.
		return nil, NewPathPropsError(
			"os.Lstat(..) failed", "Lstat(..)", pp, "fu.PPnew.L30")
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
		panic("newPPrelTo: not an abs.FP: " + relTo)
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
		return pp, NewPathPropsError(
			"os.Lstat(..) failed", "Lstat(..)", pp, "fu.PPnewrelto.L62")
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
