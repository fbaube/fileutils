package fileutils

import (
	"fmt"
	"io/ioutil"
	"os"
	fp "path/filepath"
	S "strings"

	SU "github.com/fbaube/stringutils"
	"github.com/pkg/errors"
)

// FileFullName holds the complete, fully-qualified
// absolute path and name of a file or directory.
// If DirPath is "", the FileFullName is empty/invalid.
//
// DirPath must end with a slash "/", or else we cannot
// distinguish empty/invalid from the filesystem root.
// Suffix must start with a dot ".", or else we cannot
// distinguish when a name ends with a dot.
// If FileFullName is a directory, the entire path is in
// DirPath (ending with "/"), and BaseName & Suffix are "".
//
// Its String() yields the full absolute path and name,
// so it is OK for production use, and it dusnt need
// to actually store the string-as-a-whole anywhere.
type FileFullName struct {
	// DirPath holds the absolute path (from "filepath.Ext(path)"),
	// up to (and including) the last "/" directory separator.
	DirPath string
	// BaseName has no path (absolute OR relative), no "/"
	// directory separator, no final "." dot, no suffix.
	BaseName string
	// FileExt holds the file extension, and includes the
	// leading dot; this matches "filepath.Ext(path)".
	FileExt string
}

// String yields the full absolute path and name, so it is OK for
// production use. If "DirPath" is "", the FileFullName is empty/invalid.
func (p *FileFullName) String() string {
	dp := p.DirPath
	fx := p.FileExt
	if dp == "" {
		return ""
	}
	if !S.HasPrefix(dp, "/") || !S.HasSuffix(dp, "/") {
		panic("fileutils.FileFullName.missingDirSep<" + dp + ">")
	}
	if fx != "" && !S.HasPrefix(fx, ".") {
		panic("fileutils.FileFullName.missingFileExtDot<" + fx + ">")
	}
	return dp + p.BaseName + fx
}

// InputFile describes in detail a file we have redd or will read.
// It does not though include any XML-specific stuff; it is generic.
// In normal usage, when the InputFile is created, the file is opened
// and its contents are redd into "Contents".
// NOTE An SVG or EPS file (for example) can be IsImage but !IsBinary.
// FIXME It is not currently implemented.
type InputFile struct {
	path string // argument at creation time
	FileFullName
	os.FileInfo
	MimeType string
	IsImage  bool
	IsBinary bool
	Contents string
}

// DString is for debug output.
func (p *InputFile) DString() string {
	return fmt.Sprintf(
		"InputFile<%s>sz<%d>dir?<%s>bin?<%s>img?<%s>mime<%s>",
		p.FileFullName.String(), p.FileInfo.Size(),
		SU.Yn(p.FileInfo.IsDir()), SU.Yn(p.IsBinary),
		SU.Yn(p.IsImage), p.MimeType)
}

