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
	relFP   string
	absFP   AbsFilePath
	exists  bool
	isDir   bool
	isFile  bool
	isSymL  bool
	// size is here and not elsewhere because
	// its value is already available to us when
	// "os.FileInfo os.Lstat(..)" is called, below.
	size int
}

func (pi *PathInfo) String() (s string) {
	if pi.IsOkayFile()    { s = "OK-File " } else
	if pi.IsOkayDir()     { s = "OK-Dirr " } else
	if pi.IsOkaySymlink() { s = "OK-SymL " } else
	                      { s = "Not-OK " }
	if pi.bpError != nil  { s+= "ERROR " }
	s += fmt.Sprintf("[%d] ", pi.size)
	s += pi.absFP.Enhomed()
	return s
}

// Echo implements Markupper.
func (p PathInfo) Echo() string {
	return p.AbsFP()
}

func (p *PathInfo) Size() int {
	return p.size
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

// Exists is a convenience function.
func (p *PathInfo) Exists() bool {
	return p.exists
}

// IsOkayFile is a convenience function.
func (p *PathInfo) IsOkayFile() bool {
	return p.bpError == nil && p.exists && p.isFile && !p.isDir
}

// IsOkayDir is a convenience function.
func (p *PathInfo) IsOkayDir() bool {
	return p.bpError == nil && p.exists && !p.isFile && p.isDir
}

// IsOkaySymlink is a convenience function.
func (p *PathInfo) IsOkaySymlink() bool {
	return p.bpError == nil && p.exists && !p.isFile && !p.isDir && p.isSymL
}

// NewPathInfo requires a non-nil "RelFilePath" and analyzes it.
// It returns a pointer that can be used in a "CheckedPath" to
// start a method chain.
func NewPathInfo(rfp string) *PathInfo {
	pi := new(PathInfo)
	pi.absFP = AbsFP(rfp)
	absFPstr := string(pi.absFP)
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
	pi.exists = pi.isDir || pi.isFile || pi.isSymL
	if pi.isFile {
		 pi.size = int(FI.Size())
	}
	return pi
}

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
