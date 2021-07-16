package fileutils

import (
	"fmt"
	"os"
	FP "path/filepath"
)

// NewPathProps requires an absolute or relative filpath, and analyzes it.
// Note that a relative path is appended to the CWD, which may not be correct.
// In such a case, use NewPatPhropsRelativeTo (below).
// This func does not load & analyse the content.
func NewPathProps(fp string) *PathProps {
	var e error
	pi := new(PathProps)
	afp, e := FP.Abs(fp)
	if e != nil {
		panic("newPP: " + e.Error())
	}
	pi.AbsFP = AbsFilePath(afp)
	var FI os.FileInfo
	FI, e = os.Lstat(afp)
	if e != nil {
		// fmt.Println("fu.newPP: os.Lstat<%s> failed: %w", Tildotted(afp), e)
		pi.SetError(fmt.Errorf("fu.newPP: os.Lstat<%s> failed: %w", Tildotted(afp), e))
		// The file or directory does not exist. DON'T PANIC.
		// Just return before any flags are set, such as Exists.
		return pi
	}
	pi.isDir = FI.IsDir()
	pi.isFile = FI.Mode().IsRegular()
	pi.isSymL = (0 != (FI.Mode() & os.ModeSymlink))
	pi.exists = pi.isDir || pi.isFile || pi.isSymL
	if pi.isFile {
		pi.size = int(FI.Size())
	}
	// println("==> new fu.pathprops:", pi.String())
	return pi
}

// NewPathPropsRelativeTo requires a relative filepath plus an absolute filepath.
// This func does not load & analyse the content.
func NewPathPropsRelativeTo(rfp, relTo string) *PathProps {
	var e error
	pi := new(PathProps)
	if !FP.IsAbs(relTo) {
		panic("newPPrelTo: not an abs.FP: " + relTo)
	}
	pi.RelFP = rfp
	afp := FP.Join(relTo, rfp)
	pi.AbsFP = AbsFP(afp)
	var FI os.FileInfo
	FI, e = os.Lstat(afp)
	if e != nil {
		// fmt.Println("fu.newPP: os.Lstat<%s> failed: %w", Tildotted(afp), e)
		pi.SetError(fmt.Errorf("fu.newPP: os.Lstat<%s> failed: %w", Tildotted(afp), e))
		// The file or directory does not exist. DON'T PANIC.
		// Just return before any flags are set, such as Exists.
		return pi
	}
	pi.isDir = FI.IsDir()
	pi.isFile = FI.Mode().IsRegular()
	pi.isSymL = (0 != (FI.Mode() & os.ModeSymlink))
	pi.exists = pi.isDir || pi.isFile || pi.isSymL
	if pi.isFile {
		pi.size = int(FI.Size())
	}
	// println("==>", SU.Gbg(" "+pi.String()+" "))
	return pi
}
