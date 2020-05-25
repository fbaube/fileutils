package fileutils

import (
	"fmt"
	"os"
	"os/user"
	FP "path/filepath"
	S "strings"
	WU "github.com/fbaube/wasmutils"
)

// AbsFilePath is a new type, based on `string`. It serves three purposes:
// - clarify and bring correctness to the processing of absolute path arguments
// - permit the use of a clearly named struct field
// - permit the definition of methods on the type
//
// Note that when working with an `os.File`, `Name()` returns the name of the
// file as was passed to `Open(..)`, so it might be a relative filepath.
//
type AbsFilePath string

// Some prior overenthusiasm.
// type RelFilePath string
// type ArgFilePath string
// type FileContent string

// A token nod to Windoze compatibility.
const PathSep = string(os.PathSeparator)

// These should end in the path separator!
// NOTE See init(), at bottom.
var currentWorkingDir, currentUserHomeDir string

var currentUser *user.User

// GetHomeDir is a convenience function, and
// refers to the invoking user's home directory.
func GetHomeDir() string {
	return currentUserHomeDir
}

// S is a utility method to keep code cleaner.
func (afp AbsFilePath) S() string {
	s := string(afp)
	if !FP.IsAbs(s) {
		// panic("FU.types: AbsFP is not abs: " + s)
		// FIXME? // println("==> fu.types: AbsFP not abs: " + s)
		s, e := FP.Abs(s)
		if e != nil { panic("su.afp.S") }
		return s
	}
	return s
}

// AbsFP is like filepath.Abs(..) except using our own types.
func AbsFP(relFP string) AbsFilePath {
	if FP.IsAbs(relFP) {
		return AbsFilePath(relFP)
	}
	afp, e := FP.Abs(relFP)
	if e != nil {
		panic("fu.AbsFP<" + relFP + ">: " + e.Error())
	}
	return AbsFilePath(afp)
}

// Tilded shortens a filepath by substituting "." or "~".
func Tilded(s string) string {
	// If it's missing, use assumed/default...
	if s == "" {
		return "."
	}
	// If it's CWD...
	if s == currentWorkingDir {
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
	// At this point, if it's not an absolute FP, it's a relative FP,
	// but let it slide and don't prepend "./".
	if !FP.IsAbs(s) {
		// panic("NiceFP barfs on: " + s)
		return s
	}
	// println("arg:", s)
	// println("cwd:", currentworkingdir)

	if S.HasPrefix(s, currentWorkingDir) {
		return ("." + PathSep + S.TrimPrefix(s, currentWorkingDir))
		// bytesToTrim := len(currentWorkingDir) + 1
		// return "." + PathSep + s[bytesToTrim:]
	}
	if S.HasPrefix(s, currentUserHomeDir) {
		return ("~" + PathSep + S.TrimPrefix(s, currentUserHomeDir))
		// bytesToTrim := len(currentUserHomeDir) + 1
		// return "~" + PathSep + s[bytesToTrim:]
	}
	// No luck
	return s
}

// Append is a convenience function to keep code cleaner.
func (afp AbsFilePath) Append(rfp string) AbsFilePath {
	return AbsFilePath(FP.Join(afp.S(), rfp))
}

// StartsWith is like strings.HasPrefix(..) but uses our types.
func (afp AbsFilePath) HasPrefix(beg AbsFilePath) bool {
	return S.HasPrefix(afp.S(), beg.S())
}

func init() {
	var e error
	if WU.IsWasm() {
		currentUserHomeDir = "?"
		currentWorkingDir = "."
		currentUser = &user.User{
			Uid: "aUid",
    	Gid: "aGid",
    	Username: "webuser",
    	Name: "webuser",
    	HomeDir: "~",
		}
		return
	}
	currentUser, e = user.Current()
	if e != nil {
		println("==> ERROR: Cannot determine current user")
		return
	}
	currentUserHomeDir, e = os.UserHomeDir()
	if e != nil {
		println("==> ERROR: Cannot determine current user's home directory")
		return
	}
	homedir := currentUser.HomeDir
	if currentUserHomeDir != homedir {
		println("==> ERROR: Inconsistent values for current user's home directory")
		return
	}
	currentWorkingDir, e = os.Getwd()
	if e != nil {
		println("==> ERROR: Cannot determine current working directory")
		return
	}
	if !S.HasSuffix(currentWorkingDir, PathSep) {
		currentWorkingDir = currentWorkingDir + PathSep
	}
	if !S.HasSuffix(currentUserHomeDir, PathSep) {
		currentUserHomeDir = currentUserHomeDir + PathSep
	}
}

// SessionDemo can be called anytime.
func SessionDemo() {
	fmt.Fprintf(os.Stderr,
		"==> User_ID: %s (%s) (uid:%s,gid:%s) \n",
		 currentUser.Username, currentUser.Name, currentUser.Uid, currentUser.Gid)
	fmt.Fprintf(os.Stderr, "==> Working: %s \n", currentWorkingDir)
	/*
	if S.HasSuffix(currentUserHomeDir, "/") {
		println("--> Trimming trailing slash from UserHomeDir:", currentUserHomeDir)
		currentUserHomeDir = S.TrimSuffix(currentUserHomeDir, "/")
		println("--> UserHomeDir:", currentUserHomeDir)
	}
	*/
}
