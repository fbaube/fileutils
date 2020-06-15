package fileutils

import (
	"fmt"
	"errors"
	"os"
	FP "path/filepath"
)

// PathInfo describes a filepath we have redd, will read, or will create.
// It might also be a directory or symlink, either of which requires
// further processing elsewhere. In the most common usage, it is a file.
// It can be nil, if e.g. its content was created on-the-fly.
type PathInfo struct {
	bpError error
	absFP   AbsFilePath
	// AbsFilePathParts stores the pieces of the absolute filepath.
	// We require that AbsFilePathParts.Echo() == AbsFilePath
	*AbsFilePathParts
	Exists bool
	isDir  bool
	isFile bool
	isSymL bool
	// Size is here and not in "struct CheckedPath" because its
	// value is already made available to us when "func check()""
	// calls "os.FileInfo os.Lstat(..)"", below.
	Size int
}

func (p *PathInfo) AbsFP() string {
	return string(p.absFP) 
}

// GetError is necessary cos "Error()"" dusnt tell you whether "error"
// is "nil", which is the indication of no error. Therefore we need
// this function, which can actually return the telltale "nil".
func (p *PathInfo) GetError() error {
	return p.bpError
}

// Error satisfies interface "error", but the
// weird thing is that "error" can be nil.
func (p *PathInfo) Error() string {
	if p.bpError != nil {
		return p.bpError.Error()
	}
	return ""
}

func (p *PathInfo) SetError(e error) {
	p.bpError = e
}

// IsOkayFile is a convenience function.
func (p *PathInfo) IsOkayFile() bool {
	return p.bpError == nil && p.Exists && p.isFile && !p.isDir
}

// IsOkayDir is a convenience function.
func (p *PathInfo) IsOkayDir() bool {
	return p.bpError == nil && p.Exists && !p.isFile && p.isDir
}

// IsOkaySymlink is a convenience function.
func (p *PathInfo) IsOkaySymlink() bool {
	return p.bpError == nil && p.Exists && !p.isFile && !p.isDir && p.isSymL
}

// NewPaths requires a non-nil "RelFilePath" and analyzes it.
// It returns a pointer that can be used in a "CheckedPath" to
// start a method chain.
func NewPaths(rfp string) *Paths {
	p := new(Paths)
	p.RelFilePath = rfp
	p.AbsFilePath = AbsFP(rfp)
	// p.AbsFilePathParts = p.AbsFilePath.NewAbsPathParts()
	return p
}

// NewPathInfo requires a non-nil "RelFilePath" and analyzes it.
// It returns a pointer that can be used in a "CheckedPath" to
// start a method chain.
func NewPathInfo(rfp string) *PathInfo {
	pi := new(PathInfo)
	pi.absFP = AbsFP(rfp)
	absFPstr := string(pi.absFP)
	pi.AbsFilePathParts = pi.absFP.NewAbsPathParts()
	var FI os.FileInfo
	FI, e := os.Lstat(absFPstr)
	if e != nil {
		pi.bpError = errors.New("fu.BasicPath.check: Lstat failed: " + absFPstr)
		// The file or directory does not exist. DON'T PANIC.
		// Just return before any flags are set, such as Exists.
		return pi
	}
	pi.isDir  = FI.IsDir()
	pi.isFile = FI.Mode().IsRegular()
	pi.isSymL = (0 != (FI.Mode() & os.ModeSymlink))
	pi.Exists = pi.isDir || pi.isFile || pi.isSymL
	if pi.isFile {
		pi.Size = int(FI.Size())
	}
	return pi
}

func (p *Paths) Filext() string {
	fx1 := FP.Ext(p.RelFilePath)
	fx2 := FP.Ext(string(p.AbsFilePath))
	// fx3 := p.AbsFilePathParts.FileExt
	if fx1 != fx2 { // || fx2 != fx3 {
		println("ERR: fu.bp.filext:", fx1, fx2) // , fx3)
	}
	return fx2
}

/*
// setFlags requires a non-nil "AbsFilePath" and checks for existence and type.
func (p *Paths) setFlags() *PathInfo {
	pi := new(PathInfo)
	if p.AbsFilePath == "" {
		pi.bpError = errors.New("fu.setFlags: Nil filepath")
		return pi
	}
	pi.absFP = p.AbsFilePath
	var FI os.FileInfo
	FI, e := os.Lstat(p.AbsFilePath.S())
	if e != nil {
		pi.bpError = errors.New("fu.BasicPath.check: Lstat failed: " + p.AbsFilePath.S())
		// The file or directory does not exist. DON'T PANIC.
		// Just return before any flags are set, such as Exists.
		return pi
	}
	pi.isDir  = FI.IsDir()
	pi.isFile = FI.Mode().IsRegular()
	pi.isSymL = (0 != (FI.Mode() & os.ModeSymlink))
	pi.Exists = pi.isDir || pi.isFile || pi.isSymL
	if pi.isFile {
		pi.Size = int(FI.Size())
	}
	return pi
}
*/

// ResolveSymlinks will follow links until it finds something else.
func (pi *PathInfo) ResolveSymlinks() *PathInfo {
	if pi.bpError != nil {
		return nil
	}
	if !pi.IsOkaySymlink() {
		return nil
	}
	var newPath string
	var e error
	for pi.IsOkaySymlink() {
		// func os.Readlink(pathname string) (string, error)
		// func FP.EvalSymlinks(path string) (string, error)
		newPath, e = FP.EvalSymlinks(pi.absFP.S())
		if e != nil {
			pi.bpError = fmt.Errorf("fu.RslvSymLx <%s>: %w", pi.absFP, e)
			return nil
		}
		println("--> Symlink from:", pi.absFP)
		println("     resolved to:", newPath)
		pi.absFP = AbsFilePath(newPath)
		pi = NewPathInfo(newPath)
	}
	return pi
}
