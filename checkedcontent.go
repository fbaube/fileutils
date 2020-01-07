package fileutils

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	S "strings"
)

// CheckedContent is a filepath we have redd, will read, or will create.
// It might also be a directory or symlink, either of which requires
// further processing elsewhere.
// In normal usage, if it is a file, it will be opened and loaded into
// `Raw` (by `func FileLoad()`), and at that point the content will be
// fully decoupled from the file system.
// Mime type guessing is done using standard libraries (both Go's and
// a 3rd party's), so this code can still be considered very low-level.
type CheckedContent struct {
	*BasicPath
	BasicContent
	error
}

func NewCheckedContent(pBP *BasicPath) *CheckedContent {
	pCC := pBP.ReadContent()
	if pBP.error != nil {
		return pCC
	}
	pCC.InspectFile()
	return pCC
}

func NewCheckedContentFromPath(path string) *CheckedContent {
	p := new(CheckedContent)
	p.BasicPath = NewBasicPath(path)
	// LOAD ?
	return p
}

// GetError is necessary cos `Error()`` dusnt tell you whether `error` is `nil`,
// which is the indication of no error. Therefore we need this function, which
// can actually return the telltale `nil`.`
func (p *CheckedContent) GetError() error {
	return p.error
}

// Error satisfied interface `error`, but the
// weird thing is that `error` can be nil.
func (p *CheckedContent) Error() string {
	if p.error != nil {
		return p.error.Error()
	}
	return ""
}

func (p *CheckedContent) SetError(e error) {
	p.error = e
}

// ReadContent reads in the file (IFF it is a file)
// and does a quick check for the MIME type.
// If it does not exist, be nice: do nothing and return no error.
// If it is not a file, be nice: do nothing and return no error.
// If `Raw` is not "", be nice: the file is already loaded and
// is quite possibly an on-the-fly temp file, so skip the load
// and just do the quick MIME analysis.
func (pBP *BasicPath) ReadContent() *CheckedContent {
	pCC := new(CheckedContent)
	pCC.BasicPath = pBP
	TheAbsFP := NiceFP(pBP.AbsFilePath.S())
	if !pBP.IsOkayFile() { // pBP.PathType() != "FILE" {
		pCC.error = errors.New("fu.ReadContent: not a file: " + TheAbsFP)
		return pCC
	}
	pCC.BasicContent = *new(BasicContent)
	if pBP.Size == 0 {
		println("==> zero-length file:", TheAbsFP)
		return nil
	}
	// If it's too big, BARF!
	if pBP.Size > 2000000 {
		pBP.error = fmt.Errorf(
			"fu.ReadContent: file too large (%d): %s", pBP.Size, TheAbsFP)
	}
	// Open it (and then immediately close it), just to check.
	var pF *os.File
	var e error
	pF, e = os.Open(TheAbsFP)
	defer pF.Close()
	if e != nil {
		pBP.error = errors.New(fmt.Sprintf(
				"fu.ReadContent.osOpen<%s>: %s", TheAbsFP, e.Error()))
		return pCC
	}
	var bb []byte
	// Read it in !
	// NEW CODE
	// func ReadAll(r io.Reader) ([]byte, error)
	bb, e = ioutil.ReadAll(pF)
	/* OLD CODE
	// TODO/FIXME Use github.com/dimchansky/utfbom Skip()
	bb, e = ioutil.ReadFile(p.AbsFilePath.S())
	*/
	if e != nil {
		pBP.error = errors.New(fmt.Sprintf(
				"fu.ReadContent.ioutilReadAll<%s>: %w", TheAbsFP, e))
	}
	if len(bb) == 0 {
		println("==> empty file:", TheAbsFP)
	}
	pCC.Raw = S.TrimSpace(string(bb))

	if !S.HasPrefix(pBP.AbsFilePathParts.FileExt, ".") {
		println("==> (oops had to add a dot to filext")
		pBP.AbsFilePathParts.FileExt = "." + pBP.AbsFilePathParts.FileExt
	}
	if S.Contains(pCC.Raw, "<!DOCTYPE HTML ") {
		println("FOUND HTML")
	}
	return pCC
}

// FileType returns `XML`, `MKDN`, or future stuff TBD.
func (p *CheckedContent) FileType() string {
	if p.MType == nil {
		println("Unallocated MType[]!")
		return "ERR/OH/CRAP"
	}
	return S.ToUpper(p.MType[0])
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
func (p *CheckedContent) InspectFile() *CheckedContent {
	if p.error != nil || !p.IsOkayFile() { // p.PathType() != "FILE" {
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

	println("cc.InspectFile: Did own; now SetFileMtype")
	p.SetFileMtype()

	// fmt.Printf("    MIME: (%s) %s \n", p.SniftMimeType, p.MagicMimeType)
	return p
}
