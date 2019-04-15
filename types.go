package fileutils

import (
	"os"
	"os/user"
	fp "path/filepath"
	S "strings"
)

// We define three new types based on `string`:
// - AbsFilePath
// - RelFilePath
// - FileContent
// The FilePath types serve three purposes:
// - to clarify and bring correctness to the processing of path arguments
// - permit the use of clearly named struct fields that are all paths
// - permit the definition of methods on each type
// Their uses are as follows:
// - `AbsFilePath` is for when we have resolved a filepath
// - `RelFilePath` is for a CLI argument, where absolute/relative is TBD,
// and maybe also for other purposes as yet unknown
//
// Note that when working with an os.File, `Name()` returns the name
// of the file as it was presented to `Open(..)`, so it might be a
// relative filepath.
//
type ArgFilePath string
type AbsFilePath string
type RelFilePath string
type FileContent string

// A token nod to Windoze compatibility.
const PathSep = string(os.PathSeparator)

// NOTE See init(), at bottom
var homedir, currentworkingdir string

// GetHomeDir is a convenience function, and
// refers to the invoking user's home directory.
func GetHomeDir() string {
	return string(homedir)
}

// S is a utility method to keep code cleaner.
func (afp AbsFilePath) S() string {
	s := string(afp)
	if !fp.IsAbs(s) {
		panic("FU.types: AbsFP is Rel: " + s)
	}
	return s
}

// S is a utility method to keep code cleaner.
func (rfp RelFilePath) S() string {
	return string(rfp)
}

// AbsFP is like filepath.Abs(..) except using our own types.
func (rfp RelFilePath) AbsFP() AbsFilePath {
	s := rfp.S()
	if fp.IsAbs(s) {
		return AbsFilePath(s)
	}
	afp, e := fp.Abs(s)
	if e != nil {
		panic("fu.AbsFP<" + s + ">: " + e.Error())
	}
	return AbsFilePath(afp)
}

// RelFP is a totally kosher downcast and is to keep code cleaner.
func (afp AbsFilePath) RelFP() RelFilePath {
	return RelFilePath(afp)
}

// NiceFP shortens a filepath by substituting "." or "~".
func NiceFP(s string) string {
	// if it's missing, and has an assumed/default...
	if s == "" {
		return "."
	}
	// If it can't be normalised...
	if s == "." || s == "~" || s == PathSep {
		return s
	}
	// If it can't be further normalised...
	if S.HasPrefix(s, "."+PathSep) || S.HasPrefix(s, "~"+PathSep) {
		return s
	}
	// At this point, if it's not an absolute FP, it's a problem.
	if !S.HasPrefix(s, PathSep) {
		panic("NiceFP barfs on: " + s)
	}
	if S.HasPrefix(s, currentworkingdir) {
		bytesToTrim := len(currentworkingdir) + 1
		return "." + PathSep + s[bytesToTrim:]
	}
	if S.HasPrefix(s, homedir) {
		bytesToTrim := len(homedir) + 1
		return "." + PathSep + s[bytesToTrim:]
	}
	return s
}

/*
// ElideUserHome converts an abs path under homedir to a path
// that uses "~" (but is still an abs file path!)."
func (afp AbsFilePath) ElideUserHome() AbsFilePath {
	s := afp.S()
	if !fp.IsAbs(s) {
		panic("fu.elideUserHome: not absolute FP: " + afp)
	}
	if !afp.StartsWith(AbsFilePath(homedir)) {
		return afp
	}
	bytesToTrim := len(homedir) + 1
	return AbsFilePath("~" + PathSep + s[bytesToTrim:])
}

// ElideUserHome converts an abs path under homedir to a path
// that uses "~" (but is still an abs file path!)."
func (afp AbsFilePath) ElideCWD() AbsFilePath {
	s := afp.S()
	if !fp.IsAbs(s) {
		panic("fu.elideCWD: not absolute FP: " + afp)
	}
	if !afp.StartsWith(AbsFilePath(currentworkingdir)) {
		return afp
	}
	bytesToTrim := len(currentworkingdir) + 1
	return AbsFilePath("." + PathSep + s[bytesToTrim:])
}
*/

// AbsFilePath is a convenience function to keep code cleaner.
func (afp AbsFilePath) Append(rfp RelFilePath) AbsFilePath {
	return AbsFilePath(string(afp) + string(rfp))
}

// StartsWith is like strings.HasPrefix(..) but uses our types.
func (afp AbsFilePath) StartsWith(beg AbsFilePath) bool {
	return S.HasPrefix(string(afp), string(beg))
}

func init() {
	username, e := user.Current()
	if e != nil {
		println("==> ERROR: Cannot determine current user")
		return
	}
	homedir = username.HomeDir
	println("HOME:", homedir)

	currentworkingdir, e = os.Getwd()
	if e != nil {
		println("==> ERROR: Cannot determine current working directory")
		return
	}
	println(" CWD:", currentworkingdir)
}
