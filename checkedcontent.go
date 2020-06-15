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
//  - "BasicPath" is a pointer if the content is created on-the-fly in-memory.
//  - "BasicContent" is basically just the content - or the file count if
//     the BasicPath is a directory.
//  - "BasicAnalysis" is where the results of content analysis are placed.
//  - Each of these has its own error field, any one of which can bubble
//    up to the "error" field of CheckedContent.
//
type CheckedContent struct {
	Paths
  PathInfo
	BasicContent // Content_raw
	BasicAnalysis
	error
}

func NewCheckedContent(pPI *PathInfo) *CheckedContent {
	pCC := new(CheckedContent)
	pCC.PathInfo = *pPI
	pBC := pPI.FetchContent()
	if pPI.bpError != nil {
		return pCC
	}
	pCC.BasicContent = *pBC
	pCC.BasicAnalysis.FileIsOkay = pPI.IsOkayFile()
	if pCC.BasicAnalysis.FileIsOkay && pCC.bpError == nil &&
		 pCC.bcError == nil && pCC.error == nil {
 		pBA, e := AnalyseFile(pBC.Raw, FP.Ext(string(pPI.absFP))) // pPI.Filext())
		if e != nil {
			panic(e)
		}
		pCC.BasicAnalysis = *pBA
	} else {
		println("fu.cc.newcc: Could not newcc()")
	}
	return pCC
}

func NewCheckedContentFromPath(path string) *CheckedContent {
	bp := NewPathInfo(path)
  return NewCheckedContent(bp)
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

// FetchContent reads in the file (IFF it is a file) and does
// a quick check of the MIME type before returning the promoted
// type, "CheckedContent".
//
//  * If it does not exist, be nice: do nothing and return no error.
//  * If it is not a file, be nice: do nothing and return no error.
//  * If "Raw" is not "", be nice: the file is already loaded and
//    is quite possibly an on-the-fly temp file, so skip the load
//    and just do the quick MIME analysis.
func (pPI *PathInfo) FetchContent() *BasicContent {
	pBC := new(BasicContent)
	// pCC.BasicPath = pBP
	DispFP := Tilded(pPI.absFP.S())
	if !pPI.IsOkayFile() { // pBP.PathType() != "FILE" {
		pBC.bcError = errors.New("fu.FetchContent: not a readable file: " + DispFP)
		return pBC
	}
	var bb []byte
	bb = pPI.GetContentBytes()
	if pPI.bpError != nil {
		 pBC.bcError = fmt.Errorf("fu.FetchContent: BP.GetContentBytes<%s> failed: %w",
			DispFP, pPI.bpError)
		return pBC
	}
	pBC.Raw = S.TrimSpace(string(bb))
	if !S.HasPrefix(pPI.AbsFilePathParts.FileExt, ".") {
		println("==> (oops had to add a dot to filext")
		pPI.AbsFilePathParts.FileExt = "." + pPI.AbsFilePathParts.FileExt
	}
	if S.Contains(pBC.Raw, "<!DOCTYPE HTML ") {
		// println("FOUND HTML")
	}
	return pBC
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
	TheAbsFP := Tilded(pPI.absFP.S())
	if !pPI.IsOkayFile() {
		pPI.bpError = errors.New("fu.BP.GetContentBytes: not a file: " + TheAbsFP)
		return nil
	}
	if pPI.Size == 0 {
		println("==> zero-length file:", TheAbsFP)
		return make([]byte, 0)
	}
	// If it's too big, BARF!
	if pPI.Size > MAX_FILE_SIZE {
		 pPI.bpError = fmt.Errorf(
			"fu.BP.GetContentBytes: file too large (%d): %s", pPI.Size, TheAbsFP)
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

// FileType returns "XML", "MKDN", or future stuff TBD.
/*
func (p *CheckedContent) FileType() string {
	return p.BasicAnalysis.FileType()
}
*/
