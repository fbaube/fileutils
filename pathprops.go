package fileutils

import (
	"errors"
	"fmt"
	"io"
	"os"
	FP "path/filepath"
	S "strings"

	L "github.com/fbaube/mlog"
)

// MAX_FILE_SIZE is set (arbitrarily) to 2 megabytes
const MAX_FILE_SIZE = 2000000

// PathProps describes a filepath we have redd, will read, or will create.
// It might also be a directory or symlink, either of which requires
// further processing elsewhere. In the most common usage, it is a file.
// It can be nil, if e.g. its content was created on-the-fly.
//
// Note that RelFP and AbsFP must be exported to be persisted to the DB.
//
type PathProps struct {
	error  error
	RelFP  string
	AbsFP  AbsFilePath
	exists bool
	isDir  bool
	isFile bool
	isSymL bool
	// size is here and not elsewhere in some other struct
	// because its value is already available to us when
	// "os.FileInfo os.Lstat(..)" is called, below.
	size int
}

func (pi *PathProps) String() (s string) {
	if pi.IsOkayFile() {
		s = fmt.Sprintf("OK-File (len:%d) ", pi.size)
	} else if pi.IsOkayDir() {
		s = "OK-Dirr "
	} else if pi.IsOkaySymlink() {
		s = "OK-SymL "
	} else {
		s = "Not-OK "
	}
	if pi.HasError() {
		s += "hasERROR  "
	}
	s += pi.AbsFP.Tildotted()
	return s
}

// Echo implements Markupper.
func (p PathProps) Echo() string {
	return p.AbsFP.S()
}

func (p *PathProps) Size() int {
	return p.size
}

/*
func (p *PathProps) AbsFP() string {
	return string(p.absFP)
}
func (p *PathProps) RelFP() string {
	return string(p.relFP)
}
*/

// Exists is a convenience function.
func (p *PathProps) Exists() bool {
	return p.exists
}

// IsOkayFile is a convenience function.
func (p *PathProps) IsOkayFile() bool {
	return (!p.HasError()) && p.exists && p.isFile && !p.isDir
}

// IsOkayDir is a convenience function.
func (p *PathProps) IsOkayDir() bool {
	return (!p.HasError()) && p.exists && !p.isFile && p.isDir
}

// IsOkaySymlink is a convenience function.
func (p *PathProps) IsOkaySymlink() bool {
	return (!p.HasError()) && p.exists && !p.isFile && !p.isDir && p.isSymL
}

// IsOkayWhat is for use with functions from github.com/samber/lo 
func (p *PathProps) IsOkayWhat() string {
	if p.IsOkayFile() { return "FILE" }
	if p.IsOkayDir() { return "DIR" }
	if p.IsOkaySymlink() { return "SYMLINK" }
	return "n/a" 
}

// ResolveSymlinks will follow links until it finds something else.
func (pp *PathProps) ResolveSymlinks() *PathProps {
	if pp.HasError() {
		return nil
	}
	if !pp.IsOkaySymlink() {
		return nil
	}
	var newPath string
	var e error
	for pp.IsOkaySymlink() {
		// func os.Readlink(pathname string) (string, error)
		// func FP.EvalSymlinks(path string) (string, error)
		newPath, e = FP.EvalSymlinks(pp.AbsFP.S())
		if e != nil {
			pp.SetError(fmt.Errorf("fu.RslvSymLx <%s>: %w", pp.AbsFP, e))
			return nil
		}
		println("--> Symlink from:", pp.AbsFP)
		println("     resolved to:", newPath)
		pp.AbsFP = AbsFilePath(newPath)
		var e error 
		pp, e = NewPathProps(newPath)
		if e != nil {
			panic(e)
			return nil 
		}
		// CHECK IT 
	}
	return pp
}

// GetContentBytes reads in the file (IFF it is a file).
// If an error, it is returned in "BasicPath.error",
// and the return value is "nil".
// The func "os.Open(fp)" defaults to R/W, altho R/O
// would probably suffice.
func (pPI *PathProps) GetContentBytes() []byte {
	if pPI.HasError() {
		return nil
	}
	TheAbsFP := pPI.AbsFP.Tildotted()
	if !pPI.IsOkayFile() {
		pPI.SetError(errors.New("fu.BP.GetContentBytes: not a file: " + TheAbsFP))
		return nil
	}
	// Zero-length ?
	if pPI.size == 0 {
		L.L.Warning("Zero-length file: " + TheAbsFP)
		return make([]byte, 0)
	}
	// Suspiciously tiny ?
	if pPI.size < 6 {
		L.L.Warning("Too-tiny file, ignoring: " + TheAbsFP)
		return make([]byte, 0)
	}
	// If it's too big, BARF!
	if pPI.size > MAX_FILE_SIZE {
		pPI.SetError(fmt.Errorf(
			"fu.BP.GetContentBytes: file too large (%d): %s", pPI.size, TheAbsFP))
		return nil
	}
	// Open it (and then immediately close it), just to check.
	var pF *os.File
	var e error
	pF, e = os.Open(TheAbsFP)
	defer pF.Close()
	if e != nil {
		pPI.SetError(errors.New(fmt.Sprintf(
			"fu.BP.GetContentBytes.osOpen<%s>: %w", TheAbsFP, e)))
		return nil
	}
	var bb []byte
	bb, e = io.ReadAll(pF)
	if e != nil {
		pPI.SetError(errors.New(fmt.Sprintf(
			"fu.BP.GetContentBytes.ioReadAll<%s>: %w", TheAbsFP, e)))
	}
	if len(bb) == 0 {
		println("==> empty file?!:", TheAbsFP)
	}
	return bb
}

// FetchContent reads in the file (IFF it is a file) and trims away
// leading and trailing whitespace, but then adds a final newline.
func (pPI *PathProps) FetchContent() (raw string, e error) {
	DispFP := pPI.AbsFP.Tildotted()
	if !pPI.IsOkayFile() {
		return "", errors.New("fu.fetchcontent: not a readable file: " + DispFP)
	}
	var bb []byte
	bb = pPI.GetContentBytes()
	if pPI.HasError() {
		return "", fmt.Errorf("fu.fetchcontent: PI.GetContentBytes<%s> failed: %w",
			DispFP, pPI.GetError())
	}
	raw = S.TrimSpace(string(bb)) + "\n"
	return raw, nil
}

// === Implement interface Errable

func (p *PathProps) HasError() bool {
	return p.error != nil && p.error.Error() != ""
}

// GetError is necessary cos "Error()"" dusnt tell you whether "error"
// is "nil", which is the indication of no error. Therefore we need
// this function, which can actually return the telltale "nil".
func (p *PathProps) GetError() error {
	return p.error
}

// Error satisfies interface "error", but the
// weird thing is that "error" can be nil.
func (p *PathProps) Error() string {
	if p.error != nil {
		return p.error.Error()
	}
	return ""
}

func (p *PathProps) SetError(e error) {
	p.error = e
}
