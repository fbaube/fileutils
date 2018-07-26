package fileutils

import (
	"os"
	"os/user"
	fp "path/filepath"
	S "strings"
)

// AbsFilePath, RelFilePath, and FileContent are declared as new types.
// Why? Bevause in our code, we want in general to use (and require)
// absolute filepaths; therefore we clearly mark relative/short
// filepaths (as used in the CLI)
//
// Note that when working with an os.File, `Name()` returns the name
// of the file as it was presented to `Open(..)`, so it might be a
// relative filepath.
//
type AbsFilePath string
type RelFilePath string
type FileContent string

// A token nod to Windoze compatibility.
const PathSep = string(os.PathSeparator)

var homedir AbsFilePath

// GetHomeDir is a convenience function, and
// refers to the invoking user's home directory.
func GetHomeDir() string {
	return string(homedir)
}

// S is a utility method to keep code cleaner.
func (afp AbsFilePath) S() string {
	s := string(afp)
	if !S.HasPrefix(s, "/") {
		panic("FU.types: AbsFP is Rel: " + s)
	}
	return string(s)
}

// S is a utility method to keep code cleaner.
func (rfp RelFilePath) S() string {
	return string(rfp)
}

// ResolveToAbsoluteFP is the kind of function we use to keep code
// un-confused, and it is why we defined these types for filepaths.
func ResolveToAbsoluteFP(s string) AbsFilePath {
	if S.HasPrefix(s, PathSep) {
		return AbsFilePath(s)
	}
	return RelFilePath(s).ResolveToAbsolute()
}

// RelFP is a totally kosher downcast and is to keep code cleaner.
func (apr AbsFilePath) RelFP() RelFilePath {
	return RelFilePath(apr)
}

// ResolveToAbsolute relies on `filepath.Abs(path)`.
func (rpf RelFilePath) ResolveToAbsolute() AbsFilePath {
	if S.HasPrefix(string(rpf), PathSep) {
		return AbsFilePath(rpf)
	}
	abspath, e := fp.Abs(string(rpf))
	if e != nil {
		panic("ResolveRelToAbs: " + rpf)
	}
	return AbsFilePath(abspath)
}

// MakeRelativeWRT tried to convert an abs path to a rel path,
// based on the abs filepath passed in.
func (afp AbsFilePath) MakeRelativeWRT(wrt AbsFilePath) RelFilePath {
	if !afp.StartsWith(homedir) {
		return RelFilePath(afp)
	}
	bytesToTrim := len(wrt) + 1
	return RelFilePath("~" + PathSep + string(afp)[bytesToTrim:])
}

// ElideUserHome replaces the user's home dir with tilde `~` if possible.
func (afp AbsFilePath) ElideUserHome() RelFilePath {
	// IFF barfs :: return RelFilePath(afp)
	return afp.MakeRelativeWRT(homedir)
}

// AbsFilePath is a convenience function to keep code cleaner.
func (afp AbsFilePath) Append(rfp RelFilePath) AbsFilePath {
	return AbsFilePath(string(afp) + string(rfp))
}

// StartsWith is a convenience function to keep code cleaner.
func (afp AbsFilePath) StartsWith(beg AbsFilePath) bool {
	return S.HasPrefix(string(afp), string(beg))
}

func init() {
	username, e := user.Current()
	if e != nil {
		println("==> ERROR: Could not determine current user")
		return
	}
	homedir = AbsFilePath(username.HomeDir)
}
