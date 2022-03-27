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
	error   error
	RelFP   string
	AbsFP   AbsFilePath
	ShortFP string // expressed if possible using "~" or "."
	exists  bool
	isDir   bool
	isFile  bool
	isSymL  bool
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

// Exists is a convenience function.
func (p *PathProps) Exists() bool {
	return p.exists
}

// IsOkayFile is a convenience function.
func (p *PathProps) IsOkayFile() bool {
	return p.exists && p.isFile && !p.isDir
}

// IsOkayDir is a convenience function.
func (p *PathProps) IsOkayDir() bool {
	return p.exists && !p.isFile && p.isDir
}

// IsOkaySymlink is a convenience function.
func (p *PathProps) IsOkaySymlink() bool {
	return p.exists && !p.isFile && !p.isDir && p.isSymL
}

// IsOkayWhat is for use with functions from github.com/samber/lo
func (p *PathProps) IsOkayWhat() string {
	if p.IsOkayFile() {
		return "FILE"
	}
	if p.IsOkayDir() {
		return "DIR"
	}
	if p.IsOkaySymlink() {
		return "SYMLINK"
	}
	if p.Exists() {
		return "UnknownType"
	}
	return "Non-existent"
}

// ResolveSymlinks will follow links until it finds something else.
func (pp *PathProps) ResolveSymlinks() *PathProps {
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
			L.L.Error("fu.RslvSymLx <%s>: %s", pp.AbsFP, e.Error())
			// pp.SetError(fmt.Errorf("fu.RslvSymLx <%s>: %w", pp.AbsFP, e))
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
	TheAbsFP := pPI.AbsFP.Tildotted()
	if !pPI.IsOkayFile() {
		L.L.Error("fu.BP.GetContentBytes: not a file: " + TheAbsFP)
		// pPI.SetError(errors.New("fu.BP.GetContentBytes: not a file: " + TheAbsFP))
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
		// pPI.SetError(fmt.Errorf(
		// 	"fu.BP.GetContentBytes: file too large (%d): %s", pPI.size, TheAbsFP))
		L.L.Error("fu.BP.GetContentBytes: file too large (%d): %s", pPI.size, TheAbsFP)
		return nil
	}
	// Open it (and then immediately close it), just to check.
	var pF *os.File
	var e error
	pF, e = os.Open(TheAbsFP)
	defer pF.Close()
	if e != nil {
		// pPI.SetError(errors.New(fmt.Sprintf(
		// 	"fu.BP.GetContentBytes.osOpen<%s>: %w", TheAbsFP, e)))
		L.L.Error("fu.BP.GetContentBytes.osOpen<%s>: %s", TheAbsFP, e.Error())
		return nil
	}
	var bb []byte
	bb, e = io.ReadAll(pF)
	if e != nil {
		// pPI.SetError(errors.New(fmt.Sprintf(
		// 	"fu.BP.GetContentBytes.ioReadAll<%s>: %w", TheAbsFP, e)))
		L.L.Error("fu.BP.GetContentBytes.ioReadAll<%s>: %s", TheAbsFP, e.Error())
	}
	if len(bb) == 0 {
		println("==> empty file?!:", TheAbsFP)
	}
	return bb
}

// FetchContent reads in the file (IFF it is a file) and trims away
// leading and trailing whitespace, but then adds a final newline.
func (pPI *PathProps) FetchContent() (raw string, e error) {
	if pPI.Size() == 0 {
		L.L.Progress("FetchContent: Skipping for zero-length content")
		return "", nil
	}
	DispFP := pPI.AbsFP.Tildotted()
	if !pPI.IsOkayFile() {
		return "", errors.New("fu.fetchcontent: not a readable file: " + DispFP)
	}
	var bb []byte
	bb = pPI.GetContentBytes()
	//  pPI.HasError() {
	if bb == nil || len(bb) == 0 {
		return "", fmt.Errorf(
			"fu.fetchcontent: PI.GetContentBytes<%s> failed", DispFP)
	}
	raw = S.TrimSpace(string(bb)) + "\n"
	return raw, nil
}