// NewFileFullName uses fp.Abs(path) to initialize the structure,
// but it does not check existence or file mode.
// fp.Abs(path) has the nice side effect that if path is a directory,
// then the directory path that is returned by fp.Split(abspath) is
// guaranteed to end with a slash "/". Also, for convenience, the
// file extension is forced to all lower-case.
func NewFileFullName(path string) *FileFullName {
	if path == "" {
		return nil // BAD ARGUMENT !
	}
	var abspath, filext string
	p := new(FileFullName)
	abspath, _ = fp.Abs(path)
	p.DirPath, p.BaseName = fp.Split(abspath)
	if !S.HasSuffix(p.DirPath, "/") {
		panic("fileutils.NewFileFullName.DirPath.missingEndSlash")
	}
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
func NewInputFile(path string) (*InputFile, error) {
	var f *os.File
	var bb []byte
	var fullpath string
	var e error
	// Check that the file exists and is readable
	f, e = os.Open(path)
	defer f.Close()
	if e != nil {
		return nil, errors.Wrap(e, "fileutils.NewInputFile.osOpen<"+path+">")
	}
	// We're good to go
	p := new(InputFile)
	p.FileFullName = *NewFileFullName(path)
	fullpath = p.FileFullName.String()
	p.FileInfo, e = os.Stat(fullpath)
	if e != nil {
		return nil, errors.Wrap(e, "fileutils.NewInputFile.osStat<"+fullpath+">")
	}
	// If it's too big, BARF!
	if p.Size() > 2000000 {
		return p, fmt.Errorf("fileutils.NewInputFile<%s>: file too large: %d",
			fullpath, p.Size())
	}
	// Read it in !
	bb, e = ioutil.ReadFile(fullpath)
	if e != nil {
		return nil, errors.Wrap(e, "fileutils.NewInputFile.ioutilReadFile<"+fullpath+">")
	}
	p.MimeType, _ = MimeBuffer(bb, int(MimeType))
	if S.HasPrefix(p.MimeType, "image/") {
		p.IsImage = true
		// FIXME Not true for SVG, EPS
		p.IsBinary = true
	}
	p.Contents = string(bb)
	return p, nil
}

var gotOkayExts bool
var theOkayExts []string
var theOkayFiles []string

// GatherInputFiles handles the case where the path is a directory
// (altho it can also handle a simple file argument).
// It always excludes dotfiles (filename begins with ".") and emacs
// backups (filename ends with "~").
// It includes only files that end with any extension in the slice
// "okayExts" (the check is case-insensitive). Each extension in the
// slice argument should include the period; the function will get
// additional functionality if & when the periods are not included.
// If "okayExts" is nil, *all* file extensions are included.
func GatherInputFiles(path string, okayExts []string) (okayFiles []string, err error) {

	theOkayExts = okayExts
	gotOkayExts = (okayExts != nil && len(okayExts) > 0)

	// A single file ?
	if !IsDirectory(path) {
		sfx := fp.Ext(path)
		_, found := SU.InSliceIgnoreCase(sfx, okayExts)
		if found || !gotOkayExts {
			abs, _ := fp.Abs(path)
			theOkayFiles = append(theOkayFiles, abs)
		}
		return theOkayFiles, nil
	}
	// PROCESS THE DIRECTORY
	err = fp.Walk(path, myWalkFunc)
	if err != nil {
		return nil, errors.Wrap(err, "fileutils.GatherInputFiles.walkTo<"+path+">")
	}
	return theOkayFiles, nil
}

func myWalkFunc(path string, finfo os.FileInfo, inerr error) error {
	var abspath string
	var f *os.File
	var e error
	// print("path|" + path + "|finfo|" + finfo.Name() + "|\n")
	if !S.HasSuffix(path, finfo.Name()) {
		panic("myWalkFunc")
	}
	if inerr != nil {
		return errors.Wrap(inerr, "fileutils.myWalkFunc<"+path+">")
	}
	// Is it hidden, or an emacs backup ? If so, ignore.
	if S.HasPrefix(path, ".") || S.HasSuffix(path, "~") {
		if finfo.IsDir() {
			return fp.SkipDir
		} else {
			return nil
		}
	}
	// Is it a (non-hidden) directory ? If so, carry on.
	if finfo.IsDir() {
		return nil
	}
	if _, found := SU.InSliceIgnoreCase(
		fp.Ext(path), theOkayExts); gotOkayExts && !found {
		return nil
	}
	abspath, e = fp.Abs(path)
	if e != nil {
		return errors.Wrap(e, "fileutils.myWalkFunc<"+path+">")
	}
	f, e = MustOpenRO(abspath)
	defer f.Close()
	if e != nil {
		return errors.Wrap(e, "fileutils.myWalkFunc.MustOpenRO<"+path+">")
	}
	// fmt.Printf("(DD) Infile OK: %+v \n", abspath)
	theOkayFiles = append(theOkayFiles, abspath)
	return nil
}
