package fileutils

import (
	"fmt"
	"errors"
	"os"
	FP "path/filepath"
)

// BasicPath is a filepath we have redd, will read, or will create.
// It might also be a directory or symlink, either of which requires
// further processing elsewhere. In the most common usage, it is a file.
// It can be nil, if e.g. its content was created on-the-fly.
type BasicPath struct {
	error
	// RelFilePath is a "short" argument passed in at creation time, e.g.
	// a filename specified on the command line, and relative to the CWD.
	RelFilePath string
	AbsFilePath
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

// GetError is necessary cos "Error()"" dusnt tell you whether "error"
// is "nil", which is the indication of no error. Therefore we need
// this function, which can actually return the telltale "nil".
func (p *BasicPath) GetError() error {
	return p.error
}

// Error satisfies interface "error", but the
// weird thing is that "error" can be nil.
func (p *BasicPath) Error() string {
	if p.error != nil {
		return p.error.Error()
	}
	return ""
}

func (p *BasicPath) SetError(e error) {
	p.error = e
}

// IsOkayFile is a convenience function.
func (p *BasicPath) IsOkayFile() bool {
	return p.error == nil && p.Exists && p.isFile && !p.isDir
}

// IsOkayDir is a convenience function.
func (p *BasicPath) IsOkayDir() bool {
	return p.error == nil && p.Exists && !p.isFile && p.isDir
}

// IsOkaySymlink is a convenience function.
func (p *BasicPath) IsOkaySymlink() bool {
	return p.error == nil && p.Exists && !p.isFile && !p.isDir && p.isSymL
}

// NewBasicPath requires a non-nil "RelFilePath" and analyzes it.
// It returns a pointer that can be used in a "CheckedPath" to
// start a method chain.
func NewBasicPath(rfp string) *BasicPath {
	rp := new(BasicPath)
	rp.RelFilePath = rfp
	rp.AbsFilePath = AbsFP(rfp)
	rp.AbsFilePathParts = rp.AbsFilePath.NewAbsPathParts()
	return rp.setFlags()
}

// setFlags requires a non-nil "AbsFilePath" and checks for existence and type.
func (p *BasicPath) setFlags() *BasicPath {
	if p.error != nil {
		return p // or nil ?
	}
	if p.AbsFilePath == "" {
		p.error = errors.New("fu.BasicPath.check: Nil filepath")
		return p // nil
	}
	var FI os.FileInfo
	FI, e := os.Lstat(p.AbsFilePath.S())
	if e != nil {
		p.error = errors.New("fu.BasicPath.check: Lstat failed: " + p.AbsFilePath.S())
		// The file or directory does not exist. DON'T PANIC.
		// Just return before any flags are set, such as Exists.
		return p // nil
	}
	p.isDir = FI.IsDir()
	p.isFile = FI.Mode().IsRegular()
	p.isSymL = (0 != (FI.Mode() & os.ModeSymlink))
	p.Exists = p.isDir || p.isFile || p.isSymL
	if p.isFile {
		p.Size = int(FI.Size())
	}
	return p
}

// ResolveSymlinks will follow links until it finds something else. 
func (p *BasicPath) ResolveSymlinks() bool {
	if p.error != nil {
		return false
	}
	if !p.IsOkaySymlink() {
		return false
	}
	var newPath string
	var wasResolved = false
	var e error
	for p.IsOkaySymlink() {
		// func os.Readlink(pathname string) (string, error)
		// func FP.EvalSymlinks(path string) (string, error)
		newPath, e = FP.EvalSymlinks(p.AbsFilePath.S())
		if e != nil {
			p.error = fmt.Errorf("fu.RslvSymLx <%s>: %w", p.AbsFilePath, e)
			return wasResolved
		}
		println("--> Symlink from:", p.AbsFilePath)
		println("     resolved to:", newPath)
		p.AbsFilePath = AbsFilePath(newPath)
		p.setFlags()
		wasResolved = true
	}
	return wasResolved
}
