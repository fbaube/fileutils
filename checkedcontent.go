package fileutils

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	S "strings"
	FP "path/filepath"
)

// MAX_FILE_SIZE is set (arbitrarily) to 2 megabytes
const MAX_FILE_SIZE = 2000000

// CheckedContent is a file we have redd, will read, or will create.
// It might also be a directory or symlink, either of which requires
// further processing elsewhere.
//
// CheckedContent comprises three sub-structs:
//  - "PathInfo" is a pointer if the content is created on-the-fly in-memory.
//  - "BasicAnalysis" is where the results of content analysis are placed.
//  - Each of these has its own error field, any one of which can bubble
//    up to the "error" field of CheckedContent.
//
type CheckedContent struct {
	RelFilePath string
  PathInfo
	Raw string
	BasicAnalysis
	error
}

// NewCheckedContent works for directories and symlinks too.
func NewCheckedContent(pPI *PathInfo) *CheckedContent {
	var e error
	pCC := new(CheckedContent)
	pCC.PathInfo = *pPI
	if pPI.IsOkayDir() || pPI.IsOkaySymlink() {
		return pCC
	}
	if !pPI.IsOkayFile() {
		pCC.error = errors.New("Is not valid file, directory, or symlink")
		return pCC
	}
	// OK, it's a file.
	pCC.Raw, e = pPI.FetchContent()
	if e != nil {
		pCC.error = errors.New("Could not fetch content")
		return pCC
	}
	pCC.BasicAnalysis.FileIsOkay = true
	pBA, e := AnalyseFile(pCC.Raw, FP.Ext(string(pPI.absFP)))
	if e != nil {
		panic(e)
	}
	pCC.BasicAnalysis = *pBA
	return pCC
}

func NewCheckedContentFromPath(path string) *CheckedContent {
	bp := NewPathInfo(path)
  return NewCheckedContent(bp)
}

// String implements Markupper.
func (p CheckedContent) String() string {
	if p.IsOkayDir() {
		return fmt.Sprintf("PathInfo: DIR[%d]: %s | %s",
			p.size, p.RelFilePath, p.AbsFP()) // FilePathParts.Echo())
	}
	var isXML string
	if p.IsXML() {
		isXML = "[XML] "
	}
	s := fmt.Sprintf("ChP: %sLen<%d>MType<%s>",
		isXML, p.size, p.MType)
	s += fmt.Sprintf("\n\t %s | %s", p.RelFilePath, Enhomed(p.AbsFP()))
	s += fmt.Sprintf("\n\t (snift) %s ", p.MimeType)
	return s
}

// GetError is necessary cos "Error()"" dusnt tell you whether "error"
// is "nil", which is the indication of no error. Therefore we need
// this function, which can actually return the telltale "nil".
func (p *CheckedContent) GetError() error {
	return p.bpError
}

// Error satisfied interface "error", but the
// weird thing is that "error" can be nil.
func (p *CheckedContent) Error() string {
	if p.bpError != nil {
		return p.bpError.Error()
	}
	return ""
}

// SetError sets "error" to the error.
func (p *CheckedContent) SetError(e error) {
	p.bpError = e
}

// FetchContent reads in the file (IFF it is a file) and trims away
// leading and trailing whitespace, but then adds a final newline.
func (pPI *PathInfo) FetchContent() (raw string, e error) {
	DispFP := pPI.absFP.Tildotted()
	if !pPI.IsOkayFile() {
		return "", errors.New("fu.fetchcontent: not a readable file: " + DispFP)
	}
	var bb []byte
	bb = pPI.GetContentBytes()
	if pPI.bpError != nil {
		 return "", fmt.Errorf("fu.fetchcontent: PI.GetContentBytes<%s> failed: %w",
				DispFP, pPI.bpError)
	}
	raw = S.TrimSpace(string(bb)) + "\n"
	return raw, nil
}

// GetContentBytes reads in the file (IFF it is a file).
// If an error, it is returned in "BasicPath.error",
// and the return value is "nil".
// The func "os.Open(fp)" defaults to R/W, altho R/O
// would probably suffice.
func (pPI *PathInfo) GetContentBytes() []byte {
	if pPI.bpError != nil {
		return nil
	}
	TheAbsFP := pPI.absFP.Tildotted()
	if !pPI.IsOkayFile() {
		pPI.bpError = errors.New("fu.BP.GetContentBytes: not a file: " + TheAbsFP)
		return nil
	}
	if pPI.size == 0 {
		println("==> zero-length file:", TheAbsFP)
		return make([]byte, 0)
	}
	// If it's too big, BARF!
	if pPI.size > MAX_FILE_SIZE {
		 pPI.bpError = fmt.Errorf(
			"fu.BP.GetContentBytes: file too large (%d): %s", pPI.size, TheAbsFP)
		return nil
	}
	// Open it (and then immediately close it), just to check.
	var pF *os.File
	var e error
	pF, e = os.Open(TheAbsFP)
	defer pF.Close()
	if e != nil {
		pPI.bpError = errors.New(fmt.Sprintf(
				"fu.BP.GetContentBytes.osOpen<%s>: %w", TheAbsFP, e))
		return nil
	}
	var bb []byte
	bb, e = ioutil.ReadAll(pF)
	if e != nil {
		pPI.bpError = errors.New(fmt.Sprintf(
				"fu.BP.GetContentBytes.ioutilReadAll<%s>: %w", TheAbsFP, e))
	}
	if len(bb) == 0 {
		println("==> empty file?!:", TheAbsFP)
	}
	return bb
}
