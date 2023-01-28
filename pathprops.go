package fileutils

import (
	"errors"
	"fmt"
	L "github.com/fbaube/mlog"
	SU "github.com/fbaube/stringutils"
	"io"
	"os"
	FP "path/filepath"
	S "strings"
)

// MAX_FILE_SIZE is set (arbitrarily) to 2 megabytes
const MAX_FILE_SIZE = 2000000

// PathProps describes a filepath we have redd, will read, or will create.
// It might also be a directory or symlink, either of which requires
// further processing elsewhere. In the most common usage, it is a file.
// It can be nil, if e.g. its content was created on-the-fly.
//
// PathProps is embedded in ContentityRecord. (FIXME)
//
// Note that RelFP and AbsFP must be exported to be persisted to the DB.
// .
type PathProps struct { // this has Raw
	Raw     string
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
		s = "Not-OK (PathProps uninitialized?)"
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

// getContentBytes reads in the file (IFF it is a file) into the field [Raw].
// It is tolerant about non-files and empty files, returning nil for error.
// The func "os.Open(fp)" defaults to R/W, altho R/O would probably suffice.
// .
func (pPI *PathProps) getContentBytes() error {
	if pPI.Raw != "" {
		L.L.Warning("pp.GetContentBytes: overwriting[%d]", len(pPI.Raw))
		pPI.Raw = ""
	}
	pPI.Raw = ""
	var shortAbsFP string
	shortAbsFP = SU.ElideHomeDir(pPI.AbsFP.S())
	if !pPI.IsOkayFile() {
		s := "Not a file: " + shortAbsFP
		L.L.Warning(s)
		return nil
	}
	// Zero-length ?
	if pPI.size == 0 {
		L.L.Warning("Zero-length file: " + shortAbsFP)
		return nil
	}
	// Suspiciously tiny ?
	if pPI.size < 6 {
		L.L.Warning("pp.GetContentBytes: tiny file [%d]: "+
			shortAbsFP, pPI.size)
	}
	// If it's too big, BARF!
	if pPI.size > MAX_FILE_SIZE {
		// L.L.Error("pp.GetContentBytes: file too large [%d]: %s",
		//	pPI.size, shortAbsFP)
		return fmt.Errorf("pp.getContentBytes: "+
			"file too large [%d]: %s", pPI.size, shortAbsFP)
	}
	// Open it (and then immediately close it), just to check.
	var pF *os.File
	var e error
	pF, e = os.Open(pPI.AbsFP.S())
	defer pF.Close()
	if e != nil {
		// L.L.Error("fu.BP.GetContentBytes.osOpen<%s>: %s",
		//       shortAbsFP, e.Error())
		return fmt.Errorf("ppgetContentBytes.osOpen<%s>: %w",
			shortAbsFP, e)
	}
	var bb []byte
	bb, e = io.ReadAll(pF)
	if e != nil {
		// L.L.Error("pp.getContentBytes.ioReadAll<%s>: %s",
		//      shortAbsFP, e.Error())
		return fmt.Errorf("pp.getContentBytes.ioReadAll<%s>: %w",
			shortAbsFP, e)
	}
	if len(bb) == 0 {
		panic("==> empty file?!: " + shortAbsFP)
	}
	pPI.Raw = string(bb)
	return nil
}

// FetchRaw reads in the file (IFF it is a file) and trims away
// leading and trailing whitespace, but then adds a final newline
// (which might be kinda dumb if it's a binary file).
// .
func (pPI *PathProps) FetchRaw() error {
	if pPI.Size() == 0 {
		L.L.Progress("fetchRaw: Skipping for zero-length content")
		return nil
	}
	DispFP := pPI.AbsFP.Tildotted()
	if !pPI.IsOkayFile() {
		errors.New("fetchRaw: not a readable file: " + DispFP)
	}
	var e error
	e = pPI.getContentBytes()
	//  pPI.HasError() {
	if e != nil {
		return fmt.Errorf("fetchRaw: "+
			"PI.GetContentBytes<%s> failed: %w", DispFP, e)
	}
	if len(pPI.Raw) == 0 {
		return fmt.Errorf(
			"fetchRaw: PI.GetContentBytes<%s> got zilch", DispFP)
	}
	pPI.Raw = S.TrimSpace(pPI.Raw) + "\n"
	pPI.size = len(pPI.Raw)
	return nil
}
