package fileutils

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	S "strings"
)

// CheckedPath is a filepath we have redd, will read, or will create.
// It might also be a directory or symlink, either of which requires
// further processing elsewhere.
// In normal usage, if it is a file, it will be opened and loaded into
// `Raw` (by `func FileLoad()`), and at that point it will be fully
// decoupled from the file system.
// Mime type guessing is done using standard libraries (both Go's
// and a 3rd party's), so this code is still very low-level.
type CheckedPath struct {
	// error if non-nil can abort execution.
	error
	// ArgFilePath should not be necessary.
	// // ArgFilePath string
	// RelFilePath is a "short" argument passed in at creation time, e.g.
	// a filename specified on the command line, and relative to the CWD.
	RelFilePath string
	AbsFilePath
	// We require that AbsFilePathParts.Echo() == AbsFilePath
	AbsFilePathParts
	Exists bool
	IsDir  bool
	IsFile bool
	IsSymL bool
	// FileCt is >1 IFF this struct refers to a directory, and multifile
	// processing is needed. In the future it might also handle wildcards.
	FileCt int
	// Raw and Size apply to files only, not directories or symlinks.
	Raw  string
	Size int
	// MagicMimeType is set using a 3rd party binding to libmagic.
	MagicMimeType string
	// SniftMimeType is set using the Golang stdlib.
	SniftMimeType string
	// Mtype is set by our own code, based on MagicMimeType, SniftMimeType,
	// and shallow analysis of the file contents.
	Mtype []string
	// IsXML is set by our own code, using various heuristics of our own devising.
	IsXML bool
}

func (p *CheckedPath) Error() string {
	if p.error == nil {
		return ""
	}
	return p.error.Error()
}

