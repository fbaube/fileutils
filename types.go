package fileutils

import (
	"os"
	"os/user"
	FP "path/filepath"
	S "strings"
)

// We define ´AbsFilePath`, a new type based on `string`.
// A FilePath type serves three purposes:
// - to clarify and bring correctness to the processing of path arguments
// - permit the use of   clearly named struct field that is a paths
// - permit the definition of methods on a type
// The uses is as follows:
// - `AbsFilePath` is for when we have resolved a filepath
//
// Note that when working with an os.File, `Name()` returns the name
// of the file as it was presented to `Open(..)`, so it might be a
// relative filepath.
//
type AbsFilePath string

// type RelFilePath string
// type ArgFilePath string
// type FileContent string

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
	if !FP.IsAbs(s) {
		panic("FU.types: AbsFP is Rel: " + s)
	}
	return s
}

// AbsFP is like filepath.Abs(..) except using our own types.
func AbsFP(rfp string) AbsFilePath {
	if FP.IsAbs(rfp) {
		return AbsFilePath(rfp)
	}
	afp, e := FP.Abs(rfp)
	if e != nil {
		panic("fu.AbsFP<" + rfp + ">: " + e.Error())
	}
	return AbsFilePath(afp)
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
	// println("arg:", s)
	// println("cwd:", currentworkingdir)
	if s == currentworkingdir {
		return "."
	}
	if S.HasPrefix(s, currentworkingdir) {
		bytesToTrim := len(currentworkingdir) + 1
		return "." + PathSep + s[bytesToTrim:]
	}
	if S.HasPrefix(s, homedir) {
		bytesToTrim := len(homedir) + 1
		return "~" + PathSep + s[bytesToTrim:]
	}
	return s
}

// Append is a convenience function to keep code cleaner.
func (afp AbsFilePath) Append(rfp string) AbsFilePath {
	// return AbsFilePath(afp.S() + rfp)
	return AbsFilePath(FP.Join(afp.S(), rfp))
}

// StartsWith is like strings.HasPrefix(..) but uses our types.
func (afp AbsFilePath) StartsWith(beg AbsFilePath) bool {
	return S.HasPrefix(afp.S(), beg.S())
}

func init() {
	username, e := user.Current()
	if e != nil {
		println("==> ERROR: Cannot determine current user")
		return
	}
	homedir = username.HomeDir
	// println("HOME:", homedir)

	currentworkingdir, e = os.Getwd()
	if e != nil {
		println("==> ERROR: Cannot determine current working directory")
		return
	}
	// println(" CWD:", currentworkingdir)
}
