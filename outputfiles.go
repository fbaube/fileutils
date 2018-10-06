package fileutils

import (
	"io"
	"os"
	fp "path/filepath"
	S "strings"

	SU "github.com/fbaube/stringutils"
	"github.com/pkg/errors"
)

// LwdxFormats is a list of the types of text-based markup that
// are supported by LwDITA, and their corresponding file extensions.
// NOTE maybe this belongs in another package.
//
// Note that the "md" Markdown format for MDITA is a bit of a mishmash,
// though based mainly on CommonMark, with a few extensions, e.g. from
// GFM, and YAML for file header metadata extensions.
//
// Assume `html` is HTML5 *only*, so expect `<!DOCTYPE html>`.
var LwdxFormats = []string{
	"md", "xml", "xhtml", "html", "dita", "map", "ditamap", "bookmap"}

// ParserNames is a list of XML parsers we can use, and also file
// name modifiers for writing out parser-related temp & debug files.
var ParserNames = []string{"", "gohtml", "xmlx", "etree", "x2j", "mxj"}

// OutputFileExt is used for file input/output operations, when we
// understand that the file path and base name are stored elsewhere
// (i.e. in an InputFile), and that this struct specifies one file
// in a set.
type OutputFileExt struct {
	// Includes the period "."
	FileExt string
	io.WriteCloser
}

// Echo implements Markupper.
func (of OutputFiles) Echo() string {
	return "[OutputFiles]"
}

// String implements Markupper.
func (of OutputFiles) String() string {
	return "[OutputFiles]"
}

// OutputFiles is a list of all output files associated with the `InputFile`.
// Assume they all go to the same directory, but it does not have to be
// the same directory as the `InputFile`.
type OutputFiles struct {
	pInputFile *InputFile
	// OutDirPath is the full absolute directory path (but without file base
	// name or file extension). Normally it is the same as the input file's,
	// but it can also be a subdirectory whose name is based on the input file.
	// See func ../stringutils.DirNameFromFileName(..)
	OutputDirPath   AbsFilePath
	pOutputFileExts []*OutputFileExt
}

// NewOutputFiles creates the directory specified by adding `subdirSuffix`
// to the `InputFile`'s path + base name.
//
// It does not examine the file content, so it cannot decide not to create
// the directory for inappropriate file types, such as binary images.
//
// For convenience, if `subdirSuffix` is "", output files are placed in the
// same directory as the `InputFile`.
func (pIF *InputFile) NewOutputFiles(subdirSuffix string) (*OutputFiles, error) {

	p := new(OutputFiles)
	p.pInputFile = pIF

	if subdirSuffix == "" {
		p.OutputDirPath = pIF.DirPath
		return p, nil
	}
	// Transform the file name (an absolute filepath)
	// into a nearly-same directory name.
	var dn AbsFilePath
	sdn, ok := SU.DirNameFromFileName(pIF.String(), subdirSuffix)
	// !ok indicates a name pattern where no subdirectory is desired.
	if !ok {
		return nil, nil
	}
	dn = AbsFilePath(sdn)
	// Create (or open) the directory
	f, e := MustOpenOrCreateDir(dn)
	defer f.Close()
	if e != nil {
		return p, errors.Wrapf(e, "fu.NewOutputFiles.MustOpenOrCreateDir<%s>", dn)
	}
	p.OutputDirPath = dn
	return p, nil
}

// NewOutputExt opens a new empty file for writing.
// The `io.Writer` gets stored away at `pOFE.Writer`.
//
// Argument `filext`: aleading period `.` is optional.
func (pOF *OutputFiles) NewOutputExt(filext string) (*OutputFileExt, error) {
	var newpath string
	var f *os.File
	var pOFE *OutputFileExt
	var e error
	if filext == "" {
		return nil, errors.New("fu.NewOutputExt.emptyArg")
	}
	if !S.HasPrefix(filext, ".") {
		filext = "." + filext
	}
	newpath = fp.Join(string(pOF.OutputDirPath), pOF.pInputFile.BaseName+filext)
	f, e = os.OpenFile(newpath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	// Alternatively: f,e = MustCreate(newpath)
	if e != nil {
		return nil, errors.Wrapf(e, "fu.NewOutputExt<%s>", filext)
	}
	pOFE = new(OutputFileExt)
	pOFE.FileExt = filext
	pOFE.WriteCloser = f
	pOF.pOutputFileExts = append(pOF.pOutputFileExts, pOFE)
	return pOFE, nil
}
