package fileutils

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	fp "path/filepath"
	S "strings"

	ft "github.com/h2non/filetype"
	"github.com/pkg/errors"
	// "github.com/dimchansky/utfbom"
)

// InputFile describes in detail a file we have redd or will read.
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
type InputFile struct {
	// RelFilePath is a "short" argument that can be passed in at
	// creation time. It may of course store an absolute (full) file
	// path instead. If this is "" then probably the next field is nil.
	RelFilePath
	// FileFullName can be nil, e.g. if the
	// content is created on-the-fly.
	*FileFullName
	// FileInfo dusnt really have to be kept around. FIXME.
	os.FileInfo
	// MimeType is the type returned by a third-party Mime-type library,
	// with some possible modifications (e.g. recognising DTDs). Deeper
	// analysis of the file's contents occurs elsewhere.
	FileContent
	MimeType string
	IsXML    bool
	MMCtype  []string
	Mtype    []string
}

// FileFullName holds the complete, fully-qualified absolute
// path and base name and file extension of a file or directory.
// If `DirPath` is "", the entire `FileFullName` is empty/invalid.
//
// Notes on usage:
// - `DirPath` must end with a slash `/`, or else we cannot
// distinguish empty/invalid from the filesystem root.
// - `Suffix` must start with a dot `.`, or else we cannot
// distinguish when a name ends with a dot.
// - If `FileFullName` is a directory, the entire path is in
// `DirPath` (ending with "/"), and `BaseName` and `Suffix` are "".
//
// Its Echo() method yields the full absolute path and name,
// so it is OK for production use, and it dusnt need to
// store the string-as-a-whole as another separate field.
type FileFullName struct {
	// DirPath holds the absolute path (from "filepath.Ext(path)"),
	// up to (and including) the last "/" directory separator.
	DirPath AbsFilePath
	// BaseName has no path (absolute OR relative), no "/"
	// directory separator, no final "." dot, no suffix.
	BaseName string
	// FileExt holds the file extension, and includes the
	// leading dot; this matches "filepath.Ext(path)".
	FileExt string
}

// Echo yields the full absolute path and name, so it is OK for
// production use. If "DirPath" is "", the FileFullName is empty/invalid.
func (p FileFullName) Echo() string {
	dp := string(p.DirPath)
	fx := p.FileExt
	if dp == "" {
		return ""
	}
	if !S.HasPrefix(dp, "/") || !S.HasSuffix(dp, "/") {
		panic("fu.FileFullName.missingDirSep<" + dp + ">")
	}
	if fx != "" && !S.HasPrefix(fx, ".") {
		panic("fu.FileFullName.missingFileExtDot<" + fx + ">")
	}
	return dp + p.BaseName + fx
}

// String implements Markupper.
func (p FileFullName) String() string {
	// s := p.String()
	s := p.DirPath.S() + p.BaseName + p.FileExt
	username, e := user.Current()
	if e != nil {
		return s
	}
	homedir := username.HomeDir
	if !S.HasPrefix(s, homedir) {
		return s
	}
	L := len(homedir)
	s = "~" + s[L:]
	return s
}

// Echo implements Markupper.
func (p InputFile) Echo() string {
	return p.FileFullName.Echo()
}

// String implements Markupper.
func (p InputFile) String() string {
	var s = "Dir"
	if !p.FileInfo.IsDir() {
		s = fmt.Sprintf("Len<%d>Mime<%s>", p.FileInfo.Size(), p.MimeType)
	}
	return fmt.Sprintf("InputFile<%s=%s>:%s",
		p.RelFilePath, p.FileFullName.String(), s)
}

// NewFileFullName accepts a relative filepath and uses `filepath.Abs(path)`
// to initialize the structure, but it does not check existence or file mode.
// `filepath.Abs(path)`` has the nice side effect that if `path` is a directory,
// then the directory path that is returned by `filepath.Split(abspath)` is
// guaranteed to end with a path separator slash `/`. <br/>
// Also, for consistency, the file extension is forced to all lower-case.
func NewFileFullName(path RelFilePath) *FileFullName {
	if path == "" {
		return nil // BAD ARGUMENT !
	}
	var abspath, filext, sdp string
	p := new(FileFullName)
	abspath, _ = fp.Abs(string(path))
	sdp, p.BaseName = fp.Split(abspath)
	if !S.HasSuffix(sdp, "/") {
		panic("fu.NewFileFullName.DirPath.missingEndSlash<" + p.DirPath + ">")
	}
	p.DirPath = AbsFilePath(sdp)
	if p.BaseName == "" {
		return p // DIRECTORY !
	}
	filext = fp.Ext(p.BaseName)
	if filext != "" {
		p.BaseName = S.TrimSuffix(p.BaseName, filext)
		p.FileExt = S.ToLower(filext)
	}
	return p
}

