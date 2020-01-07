package fileutils

import (
	"errors"
	"os"
)

// BasicPath is a filepath we have redd, will read, or will create.
// It might also be a directory or symlink, either of which requires
// further processing elsewhere. In the most common usage, it is a file.
// It can be nil, e.g. if the content was created on-the-fly.
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
	IsDir  bool
	IsFile bool
	IsSymL bool
	// Size is here and not in `struct CheckedPath` because its
	// value is already made available to us when `func check()`
	// calls `os.FileInfo os.Lstat(..)`, below.
	Size int
}

// GetError is necessary cos `Error()`` dusnt tell you whether `error` is `nil`,
// which is the indication of no error. Therefore we need this function, which
// can actually return the telltale `nil`.`
func (p *BasicPath) GetError() error {
	return p.error
}

// Error satisfied interface `error`, but the
// weird thing is that `error` can be nil.
func (p *BasicPath) Error() string {
	if p.error != nil {
		return p.error.Error()
	}
	return ""
}

func (p *BasicPath) SetError(e error) {
	p.error = e
}

// TODO: IsOkayFile(), IsOkayDir(), IsOkaySymlink()

func (p *BasicPath) PathType() string {
	if p.AbsFilePath == "" {
		panic("fu.BasicPath.PathType: AFP not initialized")
	}
	if p.error != nil || !p.Exists {
		return "NIL"
	}
	if p.IsDir && !p.IsFile {
		return "DIR"
	}
	if p.IsFile && !p.IsDir {
		return "FILE"
	}
	panic("fu.BasicPath.PathType: bad state (symlink?)")
}

// NewBasicPath requires a non-nil `RelFilePath` and analyzes it.
// It returns a pointer that can be used in a CheckedPath to
// start a method chain.
func NewBasicPath(rfp string) *BasicPath {
	rp := new(BasicPath)
	rp.RelFilePath = rfp
	rp.AbsFilePath = AbsFP(rfp)
	rp.AbsFilePathParts = rp.AbsFilePath.NewAbsPathParts()
	return rp.check()
}

// check requires a non-nil `AbsFilePath` and checks for existence and type.
func (p *BasicPath) check() *BasicPath {
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
	p.IsDir = FI.IsDir()
	p.IsFile = FI.Mode().IsRegular()
	p.IsSymL = (0 != (FI.Mode() & os.ModeSymlink))
	p.Exists = p.IsDir || p.IsFile || p.IsSymL
	if p.IsFile {
		p.Size = int(FI.Size())
	}
	return p
}