func (p *CheckedPath) Type() string {
	if p.AbsFilePath == "" {
		panic("fu.CheckedPath.Type: AFP not initialized")
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
	panic("fu.CheckedPath.Type: bad state")
}

// NewCheckedPath requires a non-nil `RelFilePath` and analyzes it.
// It returns a pointer that can be used in a method chain.
func NewCheckedPath(rfp string) *CheckedPath {
	rp := &CheckedPath{RelFilePath: rfp}
	rp.AbsFilePath = AbsFP(rfp)
	rp.AbsFilePathParts = rp.AbsFilePath.GetAbsPathParts()
	return rp.check()
}

// check requires a non-nil `AbsFilePath` and checks for existence and type.
func (p *CheckedPath) check() *CheckedPath {
	if p.error != nil {
		return p // nil
	}
	if p.AbsFilePath == "" {
		p.error = errors.New("fu.CheckedPath.check: Nil filepath")
		return p // nil
	}
	var FI os.FileInfo
	FI, e := os.Lstat(p.AbsFilePath.S())
	if e != nil {
		p.error = errors.New("fu.CheckedPath.check: Lstat failed: " + p.AbsFilePath.S())
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

// LoadFile reads in the file (IFF it is a file)
// and does a quick check for the MIME type.
// If it does not exist, be nice: do nothing and return no error.
// If it is not a file, be nice: do nothing and return no error.
// If `Raw` is not "", be nice: the file is an on-the-fly temp
// file, so skip the load and just do the quick MIME analysis.
func (p *CheckedPath) LoadFile() *CheckedPath {
	/*
		if p.error != nil {
			return p // nil
		}
		if p.AbsFilePath == "" {
			p.error = errors.New("fu.GGFile.FileLoad: Nil filepath")
			return p // nil
		} */
	if p.Type() != "FILE" {
		return p
	}
	if p.Raw != "" {
		p.error = errors.New("fu.CheckedPath.LoadFile: Already loaded")
		return p
	}
	if p.Size == 0 {
		println("==> zero-length file:", p.AbsFilePath.S())
		return nil
	}
	// If it's too big, BARF!
	if p.Size > 2000000 {
		p.error = fmt.Errorf(
			"fu.CheckedPath.LoadFile: file too large (%d): %s",
			p.Size, p.AbsFilePath)
	}
	// Open it (and then immediately close it), just to check.
	var pF *os.File
	var e error
	pF, e = os.Open(p.AbsFilePath.S())
	pF.Close()
	if e != nil {
		p.error = // errors.Wrapf(e, "fu.GGFile.FileLoadAndAnalyze.osOpen<%s>", p.AbsFilePath.S())
			errors.New(fmt.Sprintf(
				"fu.CheckedPath.LoadFile.osOpen<%s>: %s", p.AbsFilePath.S(), e.Error()))
		return p
	}
	// Read it in !
	// TODO/FIXME Use github.com/dimchansky/utfbom Skip()
	var bb []byte
	bb, e = ioutil.ReadFile(p.AbsFilePath.S())
	if e != nil {
		p.error = // errors.Wrapf(e, "fu.OpenAndLoadContent.ioutilReadFile<%s>", fullpath)
			errors.New(fmt.Sprintf(
				"fu.CheckedPath.LoadFile.ioutilReadFile<%s>: %s", p.AbsFilePath.S(), e.Error()))
	}
	if len(bb) == 0 {
		println("==> empty file:", p.AbsFilePath.S())
	}
	p.Raw = S.TrimSpace(string(bb))

	if !S.HasPrefix(p.AbsFilePathParts.FileExt, ".") {
		println("==> (oops had to add a dot to filext")
		p.AbsFilePathParts.FileExt = "." + p.AbsFilePathParts.FileExt
	}
	return p
}

// InspectFile comprises four steps:
//
// * use stdlib and third-party libraries to make initial guesses
// * dump those guesses for the purpose of evaluating those libraries
// * call custom code to evaluate more deeply XML and/or as mixed content
// * dump those results for the purpose of refining the code
//
// The fields of interest in `struct fileutiles.InputFile`:
//
// - Set using various heuristics of our own devising: IsXML bool
// - Set using Golang stdlib: SniftMimeType string
// - Set using 3rd-party lib: MagicMimeType string
// - Set by our own code, based on the results set
// in the preceding string fields: Mtype []string
//
func (p *CheckedPath) InspectFile() *CheckedPath {
	if p.error != nil || p.Type() != "FILE" {
		return p
	}
	p.MagicMimeType = GoMagic(p.Raw)
	// Trim long JPEG descriptions
	if s := p.MagicMimeType; S.HasPrefix(s, "JPEG") {
		if i := S.Index(s, "xres"); i > 0 {
			p.MagicMimeType = "JPEG, " + s[i:]
		}
	}
	// This next call assigns "text/html" to DITA maps :-/
	contyp := http.DetectContentType([]byte(p.Raw)) // (content))
	p.SniftMimeType = S.TrimSuffix(contyp, "; charset=utf-8")

	// fmt.Printf("    MIME: (%s) %s \n", p.SniftMimeType, p.MagicMimeType)
	return p
}

// SegmentedFile describes in detail a file we have redd or will read.
// (If field `FileFullName` is nil, it has been created on-the-fly.)
// In normal usage, the file is opened and its contents are redd
// into `Contents`, and then it is decoupled from the file system.
//
// Because our goal is to process LwDITA, we examine a text file
// (and its DTDs, if present) and set a type of XDITA, HDITA,
// MDITA, or DITA. This amounts to making an assertion, and can
// be rolled back (i.e. the bool can be set back to `false`) if
// further processing of the file shows that the file does not
// in fact even try to conform to the previously+incorrectly
// asserted file type.
//
// NOTE A text-based image file (i.e. SVG or EPS) can be
// `IsImage` but `!IsBinary`.
//
/*
type SegmentedFile struct {
	CheckedPath
	// FileContent holds the file's entire contents (i.e. `GGFile.Raw`)
	// MINUS any content analysed and read into `Header`.
	FileContent
	// Header basically holds all the in-file metadata, no matter what
	// the format of the file. For MDITA (Markdown-XP) this is YAML.
	// For XDITA and HDITA ([X]HTML) this is head/meta elements.
	// It can be non-nil but have its sero value, which indicates
	// that metadata was checked for but none was found.
	// Storing metadata this way makes it easier to manage it in a
	// consistent and format-independent manner, and makes it easier
	// to add to it and modify it at runtime.
	// TODO: When in-file metadata is stored as JSON K/V pairs,
	// it can be accessed from the command line using Sqlite tools.
	*Header
}
*/