// NewInputFile reads in the file's content. So, it can return
// an error. It is currently limited to about 2 megabytes.
func NewInputFile(path RelFilePath) (*InputFile, error) {
	var f *os.File
	var bb []byte
	var fullpath string
	var e error

	// Check that the file exists and is readable
	f, e = os.Open(string(path))
	defer f.Close()
	if e != nil {
		return nil, errors.Wrapf(e, "fu.NewInputFile.osOpen<%s>", path)
	}
	// We're good to go
	p := new(InputFile)
	p.FileFullName = NewFileFullName(path)
	fullpath = p.FileFullName.Echo()
	p.FileInfo, e = os.Stat(fullpath)
	if e != nil {
		return nil, errors.Wrapf(e, "fu.NewInputFile.osStat<%s>", fullpath)
	}
	// If it's too big, BARF!
	if p.Size() > 2000000 {
		return p, errors.New(fmt.Sprintf("fu.NewInputFile<%s>: file too large: %d",
			fullpath, p.Size()))
	}
	// Read it in !
	// TODO/FIXME Use github.com/dimchansky/utfbom Skip()
	bb, e = ioutil.ReadFile(fullpath)
	if e != nil {
		return nil, errors.Wrapf(e, "fu.NewInputFile.ioutilReadFile<%s>", fullpath)
	}

	// OBSOLETE!
	// p.MimeType, _ = MimeBuffer(bb, int(HgMimeType))
	tt, e := ft.Match(bb)
	// type Type struct {
	//      MIME      MIME
	//      Extension string
	println("EXT", tt.Extension, "MIME", tt.MIME.Type, tt.MIME.Subtype, tt.MIME.Value)
	p.MimeType = tt.MIME.Type + "/" + tt.MIME.Subtype + "/" + tt.MIME.Value

	// Trim away whitespace! We do this so that other code can
	// check for known patterns at the "start" of the file.
	p.FileContent = FileContent(S.TrimSpace(string(bb)))
	// println("(DD:fu.InF) MIME as analyzed:", p.FileFullName.FileExt, p.MimeType)
	return p, nil
}

// ================================================================

// CreateFromFilePath promotes the filepath string to an InputFile.
func (p *InputFile) CreateFromFilePath(inputFile string) *InputFile {
	p.RelFilePath = RelFilePath(inputFile)
	return p
}

func (p *InputFile) ParseFileName() *InputFile {
	if p.RelFilePath == "" {
		return p // BAD ARGUMENT! But not fatal.
	}
	var abspath, filext, dirpath string
	p.FileFullName = new(FileFullName)
	abspath, _ = fp.Abs(string(p.RelFilePath))
	dirpath, p.BaseName = fp.Split(abspath)
	// FIXME Use os.PathSeparator
	if !S.HasSuffix(dirpath, "/") {
		panic("fu.ParseFileName.DirPath.missingEndSlash<" + dirpath + ">")
	}
	p.DirPath = AbsFilePath(dirpath)
	if p.BaseName == "" {
		return p // DIRECTORY! BaseName and FileExt are left to be "".
	}
	filext = fp.Ext(p.BaseName)
	if filext != "" {
		p.BaseName = S.TrimSuffix(p.BaseName, filext)
		p.FileExt = S.ToLower(filext)
	}
	return p
}

// OpenAndLoadContent does just that. It also sets
// MimeType, using some really simple code.
func (p *InputFile) OpenAndLoadContent() (*InputFile, error) {
	var e error
	fullpath := p.FileFullName.Echo()
	p.FileInfo, e = os.Stat(fullpath)
	if e != nil {
		return nil, errors.Wrapf(e, "fu.OpenAndLoadContent.osStat<%s>", fullpath)
	}
	// If it's too big, BARF!
	if p.Size() > 2000000 {
		return p, errors.New(fmt.Sprintf(
			"fu.OpenAndLoadContent<%s>: file too large: %d",
			fullpath, p.Size()))
	}
	// Open it (and then immediately close it), just to check.
	var pF *os.File
	pF, e = os.Open(fullpath)
	pF.Close()
	if e != nil {
		return nil, errors.Wrapf(e, "fu.OpenAndLoadContent.osOpen<%s>", fullpath)
	}
	// Read it in !
	// TODO/FIXME Use github.com/dimchansky/utfbom Skip()
	var bb []byte
	bb, e = ioutil.ReadFile(fullpath)
	if e != nil {
		return nil, errors.Wrapf(e, "fu.OpenAndLoadContent.ioutilReadFile<%s>", fullpath)
	}
	p.MimeType, _ = MimeBuffer(bb, int(HgMimeType))
	// Trim away whitespace! We do this so that other code can
	// check for known patterns at the "start" of the file.
	p.FileContent = FileContent(S.TrimSpace(string(bb)))
	// println("(DD:fu.InF) MIME as analyzed:", p.FileFullName.FileExt, p.MimeType)
	return p, nil
}

// NewInputFileFromStdin reads `os.Stdin` completely and returns a new
// `InputFile`.
func NewInputFileFromStdin() (*InputFile, error) {

	p := new(InputFile)
	p.RelFilePath = "-"
	// p.FileFullName is left at "nil"

	bb, e := ioutil.ReadAll(os.Stdin)
	if e != nil {
		return nil, errors.Wrap(e, "Can't read standard input")
	}
	p.FileContent = FileContent(S.TrimSpace(string(bb)))
	p.MimeType = "text/plain"
	return p, nil
}
