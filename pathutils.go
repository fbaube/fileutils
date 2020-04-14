package fileutils

import (
	"fmt"
	"os/user"
	FP "path/filepath"
	S "strings"
	// "github.com/dimchansky/utfbom"
)

// AbsFilePathParts holds the complete, fully-qualified absolute
// path and base name and file extension of a file or directory.
// If "DirPath" is "", the entire "FileFullName" is empty/invalid.
//
// Notes on usage:
//  * "DirPath" must end with a slash "/", or else we cannot
//     distinguish an empty/invalid path from the filesystem root.
//  * "Suffix" must start with a dot ".", or else we cannot
//     distinguish the edge case where a name ends with a dot.
//  * If "AbsFilePathParts" describes a directory, the entire
//    path is in "DirPath" (ending with "/"), and both "BaseName"
//    and "Suffix" are "".
//
// Its Echo() method yields the full absolute path and name,
// OK for production use, so this struct dusnt need to store
// the string-as-a-whole as another separate field.
//
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
func (p BasicPath) Echo() string {
	return p.AbsFilePathParts.Echo()
}

// String implements Markupper.
func (p CheckedContent) String() string {
	if p.IsOkayDir() {
		return fmt.Sprintf("BscPth: DIR[%d]: %s | %s",
			p.Size, p.RelFilePath, p.AbsFilePathParts.Echo())
	}
	var isXML string
	if p.IsXML {
		isXML = "[XML] "
	}
	s := fmt.Sprintf("ChP: %sLen<%d>Mtype<%s>",
		isXML, p.Size, p.Mstring())
	s += fmt.Sprintf("\n\t %s | %s",
		p.RelFilePath, Tilded(p.AbsFilePathParts.Echo()))
	s += fmt.Sprintf("\n\t (snift) %s | (magic) %s",
		p.SniftMimeType, p.MagicMimeType)
	return s
}

// GetAbsPathParts takes an absolute filepath and uses "filepath.Abs(path)""
// to initialize the structure, but it does not check existence or file mode.
// "filepath.Abs(path)" has a nice side effect that if "path" is a directory,
// then the directory path that is returned by "filepath.Split(abspath)" is
// guaranteed to end with a path separator slash `/`.
//
// Also, for consistency, the file extension is forced to all lower-case.
func (afp AbsFilePath) NewAbsPathParts() *AbsFilePathParts {
	var filext, sdp string
	r := new(AbsFilePathParts)
	if !FP.IsAbs(afp.S()) {
		return r
	}
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
