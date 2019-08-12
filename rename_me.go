package fileutils

import (
	"fmt"
	"os/user"
	FP "path/filepath"
	S "strings"
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

/* =================

type InputFile struct {
	// RelFilePath is a "short" argument that can be passed in at
	// creation time. It may of course store an absolute (full) file
	// path instead. If this is "" then probably the next field is nil.
	RelFilePath
	// FileFullName stores the absolute filepath. It can be nil,
	// e.g. if the content is created on-the-fly.
	*AbsFilePathParts
	// FileInfo dusnt really have to be kept around. FIXME.
	os.FileInfo
	// Raw should be ignored after FileContent is non-empty.
	Raw string
	// FileContent holds the file's entire contents if (and only if)
	// Header is nil. After in-file metadata has been analysed and
	// stored into Header, Header is non-nil (altho possibly zero-value),
	// and FileContent contains only the non-metadata content.
	FileContent
	// Header basically holds all the in-file metadata, no matter what
	// the format of the file. For MDITA (Markdown-XP) this is YAML.
	// For XDITA and HDITA ([X]HTML) this is head/meta elements.
	// Storing metadata this way makes it easier to manage it in a
	// format-independent manner, and makes it easier to add to it
	// and modify it at runtime, and (TODO) when it is stored as
	// JSON K/V pairs, it can be accessed from the command line
	// using Sqlite tools.
	*Header
	// IsXML is set using various heuristics of our own devising.
	IsXML bool
	// MagicMimeType is set using a 3rd party binding to libmagic.
	MagicMimeType string
	// SniftMimeType is set using the Golang stdlib.
	SniftMimeType string
	// Mtype is set by our own code, based on MagicMimeType, SniftMimeType,
	// and analysis of the file contents possibly including recognising DTDs.
	Mtype []string
}

========================= */

// Header holds metadata. In default usage, this is metadata stored in the
// file, e.g. YAML in LwDITA Markdown-XP, or `head/meta` tags in [X]HTML.
// We store it here so that it is at the same level as the file "content".
// The file content is the entire file, until the metadata is split off.
// See `struct InputFile` for how to signal that the metadata has been
// split off.
type Header struct {
	HedRaw string
	Format string // "yaml", "dita", "html", etc.
	Props  map[string]string
}

// AbsFilePathParts holds the complete, fully-qualified absolute
// path and base name and file extension of a file or directory.
// If `DirPath` is "", the entire `FileFullName` is empty/invalid.
//
// Notes on usage:
// - `DirPath` must end with a slash `/`, or else we cannot
// distinguish empty/invalid from the filesystem root.
// - `Suffix` must start with a dot `.`, or else we cannot
// distinguish when a name ends with a dot.
// - If `FileNameParts` is a directory, the entire path is in
// `DirPath` (ending with "/"), and both `BaseName` and `Suffix` are "".
//
// Its Echo() method yields the full absolute path and name,
// so it is OK for production use, and it dusnt need to
// store the string-as-a-whole as another separate field.
type AbsFilePathParts struct {
	// DirPath holds the absolute path (from "filepath.Ext(path)"),
	// up to (and including) the last "/" directory separator.
	DirPath AbsFilePath
	// BaseName is the basic file name withOUT the extension.
	// It has no path (absolute OR relative), no "/" directory
	// separator, no final "." dot, no suffix.
	BaseName string
	// FileExt holds the file extension, and INCLUDES the
	// leading dot. This matches the result returned by
	// the stdlib API call "filepath.Ext(path)".
	FileExt string
}

// Echo yields the full absolute path and name, so it is OK for
// production use. If "DirPath" is "", the FileFullName is empty/invalid.
func (p AbsFilePathParts) Echo() string {
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
func (p AbsFilePathParts) String() string {
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
func (p CheckedPath) Echo() string {
	return p.AbsFilePathParts.Echo()
}

// String implements Markupper.
func (p CheckedPath) String() string {
	var s = "Dir"
	if !p.IsDir { // FileInfo.IsDir() {
		s = fmt.Sprintf("Len<%d>Mime<%s>", p.Size /*FileInfo.Size()*/, p.MagicMimeType)
	}
	return fmt.Sprintf("CheckedPath<%s=%s>:%s",
		p.RelFilePath, p.AbsFilePathParts.String(), s)
}

// GetAbsPathParts takes a filepath and uses `filepath.Abs(path)` to
// initialize the structure, but it does not check existence or file mode.
// `filepath.Abs(path)`` has the nice side effect that if `path` is a directory,
// then the directory path that is returned by `filepath.Split(abspath)` is
// guaranteed to end with a path separator slash `/`. <br/>
// Also, for consistency, the file extension is forced to all lower-case.
func (afp AbsFilePath) GetAbsPathParts() AbsFilePathParts {
	var filext, sdp string
	r := *new(AbsFilePathParts)
	if !FP.IsAbs(afp.S()) {
		return r
	}
	// abspath, _ = fp.Abs(string(relFP))
	sdp, r.BaseName = FP.Split(afp.S())
	if !S.HasSuffix(sdp, "/") {
		panic("fu.AbsFP.GetAbsPathParts.missingEndSlash<" + afp + ">")
	}
	r.DirPath = AbsFilePath(sdp)
	if r.BaseName == "" {
		return r // DIRECTORY !
	}
	filext = FP.Ext(r.BaseName)
	if filext != "" {
		r.BaseName = S.TrimSuffix(r.BaseName, filext)
		r.FileExt = S.ToLower(filext)
	}
	return r
}
